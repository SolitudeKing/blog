package service

import (
	"errors"
	"strconv"
	"sync"
	"time"

	"gorm.io/gorm"

	apperrors "solitude-blog/server/internal/errors"
	"solitude-blog/server/internal/model"
)

type TagService struct {
	db    *gorm.DB
	mu    sync.RWMutex
	items []TagItem
}

type TagItem struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	Color       string    `json:"color"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TagSaveRequest struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

func NewTagService(db *gorm.DB) *TagService {
	now := time.Now().UTC()
	return &TagService{
		db: db,
		items: []TagItem{
			{ID: 1, Name: "Go", Slug: "go", Color: "#5f8d62", CreatedAt: now, UpdatedAt: now},
			{ID: 2, Name: "Vue", Slug: "vue", Color: "#557ea8", CreatedAt: now, UpdatedAt: now},
		},
	}
}

func (s *TagService) List() ([]TagItem, error) {
	if s.db != nil {
		var rows []model.Tag
		if err := s.db.Order("created_at DESC, id DESC").Find(&rows).Error; err != nil {
			return nil, apperrors.New(apperrors.CodeDatabaseUnavailable)
		}
		items := make([]TagItem, 0, len(rows))
		for _, row := range rows {
			items = append(items, tagFromModel(row))
		}
		return items, nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]TagItem, len(s.items))
	copy(items, s.items)
	return items, nil
}

func (s *TagService) Create(req TagSaveRequest) (TagItem, error) {
	if req.Name == "" || req.Slug == "" {
		return TagItem{}, apperrors.New(apperrors.CodeMissingRequiredField)
	}
	if s.db != nil {
		var count int64
		if err := s.db.Model(&model.Tag{}).Where("slug = ?", req.Slug).Count(&count).Error; err != nil {
			return TagItem{}, apperrors.New(apperrors.CodeDatabaseUnavailable)
		}
		if count > 0 {
			return TagItem{}, apperrors.New(apperrors.CodeResourceConflict)
		}
		row := model.Tag{
			Name:        req.Name,
			Slug:        req.Slug,
			Description: req.Description,
			Color:       req.Color,
		}
		if err := s.db.Create(&row).Error; err != nil {
			return TagItem{}, apperrors.New(apperrors.CodeDatabaseUnavailable)
		}
		return tagFromModel(row), nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, item := range s.items {
		if item.Slug == req.Slug {
			return TagItem{}, apperrors.New(apperrors.CodeResourceConflict)
		}
	}
	now := time.Now().UTC()
	item := TagItem{
		ID:          uint64(len(s.items) + 1),
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Color:       req.Color,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	s.items = append(s.items, item)
	return item, nil
}

func (s *TagService) Update(id string, req TagSaveRequest) (TagItem, error) {
	parsed, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return TagItem{}, apperrors.New(apperrors.CodeInvalidParameter)
	}
	if req.Name == "" || req.Slug == "" {
		return TagItem{}, apperrors.New(apperrors.CodeMissingRequiredField)
	}
	if s.db != nil {
		var row model.Tag
		err := s.db.First(&row, parsed).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return TagItem{}, apperrors.New(apperrors.CodeResourceNotFound)
		}
		if err != nil {
			return TagItem{}, apperrors.New(apperrors.CodeDatabaseUnavailable)
		}
		var count int64
		if err := s.db.Model(&model.Tag{}).
			Where("slug = ? AND id <> ?", req.Slug, parsed).
			Count(&count).Error; err != nil {
			return TagItem{}, apperrors.New(apperrors.CodeDatabaseUnavailable)
		}
		if count > 0 {
			return TagItem{}, apperrors.New(apperrors.CodeResourceConflict)
		}
		row.Name = req.Name
		row.Slug = req.Slug
		row.Description = req.Description
		row.Color = req.Color
		if err := s.db.Save(&row).Error; err != nil {
			return TagItem{}, apperrors.New(apperrors.CodeDatabaseUnavailable)
		}
		return tagFromModel(row), nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, item := range s.items {
		if item.Slug == req.Slug && item.ID != parsed {
			return TagItem{}, apperrors.New(apperrors.CodeResourceConflict)
		}
	}
	for index, item := range s.items {
		if item.ID == parsed {
			item.Name = req.Name
			item.Slug = req.Slug
			item.Description = req.Description
			item.Color = req.Color
			item.UpdatedAt = time.Now().UTC()
			s.items[index] = item
			return item, nil
		}
	}
	return TagItem{}, apperrors.New(apperrors.CodeResourceNotFound)
}

func (s *TagService) Delete(id string) error {
	parsed, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return apperrors.New(apperrors.CodeInvalidParameter)
	}
	if s.db != nil {
		var count int64
		if err := s.db.Table("article_tags").
			Joins("JOIN articles ON articles.id = article_tags.article_id").
			Where("article_tags.tag_id = ? AND articles.deleted_at IS NULL", parsed).
			Count(&count).Error; err != nil {
			return apperrors.New(apperrors.CodeDatabaseUnavailable)
		}
		if count > 0 {
			return apperrors.New(apperrors.CodeReferencedResourceUsed)
		}
		result := s.db.Delete(&model.Tag{}, parsed)
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

func tagFromModel(row model.Tag) TagItem {
	return TagItem{
		ID:          row.ID,
		Name:        row.Name,
		Slug:        row.Slug,
		Description: row.Description,
		Color:       row.Color,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}
