package service

import (
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"

	apperrors "solitude-blog/server/internal/errors"
	"solitude-blog/server/internal/model"
	"solitude-blog/server/internal/pagination"
)

type NoticeService struct {
	db    *gorm.DB
	mu    sync.RWMutex
	items []NoticeItem
}

type NoticeListQuery struct {
	Page     int
	PageSize int
	Keyword  string
	Enabled  *bool
}

type NoticeItem struct {
	ID        uint64     `json:"id"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	Enabled   bool       `json:"enabled"`
	SortOrder int        `json:"sort_order"`
	StartsAt  *time.Time `json:"starts_at"`
	EndsAt    *time.Time `json:"ends_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type NoticeSaveRequest struct {
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	Enabled   bool       `json:"enabled"`
	SortOrder int        `json:"sort_order"`
	StartsAt  *time.Time `json:"starts_at"`
	EndsAt    *time.Time `json:"ends_at"`
}

func NewNoticeService(db *gorm.DB) *NoticeService {
	return &NoticeService{
		db:    db,
		items: []NoticeItem{},
	}
}

func (s *NoticeService) Active() (*NoticeItem, error) {
	now := time.Now().UTC()
	if s.db != nil {
		var row model.Notice
		err := s.db.
			Where("enabled = ?", true).
			Where("(starts_at IS NULL OR starts_at <= ?) AND (ends_at IS NULL OR ends_at > ?)", now, now).
			Order("sort_order ASC, created_at DESC, id DESC").
			First(&row).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		if err != nil {
			return nil, apperrors.New(apperrors.CodeDatabaseUnavailable)
		}
		item := noticeFromModel(row)
		return &item, nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, item := range sortNotices(s.items) {
		if noticeIsActive(item, now) {
			cloned := item
			return &cloned, nil
		}
	}
	return nil, nil
}

func (s *NoticeService) ManageList(query NoticeListQuery) ([]NoticeItem, pagination.ListPage, error) {
	page := pagination.NormalizePage(query.Page)
	pageSize := pagination.NormalizePageSize(query.PageSize)
	if s.db != nil {
		return s.listFromDB(query, page, pageSize)
	}
	return s.listInMemory(query, page, pageSize)
}

func (s *NoticeService) Create(req NoticeSaveRequest) (NoticeItem, error) {
	normalized, err := normalizeNotice(req)
	if err != nil {
		return NoticeItem{}, err
	}
	if s.db != nil {
		row := model.Notice{
			Title:     normalized.Title,
			Content:   normalized.Content,
			Enabled:   normalized.Enabled,
			SortOrder: normalized.SortOrder,
			StartsAt:  normalized.StartsAt,
			EndsAt:    normalized.EndsAt,
		}
		if err := s.db.Create(&row).Error; err != nil {
			return NoticeItem{}, apperrors.New(apperrors.CodeDatabaseUnavailable)
		}
		return noticeFromModel(row), nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	item := NoticeItem{
		ID:        uint64(len(s.items) + 1),
		Title:     normalized.Title,
		Content:   normalized.Content,
		Enabled:   normalized.Enabled,
		SortOrder: normalized.SortOrder,
		StartsAt:  normalized.StartsAt,
		EndsAt:    normalized.EndsAt,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.items = append(s.items, item)
	return item, nil
}

func (s *NoticeService) Update(id string, req NoticeSaveRequest) (NoticeItem, error) {
	parsed, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return NoticeItem{}, apperrors.New(apperrors.CodeInvalidParameter)
	}
	normalized, err := normalizeNotice(req)
	if err != nil {
		return NoticeItem{}, err
	}
	if s.db != nil {
		var row model.Notice
		err := s.db.First(&row, parsed).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NoticeItem{}, apperrors.New(apperrors.CodeResourceNotFound)
		}
		if err != nil {
			return NoticeItem{}, apperrors.New(apperrors.CodeDatabaseUnavailable)
		}
		row.Title = normalized.Title
		row.Content = normalized.Content
		row.Enabled = normalized.Enabled
		row.SortOrder = normalized.SortOrder
		row.StartsAt = normalized.StartsAt
		row.EndsAt = normalized.EndsAt
		if err := s.db.Save(&row).Error; err != nil {
			return NoticeItem{}, apperrors.New(apperrors.CodeDatabaseUnavailable)
		}
		return noticeFromModel(row), nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for index, item := range s.items {
		if item.ID == parsed {
			item.Title = normalized.Title
			item.Content = normalized.Content
			item.Enabled = normalized.Enabled
			item.SortOrder = normalized.SortOrder
			item.StartsAt = normalized.StartsAt
			item.EndsAt = normalized.EndsAt
			item.UpdatedAt = time.Now().UTC()
			s.items[index] = item
			return item, nil
		}
	}
	return NoticeItem{}, apperrors.New(apperrors.CodeResourceNotFound)
}

func (s *NoticeService) Delete(id string) error {
	parsed, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return apperrors.New(apperrors.CodeInvalidParameter)
	}
	if s.db != nil {
		result := s.db.Delete(&model.Notice{}, parsed)
		if result.Error != nil {
			return apperrors.New(apperrors.CodeDatabaseUnavailable)
		}
		if result.RowsAffected == 0 {
			return apperrors.New(apperrors.CodeResourceNotFound)
		}
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for index, item := range s.items {
		if item.ID == parsed {
			s.items = append(s.items[:index], s.items[index+1:]...)
			return nil
		}
	}
	return apperrors.New(apperrors.CodeResourceNotFound)
}

func (s *NoticeService) listFromDB(query NoticeListQuery, page, pageSize int) ([]NoticeItem, pagination.ListPage, error) {
	dbQuery := s.db.Model(&model.Notice{})
	if query.Keyword != "" {
		dbQuery = dbQuery.Where("title LIKE ? OR content LIKE ?", "%"+query.Keyword+"%", "%"+query.Keyword+"%")
	}
	if query.Enabled != nil {
		dbQuery = dbQuery.Where("enabled = ?", *query.Enabled)
	}

	offset := (page - 1) * pageSize
	var rows []model.Notice
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
	items := make([]NoticeItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, noticeFromModel(row))
	}
	return items, pagination.ListPage{
		Page:     page,
		PageSize: pageSize,
		HasMore:  hasMore,
	}, nil
}

func (s *NoticeService) listInMemory(query NoticeListQuery, page, pageSize int) ([]NoticeItem, pagination.ListPage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]NoticeItem, 0, len(s.items))
	for _, item := range s.items {
		if query.Enabled != nil && item.Enabled != *query.Enabled {
			continue
		}
		if query.Keyword != "" {
			keyword := strings.ToLower(query.Keyword)
			if !strings.Contains(strings.ToLower(item.Title), keyword) && !strings.Contains(strings.ToLower(item.Content), keyword) {
				continue
			}
		}
		items = append(items, item)
	}
	items = sortNoticesByCreatedAt(items)

	offset := (page - 1) * pageSize
	if offset >= len(items) {
		return []NoticeItem{}, pagination.ListPage{Page: page, PageSize: pageSize, HasMore: false}, nil
	}
	end := offset + pageSize
	hasMore := end < len(items)
	if end > len(items) {
		end = len(items)
	}
	return items[offset:end], pagination.ListPage{
		Page:     page,
		PageSize: pageSize,
		HasMore:  hasMore,
	}, nil
}

