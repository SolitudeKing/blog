package service

import (
	"context"
	"encoding/json"
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
	"solitude-blog/server/internal/pagination"
)

const articleCacheTTL = 5 * time.Minute

type ArticleService struct {
	db    *gorm.DB
	redis *redis.Client
	mu    sync.RWMutex
	items []ArticleDetail
}

type ArticleListQuery struct {
	Cursor  string
	Limit   int
	Topic   string
	Tag     string
	Keyword string
	Status  string
}

type TopicReference struct {
	ID    uint64 `json:"id"`
	Name  string `json:"name"`
	Label string `json:"label"`
	Slug  string `json:"slug"`
}

type ArticleItem struct {
	ID          uint64         `json:"id"`
	Title       string         `json:"title"`
	Slug        string         `json:"slug"`
	Summary     string         `json:"summary"`
	CoverURL    string         `json:"cover_url"`
	Status      string         `json:"status"`
	TopicID     uint64         `json:"topic_id"`
	Topic       TopicReference `json:"topic"`
	Tags        []string       `json:"tags"`
	ViewCount   uint64         `json:"view_count"`
	PublishedAt time.Time      `json:"published_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type ArticleDetail struct {
	ArticleItem
	ContentMD string   `json:"content_md"`
	TagIDs    []uint64 `json:"tag_ids"`
}

type ArticleVersionItem struct {
	ID        uint64    `json:"id"`
	ArticleID uint64    `json:"article_id"`
	Title     string    `json:"title"`
	Summary   string    `json:"summary"`
	ContentMD string    `json:"content_md"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type cachedArticleList struct {
	Items []ArticleItem         `json:"items"`
	Page  pagination.CursorPage `json:"page"`
}

type ArticleCreateRequest struct {
	Title     string   `json:"title"`
	Slug      string   `json:"slug"`
	Summary   string   `json:"summary"`
	CoverURL  string   `json:"cover_url"`
	ContentMD string   `json:"content_md"`
	TopicID   uint64   `json:"topic_id"`
	TagIDs    []uint64 `json:"tag_ids"`
	Status    string   `json:"status"`
}

type ArticleUpdateRequest = ArticleCreateRequest

func NewArticleService(db *gorm.DB, redisClient *redis.Client) *ArticleService {
	return &ArticleService{
		db:    db,
		redis: redisClient,
		items: []ArticleDetail{},
	}
}

func (s *ArticleService) PublicList(query ArticleListQuery) ([]ArticleItem, pagination.CursorPage, error) {
	if cached, ok := s.getCachedPublicList(query); ok {
		return cached.Items, cached.Page, nil
	}
	items, page, err := s.list(query, true)
	if err != nil {
		return nil, pagination.CursorPage{}, err
	}
	s.setCachedPublicList(query, cachedArticleList{Items: items, Page: page})
	return items, page, nil
}

func (s *ArticleService) ManageList(query ArticleListQuery) ([]ArticleItem, pagination.CursorPage, error) {
	return s.list(query, false)
}

func (s *ArticleService) Detail(slug string) (ArticleDetail, error) {
	if s.db != nil {
		if item, ok := s.getCachedDetail(slug); ok {
			return item, nil
		}
		var article model.Article
		err := s.db.Preload("Topic").Preload("Tags").
			Where("slug = ? AND status = ?", slug, "published").
			First(&article).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ArticleDetail{}, apperrors.New(apperrors.CodeResourceNotFound)
		}
		if err != nil {
			return ArticleDetail{}, apperrors.New(apperrors.CodeDatabaseUnavailable)
		}
		item := detailFromModel(article)
		s.setCachedDetail(slug, item)
		return item, nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, item := range s.items {
		if item.Slug == slug && item.Status == "published" {
			return item, nil
		}
	}
	return ArticleDetail{}, apperrors.New(apperrors.CodeResourceNotFound)
}

func (s *ArticleService) Info(id string) (ArticleDetail, error) {
	parsed, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return ArticleDetail{}, apperrors.New(apperrors.CodeInvalidParameter)
	}

	if s.db != nil {
		var article model.Article
		err := s.db.Preload("Topic").Preload("Tags").First(&article, parsed).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ArticleDetail{}, apperrors.New(apperrors.CodeResourceNotFound)
		}
		if err != nil {
			return ArticleDetail{}, apperrors.New(apperrors.CodeDatabaseUnavailable)
		}
		return detailFromModel(article), nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, item := range s.items {
		if item.ID == parsed {
			return item, nil
		}
	}
	return ArticleDetail{}, apperrors.New(apperrors.CodeResourceNotFound)
}

