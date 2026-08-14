package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"

	apperrors "solitude-blog/server/internal/errors"
	"solitude-blog/server/internal/model"
	"solitude-blog/server/internal/pagination"
	"solitude-blog/server/internal/storage"
)

const maxAssetUploadSize int64 = 10 * 1024 * 1024

type AssetService struct {
	db    *gorm.DB
	store storage.ObjectStorage
	mu    sync.RWMutex
	items []AssetItem
}

type AssetListQuery struct {
	Page     int
	PageSize int
	Keyword  string
	Mime     string
}

type AssetItem struct {
	ID          uint64    `json:"id"`
	DisplayName string    `json:"display_name"`
	AltText     string    `json:"alt_text"`
	StorageKey  string    `json:"storage_key"`
	URL         string    `json:"url"`
	ThumbURL    string    `json:"thumb_url"`
	MimeType    string    `json:"mime_type"`
	Ext         string    `json:"ext"`
	Size        uint64    `json:"size"`
	Width       uint      `json:"width"`
	Height      uint      `json:"height"`
	SHA256      string    `json:"sha256"`
	Status      string    `json:"status"`
	RefCount    uint      `json:"ref_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type AssetUpdateRequest struct {
	DisplayName string `json:"display_name"`
	AltText     string `json:"alt_text"`
}

type AssetReferenceItem struct {
	ID    uint64 `json:"id"`
	Type  string `json:"type"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

func NewAssetService(db *gorm.DB, store storage.ObjectStorage) *AssetService {
	return &AssetService{db: db, store: store}
}

func (s *AssetService) List(query AssetListQuery) ([]AssetItem, pagination.ListPage, error) {
	page := pagination.NormalizePage(query.Page)
	pageSize := pagination.NormalizePageSize(query.PageSize)
	if s.db != nil {
		return s.listFromDB(query, page, pageSize)
	}
	return s.listInMemory(query, page, pageSize)
}

func (s *AssetService) Upload(fileHeader *multipart.FileHeader, displayName string) (AssetItem, error) {
	if fileHeader == nil {
		return AssetItem{}, apperrors.New(apperrors.CodeMissingRequiredField)
	}
	if fileHeader.Size > maxAssetUploadSize {
		return AssetItem{}, apperrors.New(apperrors.CodeUploadFileTooLarge)
	}

	source, err := fileHeader.Open()
	if err != nil {
		return AssetItem{}, apperrors.New(apperrors.CodeInvalidRequest)
	}
	defer source.Close()

	// 先读前 512 字节嗅探 MIME；MIME 白名单不通过直接拒绝。
	// 由于 multipart.File 是顺序 reader，嗅探消耗的字节必须先写入 buffer 与 hasher，
	// 否则 io.Copy 后续会从游标处继续读，导致落盘文件丢失前 512 字节（PNG/JPG magic header 全丢）。
	head := make([]byte, 512)
	n, err := io.ReadFull(source, head)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return AssetItem{}, apperrors.New(apperrors.CodeInvalidRequest)
	}
	head = head[:n]
	mimeType := http.DetectContentType(head)
	ext, allowed := assetExtensionForMIME(mimeType)
	if !allowed {
		return AssetItem{}, apperrors.New(apperrors.CodeUnsupportedFileType)
	}

	// 单次 io.Copy 同时写哈希与内存缓冲区，避免两次读取；
	// 受 maxAssetUploadSize 限制，整文件驻留内存可接受。
	hasher := sha256.New()
	hasher.Write(head)
	var buffer bytes.Buffer
	buffer.Write(head)
	if _, err := io.Copy(io.MultiWriter(hasher, &buffer), source); err != nil {
		return AssetItem{}, apperrors.New(apperrors.CodeStorageUnavailable)
	}
	data := buffer.Bytes()
	hashSum := hasher.Sum(nil)

	width, height := decodeImageDimensions(data, mimeType)

	now := time.Now().UTC()
	storageKey := filepath.ToSlash(filepath.Join(
		now.Format("2006"),
		now.Format("01"),
		randomAssetName()+ext,
	))

	url, err := s.store.Put(context.Background(), storageKey, bytes.NewReader(data), int64(len(data)), mimeType)
	if err != nil {
		return AssetItem{}, apperrors.New(apperrors.CodeStorageUnavailable)
	}

	if displayName == "" {
		displayName = fileHeader.Filename
	}
	item := AssetItem{
		DisplayName: displayName,
		StorageKey:  storageKey,
		URL:         url,
		ThumbURL:    "",
		MimeType:    mimeType,
		Ext:         strings.TrimPrefix(ext, "."),
		Size:        uint64(len(data)),
		Width:       width,
		Height:      height,
		SHA256:      hex.EncodeToString(hashSum),
		Status:      "ready",
		RefCount:    0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if s.db != nil {
		row := assetModelFromItem(item)
		if err := s.db.Create(&row).Error; err != nil {
			// 数据库写入失败时清理已上传的对象，避免孤儿文件。
			_ = s.store.Delete(context.Background(), storageKey)
			return AssetItem{}, apperrors.New(apperrors.CodeDatabaseUnavailable)
		}
		return assetFromModel(row), nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	item.ID = uint64(len(s.items) + 1)
	s.items = append(s.items, item)
	return item, nil
}

func (s *AssetService) Update(id string, req AssetUpdateRequest) (AssetItem, error) {
	parsed, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return AssetItem{}, apperrors.New(apperrors.CodeInvalidParameter)
	}
	if req.DisplayName == "" {
		return AssetItem{}, apperrors.New(apperrors.CodeMissingRequiredField)
	}

	if s.db != nil {
		var row model.Asset
		err := s.db.First(&row, parsed).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return AssetItem{}, apperrors.New(apperrors.CodeResourceNotFound)
		}
		if err != nil {
			return AssetItem{}, apperrors.New(apperrors.CodeDatabaseUnavailable)
		}
		row.DisplayName = req.DisplayName
		row.AltText = req.AltText
		if err := s.db.Save(&row).Error; err != nil {
			return AssetItem{}, apperrors.New(apperrors.CodeDatabaseUnavailable)
		}
		return assetFromModel(row), nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for index, item := range s.items {
		if item.ID == parsed {
			item.DisplayName = req.DisplayName
			item.AltText = req.AltText
			item.UpdatedAt = time.Now().UTC()
			s.items[index] = item
			return item, nil
		}
	}
	return AssetItem{}, apperrors.New(apperrors.CodeResourceNotFound)
}

func (s *AssetService) Delete(id string) error {
	parsed, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return apperrors.New(apperrors.CodeInvalidParameter)
	}

	if s.db != nil {
		var row model.Asset
		err := s.db.First(&row, parsed).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(apperrors.CodeResourceNotFound)
		}
		if err != nil {
			return apperrors.New(apperrors.CodeDatabaseUnavailable)
		}
		if row.RefCount > 0 {
			return apperrors.New(apperrors.CodeReferencedResourceUsed)
		}
		if err := s.db.Delete(&row).Error; err != nil {
			return apperrors.New(apperrors.CodeDatabaseUnavailable)
		}
		// 删库后清理对象，失败不阻塞业务（与原行为一致）。
		_ = s.store.Delete(context.Background(), row.StorageKey)
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for index, item := range s.items {
		if item.ID == parsed {
			if item.RefCount > 0 {
				return apperrors.New(apperrors.CodeReferencedResourceUsed)
			}
			s.items = append(s.items[:index], s.items[index+1:]...)
			_ = s.store.Delete(context.Background(), item.StorageKey)
			return nil
		}
	}
	return apperrors.New(apperrors.CodeResourceNotFound)
}