func normalizeNotice(req NoticeSaveRequest) (NoticeSaveRequest, error) {
	if req.Title == "" || req.Content == "" {
		return NoticeSaveRequest{}, apperrors.New(apperrors.CodeMissingRequiredField)
	}
	if req.StartsAt != nil && req.EndsAt != nil && !req.StartsAt.Before(*req.EndsAt) {
		return NoticeSaveRequest{}, apperrors.New(apperrors.CodeInvalidParameter)
	}
	return req, nil
}

func noticeFromModel(row model.Notice) NoticeItem {
	return NoticeItem{
		ID:        row.ID,
		Title:     row.Title,
		Content:   row.Content,
		Enabled:   row.Enabled,
		SortOrder: row.SortOrder,
		StartsAt:  row.StartsAt,
		EndsAt:    row.EndsAt,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

func noticeIsActive(item NoticeItem, now time.Time) bool {
	if !item.Enabled {
		return false
	}
	if item.StartsAt != nil && item.StartsAt.After(now) {
		return false
	}
	if item.EndsAt != nil && !item.EndsAt.After(now) {
		return false
	}
	return true
}

func sortNotices(items []NoticeItem) []NoticeItem {
	sorted := append([]NoticeItem(nil), items...)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].SortOrder < sorted[i].SortOrder ||
				(sorted[j].SortOrder == sorted[i].SortOrder && sorted[j].CreatedAt.After(sorted[i].CreatedAt)) ||
				(sorted[j].SortOrder == sorted[i].SortOrder && sorted[j].CreatedAt.Equal(sorted[i].CreatedAt) && sorted[j].ID > sorted[i].ID) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	return sorted
}

func sortNoticesByCreatedAt(items []NoticeItem) []NoticeItem {
	sorted := append([]NoticeItem(nil), items...)
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