func (s *ArticleService) Create(req ArticleCreateRequest, actorID uint64) (ArticleItem, error) {
	if req.Title == "" || req.Slug == "" || req.TopicID == 0 {
		return ArticleItem{}, apperrors.New(apperrors.CodeMissingRequiredField)
	}
	if actorID == 0 {
		return ArticleItem{}, apperrors.New(apperrors.CodeUnauthorized)
	}
	status, err := normalizeArticleStatus(req.Status)
	if err != nil {
		return ArticleItem{}, err
	}
	coverURL, err := normalizeArticleCoverURL(req.CoverURL)
	if err != nil {
		return ArticleItem{}, err
	}

	if s.db != nil {
		var created ArticleDetail
		// 文章、标签关联与初始版本必须同时成功，避免接口报错后留下半成品文章。
		err := s.db.Transaction(func(tx *gorm.DB) error {
			topicID, err := resolveTopicID(tx, req.TopicID)
			if err != nil {
				return err
			}
			var count int64
			if err := tx.Unscoped().Model(&model.Article{}).Where("slug = ?", req.Slug).Count(&count).Error; err != nil {
				return apperrors.New(apperrors.CodeDatabaseUnavailable)
			}
			if count > 0 {
				return apperrors.New(apperrors.CodeDuplicateSlug)
			}
			tags, err := loadArticleTags(tx, req.TagIDs)
			if err != nil {
				return err
			}

			article := model.Article{
				Title:     req.Title,
				Slug:      req.Slug,
				Summary:   req.Summary,
				CoverURL:  coverURL,
				ContentMD: req.ContentMD,
				Status:    status,
				TopicID:   topicID,
				AuthorID:  actorID,
			}
			if status == "published" {
				now := time.Now().UTC()
				article.PublishedAt = &now
			}
			if err := tx.Create(&article).Error; err != nil {
				return apperrors.New(apperrors.CodeDatabaseUnavailable)
			}
			if err := tx.Model(&article).Association("Tags").Replace(tags); err != nil {
				return apperrors.New(apperrors.CodeDatabaseUnavailable)
			}
			if err := s.createVersion(tx, article, actorID); err != nil {
				return err
			}

			var loaded model.Article
			if err := tx.Preload("Topic").Preload("Tags").First(&loaded, article.ID).Error; err != nil {
				return apperrors.New(apperrors.CodeDatabaseUnavailable)
			}
			created = detailFromModel(loaded)
			return nil
		})
		if err != nil {
			return ArticleItem{}, normalizeArticleWriteError(err)
		}
		s.invalidateArticleCaches(created.Slug)
		return created.ArticleItem, nil
	}

	req.CoverURL = coverURL
	return s.createInMemory(req, status)
}