func (s *AssetService) ReferenceList(id string, page int, pageSize int) ([]AssetReferenceItem, pagination.ListPage, error) {
	if _, err := strconv.ParseUint(id, 10, 64); err != nil {
		return nil, pagination.ListPage{}, apperrors.New(apperrors.CodeInvalidParameter)
	}
	normalizedPage := pagination.NormalizePage(page)
	normalizedSize := pagination.NormalizePageSize(pageSize)
	return []AssetReferenceItem{}, pagination.ListPage{
		Page:     normalizedPage,
		PageSize: normalizedSize,
		Total:    0, // 存根实现，引用关系尚未启用
		HasMore:  false,
	}, nil
}

func (s *AssetService) listFromDB(query AssetListQuery, page, pageSize int) ([]AssetItem, pagination.ListPage, error) {
	dbQuery := s.db.Model(&model.Asset{})
	if query.Keyword != "" {
		dbQuery = dbQuery.Where("display_name LIKE ? OR alt_text LIKE ?", "%"+query.Keyword+"%", "%"+query.Keyword+"%")
	}
	if query.Mime != "" {
		dbQuery = dbQuery.Where("mime_type LIKE ?", query.Mime+"%")
	}

	// COUNT 复用与 Find 完全一致的 Where 条件，不加 Limit / Offset
	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, pagination.ListPage{}, apperrors.New(apperrors.CodeDatabaseUnavailable)
	}

	offset := (page - 1) * pageSize
	var rows []model.Asset
	err := dbQuery.Order("created_at DESC, id DESC").
		Limit(pageSize + 1).
		Offset(offset).
		Find(&rows).Error
	if err != nil {
		return nil, pagination.ListPage{}, apperrors.New(apperrors.CodeDatabaseUnavailable)
	}
	hasMore := len(rows) > pageSize
	if hasMore {
		rows = rows[:pageSize]
	}
	items := make([]AssetItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, assetFromModel(row))
	}
	return items, pagination.ListPage{
		Page:     page,
		PageSize: pageSize,
		Total:    int(total),
		HasMore:  hasMore,
	}, nil
}

