package service

import (
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
)

const maxAssetUploadSize int64 = 10 * 1024 * 1024

type AssetService struct {
	db      *gorm.DB
	rootDir string
	mu      sync.RWMutex
	items   []AssetItem
}

type AssetListQuery struct {
	Cursor  string
	Limit   int
	Keyword string
	Mime    string
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

func NewAssetService(db *gorm.DB, rootDir string) *AssetService {
	return &AssetService{db: db, rootDir: rootDir}
}

func (s *AssetService) List(query AssetListQuery) ([]AssetItem, pagination.CursorPage, error) {
	limit := pagination.NormalizeLimit(query.Limit)
	if s.db != nil {
		return s.listFromDB(query, limit)
	}
	return s.listInMemory(query, limit)
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

	head := make([]byte, 512)
	n, _ := source.Read(head)
	mimeType := http.DetectContentType(head[:n])
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext == ".svg" {
		mimeType = "image/svg+xml"
	}
	if !assetMimeAllowed(mimeType) {
		return AssetItem{}, apperrors.New(apperrors.CodeUnsupportedFileType)
	}
	if seeker, ok := source.(io.Seeker); ok {
		if _, err := seeker.Seek(0, io.SeekStart); err != nil {
			return AssetItem{}, apperrors.New(apperrors.CodeInvalidRequest)
		}
	} else {
		return AssetItem{}, apperrors.New(apperrors.CodeInvalidRequest)
	}

	now := time.Now().UTC()
	storageKey := filepath.ToSlash(filepath.Join(
		now.Format("2006"),
		now.Format("01"),
		randomAssetName()+ext,
	))
	fullPath := filepath.Join(s.rootDir, filepath.FromSlash(storageKey))
	if err := ensureParentDir(fullPath); err != nil {
		return AssetItem{}, apperrors.New(apperrors.CodeStorageUnavailable)
	}

	target, err := createFile(fullPath)
	if err != nil {
		return AssetItem{}, apperrors.New(apperrors.CodeStorageUnavailable)
	}
	hash := sha256.New()
	written, copyErr := io.Copy(io.MultiWriter(target, hash), source)
	closeErr := target.Close()
	if copyErr != nil || closeErr != nil {
		return AssetItem{}, apperrors.New(apperrors.CodeStorageUnavailable)
	}

	width, height := imageDimensions(fullPath, mimeType)
	if displayName == "" {
		displayName = fileHeader.Filename
	}
	item := AssetItem{
		DisplayName: displayName,
		StorageKey:  storageKey,
		URL:         "/uploads/" + storageKey,
		ThumbURL:    "",
		MimeType:    mimeType,
		Ext:         strings.TrimPrefix(ext, "."),
		Size:        uint64(written),
		Width:       width,
		Height:      height,
		SHA256:      hex.EncodeToString(hash.Sum(nil)),
		Status:      "ready",
		RefCount:    0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if s.db != nil {
		row := assetModelFromItem(item)
		if err := s.db.Create(&row).Error; err != nil {
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
		_ = removeFile(filepath.Join(s.rootDir, filepath.FromSlash(row.StorageKey)))
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
			_ = removeFile(filepath.Join(s.rootDir, filepath.FromSlash(item.StorageKey)))
			return nil
		}
	}
	return apperrors.New(apperrors.CodeResourceNotFound)
}

func (s *AssetService) ReferenceList(id string, limit int) ([]AssetReferenceItem, pagination.CursorPage, error) {
	if _, err := strconv.ParseUint(id, 10, 64); err != nil {
		return nil, pagination.CursorPage{}, apperrors.New(apperrors.CodeInvalidParameter)
	}
	pageLimit := pagination.NormalizeLimit(limit)
	return []AssetReferenceItem{}, pagination.CursorPage{
		Limit:   pageLimit,
		HasMore: false,
	}, nil
}

func (s *AssetService) listFromDB(query AssetListQuery, limit int) ([]AssetItem, pagination.CursorPage, error) {
	dbQuery := s.db.Model(&model.Asset{})
	if query.Keyword != "" {
		dbQuery = dbQuery.Where("display_name LIKE ? OR alt_text LIKE ?", "%"+query.Keyword+"%", "%"+query.Keyword+"%")
	}
	if query.Mime != "" {
		dbQuery = dbQuery.Where("mime_type LIKE ?", query.Mime+"%")
	}
	if query.Cursor != "" {
		cursor, err := pagination.DecodeCursor(query.Cursor)
		if err != nil {
			return nil, pagination.CursorPage{}, apperrors.New(apperrors.CodeInvalidCursor)
		}
		dbQuery = dbQuery.Where("created_at < ? OR (created_at = ? AND id < ?)", cursor.CreatedAt, cursor.CreatedAt, cursor.ID)
	}

	var rows []model.Asset
	err := dbQuery.Order("created_at DESC, id DESC").Limit(limit + 1).Find(&rows).Error
	if err != nil {
		return nil, pagination.CursorPage{}, apperrors.New(apperrors.CodeDatabaseUnavailable)
	}
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	items := make([]AssetItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, assetFromModel(row))
	}
	nextCursor := ""
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		var err error
		nextCursor, err = pagination.EncodeCursor(pagination.CursorPayload{CreatedAt: last.CreatedAt, ID: last.ID})
		if err != nil {
			return nil, pagination.CursorPage{}, apperrors.New(apperrors.CodeInternalServerError)
		}
	}
	return items, pagination.CursorPage{
		Cursor:     query.Cursor,
		NextCursor: nextCursor,
		Limit:      limit,
		HasMore:    hasMore,
	}, nil
}

func (s *AssetService) listInMemory(query AssetListQuery, limit int) ([]AssetItem, pagination.CursorPage, error) {
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
	end := limit
	hasMore := len(items) > limit
	if end > len(items) {
		end = len(items)
	}
	pageItems := items[:end]
	nextCursor := ""
	if hasMore && len(pageItems) > 0 {
		last := pageItems[len(pageItems)-1]
		var err error
		nextCursor, err = pagination.EncodeCursor(pagination.CursorPayload{CreatedAt: last.CreatedAt, ID: last.ID})
		if err != nil {
			return nil, pagination.CursorPage{}, apperrors.New(apperrors.CodeInternalServerError)
		}
	}
	return pageItems, pagination.CursorPage{
		Cursor:     query.Cursor,
		NextCursor: nextCursor,
		Limit:      limit,
		HasMore:    hasMore,
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
	switch mimeType {
	case "image/jpeg", "image/png", "image/gif", "image/webp", "image/svg+xml":
		return true
	default:
		return false
	}
}

func imageDimensions(path string, mimeType string) (uint, uint) {
	if mimeType != "image/jpeg" && mimeType != "image/png" && mimeType != "image/gif" {
		return 0, 0
	}
	file, err := openFile(path)
	if err != nil {
		return 0, 0
	}
	defer file.Close()
	config, _, err := image.DecodeConfig(file)
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