func (s *ArticleService) Update(id string, req ArticleUpdateRequest, actorID uint64) (ArticleDetail, error) {
	parsed, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return ArticleDetail{}, apperrors.New(apperrors.CodeInvalidParameter)
	}
	if req.Title == "" || req.Slug == "" || req.TopicID == 0 {
		return ArticleDetail{}, apperrors.New(apperrors.CodeMissingRequiredField)
	}
	if actorID == 0 {
		return ArticleDetail{}, apperrors.New(apperrors.CodeUnauthorized)
	}
	status, err := normalizeArticleStatus(req.Status)
	if err != nil {
		return ArticleDetail{}, err
	}
	coverURL, err := normalizeArticleCoverURL(req.CoverURL)
	if err != nil {
		return ArticleDetail{}, err
	}

	if s.db != nil {
		var detail ArticleDetail
		var oldSlug string
		// 正文、专题、标签与版本记录属于一次编辑操作，统一在事务中提交。
		err := s.db.Transaction(func(tx *gorm.DB) error {
			topicID, err := resolveTopicID(tx, req.TopicID)
			if err != nil {
				return err
			}
			var article model.Article
			err = tx.First(&article, parsed).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperrors.New(apperrors.CodeResourceNotFound)
			}
			if err != nil {
				return apperrors.New(apperrors.CodeDatabaseUnavailable)
			}
			oldSlug = article.Slug

			var slugCount int64
			if err := tx.Unscoped().Model(&model.Article{}).
				Where("slug = ? AND id <> ?", req.Slug, parsed).
				Count(&slugCount).Error; err != nil {
				return apperrors.New(apperrors.CodeDatabaseUnavailable)
			}
			if slugCount > 0 {
				return apperrors.New(apperrors.CodeDuplicateSlug)
			}
			tags, err := loadArticleTags(tx, req.TagIDs)
			if err != nil {
				return err
			}

			article.Title = req.Title
			article.Slug = req.Slug
			article.Summary = req.Summary
			article.CoverURL = coverURL
			article.ContentMD = req.ContentMD
			article.Status = status
			article.TopicID = topicID
			if status == "published" && article.PublishedAt == nil {
				now := time.Now().UTC()
				article.PublishedAt = &now
			}
			if err := tx.Save(&article).Error; err != nil {
				return apperrors.New(apperrors.CodeDatabaseUnavailable)
			}
			if err := tx.Model(&article).Association("Tags").Replace(tags); err != nil {
				return apperrors.New(apperrors.CodeDatabaseUnavailable)
			}
			if err := s.createVersion(tx, article, actorID); err != nil {
				return err
			}

			var loaded model.Article
			if err := tx.Preload("Topic").Preload("Tags").First(&loaded, article.ID).Error; err != nil {
				return apperrors.New(apperrors.CodeDatabaseUnavailable)
			}
			detail = detailFromModel(loaded)
			return nil
		})
		if err != nil {
			return ArticleDetail{}, normalizeArticleWriteError(err)
		}
		s.invalidateArticleCaches(oldSlug, detail.Slug)
		return detail, nil
	}

	req.CoverURL = coverURL
	return s.updateInMemory(parsed, req, status)
}

func resolveTopicID(db *gorm.DB, topicID uint64) (uint64, error) {
	if topicID == 0 {
		return 0, apperrors.New(apperrors.CodeMissingRequiredField)
	}
	var topic model.Topic
	err := db.Select("id").First(&topic, topicID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, apperrors.New(apperrors.CodeResourceNotFound)
	}
	if err != nil {
		return 0, apperrors.New(apperrors.CodeDatabaseUnavailable)
	}
	return topic.ID, nil
}

func (s *ArticleService) Delete(id string) error {
	parsed, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return apperrors.New(apperrors.CodeInvalidParameter)
	}

	if s.db != nil {
		var slug string
		// 标签关联与文章软删除必须保持一致，任一步失败都回滚。
		err := s.db.Transaction(func(tx *gorm.DB) error {
			var article model.Article
			err := tx.First(&article, parsed).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperrors.New(apperrors.CodeResourceNotFound)
			}
			if err != nil {
				return apperrors.New(apperrors.CodeDatabaseUnavailable)
			}
			slug = article.Slug
			if err := tx.Model(&article).Association("Tags").Clear(); err != nil {
				return apperrors.New(apperrors.CodeDatabaseUnavailable)
			}
			result := tx.Delete(&article)
			if result.Error != nil {
				return apperrors.New(apperrors.CodeDatabaseUnavailable)
			}
			if result.RowsAffected == 0 {
				return apperrors.New(apperrors.CodeResourceNotFound)
			}
			return nil
		})
		if err != nil {
			return normalizeArticleWriteError(err)
		}
		s.invalidateArticleCaches(slug)
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

func (s *ArticleService) VersionList(id string) ([]ArticleVersionItem, error) {
	parsed, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, apperrors.New(apperrors.CodeInvalidParameter)
	}
	if s.db == nil {
		return []ArticleVersionItem{}, nil
	}
	var count int64
	if err := s.db.Model(&model.Article{}).Where("id = ?", parsed).Count(&count).Error; err != nil {
		return nil, apperrors.New(apperrors.CodeDatabaseUnavailable)
	}
	if count == 0 {
		return nil, apperrors.New(apperrors.CodeResourceNotFound)
	}
	var rows []model.ArticleVersion
	if err := s.db.Where("article_id = ?", parsed).Order("created_at DESC, id DESC").Limit(20).Find(&rows).Error; err != nil {
		return nil, apperrors.New(apperrors.CodeDatabaseUnavailable)
	}
	items := make([]ArticleVersionItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, ArticleVersionItem{
			ID:        row.ID,
			ArticleID: row.ArticleID,
			Title:     row.Title,
			Summary:   row.Summary,
			ContentMD: row.ContentMD,
			Status:    row.Status,
			CreatedAt: row.CreatedAt,
		})
	}
	return items, nil
}

