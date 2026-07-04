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

type CategoryService struct {
	db    *gorm.DB
	mu    sync.RWMutex
	items []CategoryItem
}

type CategoryItem struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CategorySaveRequest struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
}

func NewCategoryService(db *gorm.DB) *CategoryService {
	now := time.Now().UTC()
	return &CategoryService{
		db: db,
		items: []CategoryItem{
			{
				ID:        1,
				Name:      "Notes",
				Slug:      "notes",
				SortOrder: 1,
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
	}
}

func (s *CategoryService) List() ([]CategoryItem, error) {
	if s.db != nil {
		var rows []model.Category
		if err := s.db.Order("sort_order ASC, created_at DESC, id DESC").Find(&rows).Error; err != nil {
			return nil, apperrors.New(apperrors.CodeDatabaseUnavailable)
		}
		items := make([]CategoryItem, 0, len(rows))
		for _, row := range rows {
			items = append(items, categoryFromModel(row))
		}
		return items, nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]CategoryItem, len(s.items))
	copy(items, s.items)
	return items, nil
}

func (s *CategoryService) Create(req CategorySaveRequest) (CategoryItem, error) {
	if req.Name == "" || req.Slug == "" {
		return CategoryItem{}, apperrors.New(apperrors.CodeMissingRequiredField)
	}
	if s.db != nil {
		var count int64
		if err := s.db.Model(&model.Category{}).Where("slug = ?", req.Slug).Count(&count).Error; err != nil {
			return CategoryItem{}, apperrors.New(apperrors.CodeDatabaseUnavailable)
		}
		if count > 0 {
			return CategoryItem{}, apperrors.New(apperrors.CodeResourceConflict)
		}
		row := model.Category{
			Name:        req.Name,
			Slug:        req.Slug,
			Description: req.Description,
			SortOrder:   req.SortOrder,
		}
		if err := s.db.Create(&row).Error; err != nil {
			return CategoryItem{}, apperrors.New(apperrors.CodeDatabaseUnavailable)
		}
		return categoryFromModel(row), nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, item := range s.items {
		if item.Slug == req.Slug {
			return CategoryItem{}, apperrors.New(apperrors.CodeResourceConflict)
		}
	}
	now := time.Now().UTC()
	item := CategoryItem{
		ID:          uint64(len(s.items) + 1),
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		SortOrder:   req.SortOrder,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	s.items = append(s.items, item)
	return item, nil
}

func (s *CategoryService) Update(id string, req CategorySaveRequest) (CategoryItem, error) {
	parsed, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return CategoryItem{}, apperrors.New(apperrors.CodeInvalidParameter)
	}
	if req.Name == "" || req.Slug == "" {
		return CategoryItem{}, apperrors.New(apperrors.CodeMissingRequiredField)
	}
	if s.db != nil {
		var row model.Category
		err := s.db.First(&row, parsed).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return CategoryItem{}, apperrors.New(apperrors.CodeResourceNotFound)
		}
		if err != nil {
			return CategoryItem{}, apperrors.New(apperrors.CodeDatabaseUnavailable)
		}
		var count int64
		if err := s.db.Model(&model.Category{}).
			Where("slug = ? AND id <> ?", req.Slug, parsed).
			Count(&count).Error; err != nil {
			return CategoryItem{}, apperrors.New(apperrors.CodeDatabaseUnavailable)
		}
		if count > 0 {
			return CategoryItem{}, apperrors.New(apperrors.CodeResourceConflict)
		}
		row.Name = req.Name
		row.Slug = req.Slug
		row.Description = req.Description
		row.SortOrder = req.SortOrder
		if err := s.db.Save(&row).Error; err != nil {
			return CategoryItem{}, apperrors.New(apperrors.CodeDatabaseUnavailable)
		}
		return categoryFromModel(row), nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, item := range s.items {
		if item.Slug == req.Slug && item.ID != parsed {
			return CategoryItem{}, apperrors.New(apperrors.CodeResourceConflict)
		}
	}
	for index, item := range s.items {
		if item.ID == parsed {
			item.Name = req.Name
			item.Slug = req.Slug
			item.Description = req.Description
			item.SortOrder = req.SortOrder
			item.UpdatedAt = time.Now().UTC()
			s.items[index] = item
			return item, nil
		}
	}
	return CategoryItem{}, apperrors.New(apperrors.CodeResourceNotFound)
}

func (s *CategoryService) Delete(id string) error {
	parsed, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return apperrors.New(apperrors.CodeInvalidParameter)
	}
	if s.db != nil {
		var count int64
		if err := s.db.Model(&model.Article{}).Where("category_id = ?", parsed).Count(&count).Error; err != nil {
			return apperrors.New(apperrors.CodeDatabaseUnavailable)
		}
		if count > 0 {
			return apperrors.New(apperrors.CodeReferencedResourceUsed)
		}
		result := s.db.Delete(&model.Category{}, parsed)
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

func categoryFromModel(row model.Category) CategoryItem {
	return CategoryItem{
		ID:          row.ID,
		Name:        row.Name,
		Slug:        row.Slug,
		Description: row.Description,
		SortOrder:   row.SortOrder,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}