func (s *AssetService) listInMemory(query AssetListQuery, page, pageSize int) ([]AssetItem, pagination.ListPage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]AssetItem, 0, len(s.items))
	for _, item := range s.items {
		if query.Mime != "" && !strings.HasPrefix(item.MimeType, query.Mime) {
			continue
		}
		if query.Keyword != "" {
			keyword := strings.ToLower(query.Keyword)
			if !strings.Contains(strings.ToLower(item.DisplayName), keyword) && !strings.Contains(strings.ToLower(item.AltText), keyword) {
				continue
			}
		}
		items = append(items, item)
	}
	items = sortAssetsByCreatedAt(items)

	offset := (page - 1) * pageSize
	if offset >= len(items) {
		return []AssetItem{}, pagination.ListPage{Page: page, PageSize: pageSize, Total: len(items), HasMore: false}, nil
	}
	end := offset + pageSize
	hasMore := end < len(items)
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end], pagination.ListPage{
		Page:     page,
		PageSize: pageSize,
		Total:    len(items), // 内存模式：items 已是符合条件全集
		HasMore:  hasMore,
	}, nil
}

func assetFromModel(row model.Asset) AssetItem {
	return AssetItem{
		ID:          row.ID,
		DisplayName: row.DisplayName,
		AltText:     row.AltText,
		StorageKey:  row.StorageKey,
		URL:         row.URL,
		ThumbURL:    row.ThumbURL,
		MimeType:    row.MimeType,
		Ext:         row.Ext,
		Size:        row.Size,
		Width:       row.Width,
		Height:      row.Height,
		SHA256:      row.SHA256,
		Status:      row.Status,
		RefCount:    row.RefCount,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

func assetModelFromItem(item AssetItem) model.Asset {
	return model.Asset{
		DisplayName: item.DisplayName,
		AltText:     item.AltText,
		StorageKey:  item.StorageKey,
		URL:         item.URL,
		ThumbURL:    item.ThumbURL,
		MimeType:    item.MimeType,
		Ext:         item.Ext,
		Size:        item.Size,
		Width:       item.Width,
		Height:      item.Height,
		SHA256:      item.SHA256,
		Status:      item.Status,
		RefCount:    item.RefCount,
	}
}

func assetMimeAllowed(mimeType string) bool {
	_, allowed := assetExtensionForMIME(mimeType)
	return allowed
}

// assetExtensionForMIME 由实际文件内容决定落盘扩展名，不能信任上传文件名。
// 这同时避免图片伪装成 .html/.svg 后被同源静态服务按可执行类型返回。
func assetExtensionForMIME(mimeType string) (string, bool) {
	extensions := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/gif":  ".gif",
		"image/webp": ".webp",
	}
	ext, allowed := extensions[mimeType]
	return ext, allowed
}

// decodeImageDimensions 从已读取的字节流解码尺寸，避免本地/S3 模式下 IO 语义差异。
// 仅 jpeg/png/gif 由标准库 image 支持；webp 与未识别类型返回 0,0（不影响功能）。
func decodeImageDimensions(data []byte, mimeType string) (uint, uint) {
	if mimeType != "image/jpeg" && mimeType != "image/png" && mimeType != "image/gif" {
		return 0, 0
	}
	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return 0, 0
	}
	return uint(config.Width), uint(config.Height)
}

func sortAssetsByCreatedAt(items []AssetItem) []AssetItem {
	sorted := append([]AssetItem(nil), items...)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].CreatedAt.After(sorted[i].CreatedAt) ||
				(sorted[j].CreatedAt.Equal(sorted[i].CreatedAt) && sorted[j].ID > sorted[i].ID) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	return sorted
}

func randomAssetName() string {
	raw := make([]byte, 16)
	if _, err := rand.Read(raw); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 36)
	}
	return hex.EncodeToString(raw)
}