func (s *ArticleService) createVersion(db *gorm.DB, article model.Article, actorID uint64) error {
	if db == nil || article.ID == 0 {
		return nil
	}
	version := model.ArticleVersion{
		ArticleID: article.ID,
		Title:     article.Title,
		Summary:   article.Summary,
		ContentMD: article.ContentMD,
		Status:    article.Status,
		CreatedBy: actorID,
	}
	if err := db.Create(&version).Error; err != nil {
		return apperrors.New(apperrors.CodeDatabaseUnavailable)
	}
	return nil
}

// loadArticleTags 对标签 ID 去重并校验完整性，防止请求中的无效 ID 被静默忽略。
func loadArticleTags(db *gorm.DB, requestedIDs []uint64) ([]model.Tag, error) {
	ids := uniqueTagIDs(requestedIDs)
	if len(ids) == 0 {
		return []model.Tag{}, nil
	}

	var tags []model.Tag
	if err := db.Where("id IN ?", ids).Find(&tags).Error; err != nil {
		return nil, apperrors.New(apperrors.CodeDatabaseUnavailable)
	}
	if len(tags) != len(ids) {
		return nil, apperrors.New(apperrors.CodeResourceNotFound)
	}
	return tags, nil
}

func uniqueTagIDs(ids []uint64) []uint64 {
	unique := make([]uint64, 0, len(ids))
	seen := make(map[uint64]struct{}, len(ids))
	for _, id := range ids {
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	return unique
}

func normalizeArticleWriteError(err error) error {
	var appErr apperrors.AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return apperrors.New(apperrors.CodeDatabaseUnavailable)
}

func (s *ArticleService) list(query ArticleListQuery, onlyPublished bool) ([]ArticleItem, pagination.CursorPage, error) {
	limit := pagination.NormalizeLimit(query.Limit)
	if s.db != nil {
		return s.listFromDB(query, limit, onlyPublished)
	}
	return s.listInMemory(query, limit, onlyPublished)
}

func (s *ArticleService) listFromDB(query ArticleListQuery, limit int, onlyPublished bool) ([]ArticleItem, pagination.CursorPage, error) {
	dbQuery := s.db.Model(&model.Article{}).Preload("Topic").Preload("Tags")
	if onlyPublished {
		dbQuery = dbQuery.Where("status = ?", "published")
	} else if query.Status != "" {
		dbQuery = dbQuery.Where("status = ?", query.Status)
	}
	if query.Keyword != "" {
		dbQuery = dbQuery.Where("title LIKE ?", "%"+query.Keyword+"%")
	}
	if query.Topic != "" {
		dbQuery = dbQuery.Joins("JOIN topics ON topics.id = articles.topic_id").
			Where("topics.slug = ?", query.Topic)
	}
	if query.Tag != "" {
		dbQuery = dbQuery.Joins("JOIN article_tags ON article_tags.article_id = articles.id").
			Joins("JOIN tags ON tags.id = article_tags.tag_id").
			Where("tags.slug = ?", query.Tag)
	}
	if query.Cursor != "" {
		cursor, err := pagination.DecodeCursor(query.Cursor)
		if err != nil {
			return nil, pagination.CursorPage{}, apperrors.New(apperrors.CodeInvalidCursor)
		}
		dbQuery = dbQuery.Where(
			"articles.created_at < ? OR (articles.created_at = ? AND articles.id < ?)",
			cursor.CreatedAt,
			cursor.CreatedAt,
			cursor.ID,
		)
	}

	var rows []model.Article
	err := dbQuery.Order("articles.created_at DESC, articles.id DESC").
		Limit(limit + 1).
		Find(&rows).Error
	if err != nil {
		return nil, pagination.CursorPage{}, apperrors.New(apperrors.CodeDatabaseUnavailable)
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	items := make([]ArticleItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, itemFromModel(row))
	}

	nextCursor := ""
	if hasMore && len(rows) > 0 {
		last := rows[len(rows)-1]
		nextCursor, err = pagination.EncodeCursor(pagination.CursorPayload{
			CreatedAt: last.CreatedAt,
			ID:        last.ID,
		})
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

func (s *ArticleService) listInMemory(query ArticleListQuery, limit int, onlyPublished bool) ([]ArticleItem, pagination.CursorPage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]ArticleItem, 0, len(s.items))
	for _, detail := range s.items {
		item := detail.ArticleItem
		if onlyPublished && item.Status != "published" {
			continue
		}
		if !onlyPublished && query.Status != "" && item.Status != query.Status {
			continue
		}
		if query.Keyword != "" && !strings.Contains(strings.ToLower(item.Title), strings.ToLower(query.Keyword)) {
			continue
		}
		if query.Topic != "" && item.Topic.Slug != query.Topic {
			continue
		}
		if query.Tag != "" && !contains(item.Tags, query.Tag) {
			continue
		}
		items = append(items, item)
	}

	start := 0
	if query.Cursor != "" {
		cursor, err := pagination.DecodeCursor(query.Cursor)
		if err != nil {
			return nil, pagination.CursorPage{}, apperrors.New(apperrors.CodeInvalidCursor)
		}
		for index, item := range items {
			if item.CreatedAt.Equal(cursor.CreatedAt) && item.ID == cursor.ID {
				start = index + 1
				break
			}
		}
	}

	end := start + limit
	hasMore := end < len(items)
	if end > len(items) {
		end = len(items)
	}
	pageItems := items[start:end]
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

func (s *ArticleService) createInMemory(req ArticleCreateRequest, status string) (ArticleItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, item := range s.items {
		if item.Slug == req.Slug {
			return ArticleItem{}, apperrors.New(apperrors.CodeDuplicateSlug)
		}
	}

	now := time.Now().UTC()
	publishedAt := time.Time{}
	if status == "published" {
		publishedAt = now
	}
	item := ArticleDetail{
		ArticleItem: ArticleItem{
			ID:          uint64(len(s.items) + 1),
			Title:       req.Title,
			Slug:        req.Slug,
			Summary:     req.Summary,
			CoverURL:    req.CoverURL,
			Status:      status,
			TopicID:     inMemoryTopicID(req.TopicID),
			Topic:       inMemoryTopicReference(req.TopicID),
			Tags:        []string{},
			ViewCount:   0,
			PublishedAt: publishedAt,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		ContentMD: req.ContentMD,
		TagIDs:    uniqueTagIDs(req.TagIDs),
	}
	s.items = append(s.items, item)
	return item.ArticleItem, nil
}

func (s *ArticleService) updateInMemory(id uint64, req ArticleUpdateRequest, status string) (ArticleDetail, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, item := range s.items {
		if item.Slug == req.Slug && item.ID != id {
			return ArticleDetail{}, apperrors.New(apperrors.CodeDuplicateSlug)
		}
	}
	for index, item := range s.items {
		if item.ID == id {
			now := time.Now().UTC()
			item.Title = req.Title
			item.Slug = req.Slug
			item.Summary = req.Summary
			item.CoverURL = req.CoverURL
			item.ContentMD = req.ContentMD
			item.Status = status
			item.TopicID = inMemoryTopicID(req.TopicID)
			item.Topic = inMemoryTopicReference(req.TopicID)
			item.TagIDs = uniqueTagIDs(req.TagIDs)
			item.UpdatedAt = now
			if status == "published" && item.PublishedAt.IsZero() {
				item.PublishedAt = now
			}
			s.items[index] = item
			return item, nil
		}
	}
	return ArticleDetail{}, apperrors.New(apperrors.CodeResourceNotFound)
}

func (s *ArticleService) getCachedPublicList(query ArticleListQuery) (cachedArticleList, bool) {
	if s.redis == nil || s.db == nil {
		return cachedArticleList{}, false
	}
	payload, err := s.redis.Get(context.Background(), publicListCacheKey(query)).Bytes()
	if err != nil {
		return cachedArticleList{}, false
	}
	var cached cachedArticleList
	if err := json.Unmarshal(payload, &cached); err != nil {
		return cachedArticleList{}, false
	}
	return cached, true
}

func (s *ArticleService) setCachedPublicList(query ArticleListQuery, cached cachedArticleList) {
	if s.redis == nil || s.db == nil {
		return
	}
	payload, err := json.Marshal(cached)
	if err != nil {
		return
	}
	_ = s.redis.Set(context.Background(), publicListCacheKey(query), payload, articleCacheTTL).Err()
}

func (s *ArticleService) getCachedDetail(slug string) (ArticleDetail, bool) {
	if s.redis == nil || s.db == nil {
		return ArticleDetail{}, false
	}
	payload, err := s.redis.Get(context.Background(), cache.ArticleDetailKey(slug)).Bytes()
	if err != nil {
		return ArticleDetail{}, false
	}
	var item ArticleDetail
	if err := json.Unmarshal(payload, &item); err != nil {
		return ArticleDetail{}, false
	}
	return item, true
}

func (s *ArticleService) setCachedDetail(slug string, item ArticleDetail) {
	if s.redis == nil || s.db == nil {
		return
	}
	payload, err := json.Marshal(item)
	if err != nil {
		return
	}
	_ = s.redis.Set(context.Background(), cache.ArticleDetailKey(slug), payload, articleCacheTTL).Err()
}

func (s *ArticleService) invalidateArticleCaches(slugs ...string) {
	if s.redis == nil {
		return
	}
	ctx := context.Background()
	keys := make([]string, 0, len(slugs))
	for _, slug := range slugs {
		if slug == "" {
			continue
		}
		keys = append(keys, cache.ArticleDetailKey(slug))
	}
	iter := s.redis.Scan(ctx, 0, cache.ArticleListPattern(), 100).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if len(keys) > 0 {
		_ = s.redis.Del(ctx, keys...).Err()
	}
}

func publicListCacheKey(query ArticleListQuery) string {
	filterHash := cache.ArticleListFilterHash(query.Topic, query.Tag, query.Keyword, query.Status)
	return cache.ArticleListKey(query.Cursor, pagination.NormalizeLimit(query.Limit), filterHash)
}

func detailFromModel(article model.Article) ArticleDetail {
	return ArticleDetail{
		ArticleItem: itemFromModel(article),
		ContentMD:   article.ContentMD,
		TagIDs:      tagIDsFromModel(article.Tags),
	}
}

func itemFromModel(article model.Article) ArticleItem {
	publishedAt := time.Time{}
	if article.PublishedAt != nil {
		publishedAt = *article.PublishedAt
	}
	tags := make([]string, 0, len(article.Tags))
	for _, tag := range article.Tags {
		tags = append(tags, tag.Slug)
	}
	return ArticleItem{
		ID:       article.ID,
		Title:    article.Title,
		Slug:     article.Slug,
		Summary:  article.Summary,
		CoverURL: article.CoverURL,
		Status:   article.Status,
		TopicID:  article.TopicID,
		Topic: TopicReference{
			ID:    article.Topic.ID,
			Name:  article.Topic.Name,
			Label: article.Topic.Label,
			Slug:  article.Topic.Slug,
		},
		Tags:        tags,
		ViewCount:   article.ViewCount,
		PublishedAt: publishedAt,
		CreatedAt:   article.CreatedAt,
		UpdatedAt:   article.UpdatedAt,
	}
}

func inMemoryTopicID(topicID uint64) uint64 {
	if topicID == 0 {
		return 1
	}
	return topicID
}

func inMemoryTopicReference(topicID uint64) TopicReference {
	id := inMemoryTopicID(topicID)
	initialTopics := model.DefaultTopics()
	if id > 0 && id <= uint64(len(initialTopics)) {
		topic := initialTopics[id-1]
		return TopicReference{ID: id, Name: topic.Name, Label: topic.Label, Slug: topic.Slug}
	}
	return TopicReference{ID: id}
}

func normalizeArticleStatus(status string) (string, error) {
	if status == "" {
		return "draft", nil
	}
	switch status {
	case "draft", "published", "private", "archived":
		return status, nil
	default:
		return "", apperrors.New(apperrors.CodeInvalidParameter)
	}
}

func normalizeArticleCoverURL(value string) (string, error) {
	coverURL := strings.TrimSpace(value)
	if utf8.RuneCountInString(coverURL) > 500 {
		return "", apperrors.New(apperrors.CodeInvalidParameter)
	}
	return coverURL, nil
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func tagIDsFromModel(tags []model.Tag) []uint64 {
	ids := make([]uint64, 0, len(tags))
	for _, tag := range tags {
		ids = append(ids, tag.ID)
	}
	return ids
}
