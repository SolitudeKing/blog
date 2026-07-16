package service

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"solitude-blog/server/internal/cache"
	apperrors "solitude-blog/server/internal/errors"
	"solitude-blog/server/internal/model"
)

type TopicService struct {
	db    *gorm.DB
	redis *redis.Client
	mu    sync.RWMutex
	items []TopicItem
}

type TopicItem struct {
	ID           uint64    `json:"id"`
	Name         string    `json:"name"`
	Label        string    `json:"label"`
	Slug         string    `json:"slug"`
	Description  string    `json:"description"`
	CoverURL     string    `json:"cover_url"`
	SortOrder    int       `json:"sort_order"`
	ArticleCount int64     `json:"article_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type TopicSaveRequest struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	CoverURL    string `json:"cover_url"`
	SortOrder   int    `json:"sort_order"`
}

func NewTopicService(db *gorm.DB, redisClient *redis.Client) *TopicService {
	now := time.Now().UTC()
	return &TopicService{
		db:    db,
		redis: redisClient,
		items: []TopicItem{
			{
				ID:        1,
				Name:      "Notes",
				Label:     "Notes",
				Slug:      "notes",
				SortOrder: 1,
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
	}
}

func (s *TopicService) List() ([]TopicItem, error) {
	if s.db != nil {
		var items []TopicItem
		if err := s.db.Table("topics").
			Select(`topics.id, topics.name, topics.label, topics.slug, topics.description,
				topics.cover_url, topics.sort_order, topics.created_at, topics.updated_at,
				COUNT(articles.id) AS article_count`).
			Joins(`LEFT JOIN articles ON articles.topic_id = topics.id
				AND articles.status = 'published' AND articles.deleted_at IS NULL`).
			Where("topics.deleted_at IS NULL").
			Group(`topics.id, topics.name, topics.label, topics.slug, topics.description,
				topics.cover_url, topics.sort_order, topics.created_at, topics.updated_at`).
			Order("topics.sort_order ASC, topics.created_at DESC, topics.id DESC").
			Scan(&items).Error; err != nil {
			return nil, apperrors.New(apperrors.CodeDatabaseUnavailable)
		}
		return items, nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	items := make([]TopicItem, len(s.items))
	copy(items, s.items)
	return items, nil
}

func (s *TopicService) Create(req TopicSaveRequest) (TopicItem, error) {
	req, err := normalizeTopicSaveRequest(req)
	if err != nil {
		return TopicItem{}, err
	}
	if s.db != nil {
		var count int64
		if err := s.db.Unscoped().Model(&model.Topic{}).Where("slug = ?", req.Slug).Count(&count).Error; err != nil {
			return TopicItem{}, apperrors.New(apperrors.CodeDatabaseUnavailable)
		}
		if count > 0 {
			return TopicItem{}, apperrors.New(apperrors.CodeResourceConflict)
		}
		row := model.Topic{
			Name:        req.Name,
			Label:       req.Label,
			Slug:        req.Slug,
			Description: req.Description,
			CoverURL:    req.CoverURL,
			SortOrder:   req.SortOrder,
		}
		if err := s.db.Create(&row).Error; err != nil {
			return TopicItem{}, apperrors.New(apperrors.CodeDatabaseUnavailable)
		}
		return topicFromModel(row), nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, item := range s.items {
		if item.Slug == req.Slug {
			return TopicItem{}, apperrors.New(apperrors.CodeResourceConflict)
		}
	}
	now := time.Now().UTC()
	item := TopicItem{
		ID:          uint64(len(s.items) + 1),
		Name:        req.Name,
		Label:       req.Label,
		Slug:        req.Slug,
		Description: req.Description,
		CoverURL:    req.CoverURL,
		SortOrder:   req.SortOrder,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	s.items = append(s.items, item)
	return item, nil
}

func (s *TopicService) Update(id string, req TopicSaveRequest) (TopicItem, error) {
	parsed, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return TopicItem{}, apperrors.New(apperrors.CodeInvalidParameter)
	}
	req, err = normalizeTopicSaveRequest(req)
	if err != nil {
		return TopicItem{}, err
	}
	if s.db != nil {
		var row model.Topic
		err := s.db.First(&row, parsed).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return TopicItem{}, apperrors.New(apperrors.CodeResourceNotFound)
		}
		if err != nil {
			return TopicItem{}, apperrors.New(apperrors.CodeDatabaseUnavailable)
		}
		var count int64
		if err := s.db.Unscoped().Model(&model.Topic{}).
			Where("slug = ? AND id <> ?", req.Slug, parsed).
			Count(&count).Error; err != nil {
			return TopicItem{}, apperrors.New(apperrors.CodeDatabaseUnavailable)
		}
		if count > 0 {
			return TopicItem{}, apperrors.New(apperrors.CodeResourceConflict)
		}
		row.Name = req.Name
		row.Label = req.Label
		row.Slug = req.Slug
		row.Description = req.Description
		row.CoverURL = req.CoverURL
		row.SortOrder = req.SortOrder
		if err := s.db.Save(&row).Error; err != nil {
			return TopicItem{}, apperrors.New(apperrors.CodeDatabaseUnavailable)
		}
		s.invalidateArticleCaches()
		return topicFromModel(row), nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, item := range s.items {
		if item.Slug == req.Slug && item.ID != parsed {
			return TopicItem{}, apperrors.New(apperrors.CodeResourceConflict)
		}
	}
	for index, item := range s.items {
		if item.ID == parsed {
			item.Name = req.Name
			item.Label = req.Label
			item.Slug = req.Slug
			item.Description = req.Description
			item.CoverURL = req.CoverURL
			item.SortOrder = req.SortOrder
			item.UpdatedAt = time.Now().UTC()
			s.items[index] = item
			return item, nil
		}
	}
	return TopicItem{}, apperrors.New(apperrors.CodeResourceNotFound)
}

func (s *TopicService) invalidateArticleCaches() {
	if s.redis == nil {
		return
	}
	ctx := context.Background()
	keys := make([]string, 0)
	for _, pattern := range []string{cache.ArticleListPattern(), cache.ArticleDetailPattern()} {
		iter := s.redis.Scan(ctx, 0, pattern, 100).Iterator()
		for iter.Next(ctx) {
			keys = append(keys, iter.Val())
		}
	}
	if len(keys) > 0 {
		_ = s.redis.Del(ctx, keys...).Err()
	}
}

func (s *TopicService) Delete(id string) error {
	parsed, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return apperrors.New(apperrors.CodeInvalidParameter)
	}
	if s.db != nil {
		var count int64
		if err := s.db.Model(&model.Article{}).Where("topic_id = ?", parsed).Count(&count).Error; err != nil {
			return apperrors.New(apperrors.CodeDatabaseUnavailable)
		}
		if count > 0 {
			return apperrors.New(apperrors.CodeReferencedResourceUsed)
		}
		result := s.db.Delete(&model.Topic{}, parsed)
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

func normalizeTopicSaveRequest(req TopicSaveRequest) (TopicSaveRequest, error) {
	req.Name = strings.TrimSpace(req.Name)
	req.Label = strings.TrimSpace(req.Label)
	req.Slug = strings.TrimSpace(req.Slug)
	if req.Name == "" || req.Label == "" || req.Slug == "" {
		return TopicSaveRequest{}, apperrors.New(apperrors.CodeMissingRequiredField)
	}
	if utf8.RuneCountInString(req.Label) > model.TopicLabelMaxRunes {
		return TopicSaveRequest{}, apperrors.New(apperrors.CodeInvalidParameter)
	}
	return req, nil
}

func topicFromModel(row model.Topic) TopicItem {
	return TopicItem{
		ID:          row.ID,
		Name:        row.Name,
		Label:       row.Label,
		Slug:        row.Slug,
		Description: row.Description,
		CoverURL:    row.CoverURL,
		SortOrder:   row.SortOrder,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}
