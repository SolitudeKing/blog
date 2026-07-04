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

type ArticleService struct {
	db    *gorm.DB
	mu    sync.RWMutex
	items []ArticleDetail
}

type ArticleListQuery struct {
	Cursor   string
	Limit    int
	Category string
	Tag      string
	Keyword  string
	Status   string
}

type ArticleItem struct {
	ID          uint64    `json:"id"`
	Title       string    `json:"title"`
	Slug        string    `json:"slug"`
	Summary     string    `json:"summary"`
	Status      string    `json:"status"`
	Category    string    `json:"category"`
	Tags        []string  `json:"tags"`
	ViewCount   uint64    `json:"view_count"`
	PublishedAt time.Time `json:"published_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ArticleDetail struct {
	ArticleItem
	ContentMD  string   `json:"content_md"`
	CategoryID uint64   `json:"category_id"`
	TagIDs     []uint64 `json:"tag_ids"`
}

type ArticleCreateRequest struct {
	Title      string   `json:"title"`
	Slug       string   `json:"slug"`
	Summary    string   `json:"summary"`
	ContentMD  string   `json:"content_md"`
	CategoryID uint64   `json:"category_id"`
	TagIDs     []uint64 `json:"tag_ids"`
	Status     string   `json:"status"`
}

type ArticleUpdateRequest = ArticleCreateRequest

func NewArticleService(db *gorm.DB) *ArticleService {
	now := time.Now().UTC()
	return &ArticleService{
		db: db,
		items: []ArticleDetail{
			{
				ArticleItem: ArticleItem{
					ID:          1,
					Title:       "Welcome to Solitude Blog",
					Slug:        "welcome",
					Summary:     "The first article from the new blog scaffold.",
					Status:      "published",
					Category:    "notes",
					Tags:        []string{"go", "vue"},
					ViewCount:   0,
					PublishedAt: now,
					CreatedAt:   now,
					UpdatedAt:   now,
				},
				ContentMD:  "# Welcome\n\nThis is the new blog API scaffold.",
				CategoryID: 1,
				TagIDs:     []uint64{1, 2},
			},
		},
	}
}

func (s *ArticleService) PublicList(query ArticleListQuery) ([]ArticleItem, pagination.CursorPage, error) {
	return s.list(query, true)
}

func (s *ArticleService) ManageList(query ArticleListQuery) ([]ArticleItem, pagination.CursorPage, error) {
	return s.list(query, false)
}

func (s *ArticleService) Detail(slug string) (ArticleDetail, error) {
	if s.db != nil {
		var article model.Article
		err := s.db.Preload("Category").Preload("Tags").
			Where("slug = ? AND status = ?", slug, "published").
			First(&article).Error
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
		err := s.db.Preload("Category").Preload("Tags").First(&article, parsed).Error
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

func (s *ArticleService) Create(req ArticleCreateRequest) (ArticleItem, error) {
	if req.Title == "" || req.Slug == "" {
		return ArticleItem{}, apperrors.New(apperrors.CodeMissingRequiredField)
	}
	status, err := normalizeArticleStatus(req.Status)
	if err != nil {
		return ArticleItem{}, err
	}

	if s.db != nil {
		categoryID, err := s.resolveCategoryID(req.CategoryID)
		if err != nil {
			return ArticleItem{}, err
		}
		var count int64
		if err := s.db.Model(&model.Article{}).Where("slug = ?", req.Slug).Count(&count).Error; err != nil {
			return ArticleItem{}, apperrors.New(apperrors.CodeDatabaseUnavailable)
		}
		if count > 0 {
			return ArticleItem{}, apperrors.New(apperrors.CodeDuplicateSlug)
		}

		article := model.Article{
			Title:      req.Title,
			Slug:       req.Slug,
			Summary:    req.Summary,
			ContentMD:  req.ContentMD,
			Status:     status,
			CategoryID: categoryID,
			AuthorID:   1,
		}
		if status == "published" {
			now := time.Now().UTC()
			article.PublishedAt = &now
		}
		if err := s.db.Create(&article).Error; err != nil {
			return ArticleItem{}, apperrors.New(apperrors.CodeDatabaseUnavailable)
		}
		if len(req.TagIDs) > 0 {
			var tags []model.Tag
			if err := s.db.Find(&tags, req.TagIDs).Error; err != nil {
				return ArticleItem{}, apperrors.New(apperrors.CodeDatabaseUnavailable)
			}
			if err := s.db.Model(&article).Association("Tags").Replace(tags); err != nil {
				return ArticleItem{}, apperrors.New(apperrors.CodeDatabaseUnavailable)
			}
		}
		created, err := s.Info(strconv.FormatUint(article.ID, 10))
		if err != nil {
			return ArticleItem{}, err
		}
		return created.ArticleItem, nil
	}

	return s.createInMemory(req, status)
}

func (s *ArticleService) Update(id string, req ArticleUpdateRequest) (ArticleDetail, error) {
	parsed, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return ArticleDetail{}, apperrors.New(apperrors.CodeInvalidParameter)
	}
	if req.Title == "" || req.Slug == "" {
		return ArticleDetail{}, apperrors.New(apperrors.CodeMissingRequiredField)
	}
	status, err := normalizeArticleStatus(req.Status)
	if err != nil {
		return ArticleDetail{}, err
	}

	if s.db != nil {
		categoryID, err := s.resolveCategoryID(req.CategoryID)
		if err != nil {
			return ArticleDetail{}, err
		}
		var article model.Article
		err = s.db.Preload("Tags").First(&article, parsed).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ArticleDetail{}, apperrors.New(apperrors.CodeResourceNotFound)
		}
		if err != nil {
			return ArticleDetail{}, apperrors.New(apperrors.CodeDatabaseUnavailable)
		}

		var slugCount int64
		if err := s.db.Model(&model.Article{}).
			Where("slug = ? AND id <> ?", req.Slug, parsed).
			Count(&slugCount).Error; err != nil {
			return ArticleDetail{}, apperrors.New(apperrors.CodeDatabaseUnavailable)
		}
		if slugCount > 0 {
			return ArticleDetail{}, apperrors.New(apperrors.CodeDuplicateSlug)
		}

		article.Title = req.Title
		article.Slug = req.Slug
		article.Summary = req.Summary
		article.ContentMD = req.ContentMD
		article.Status = status
		article.CategoryID = categoryID
		if status == "published" && article.PublishedAt == nil {
			now := time.Now().UTC()
			article.PublishedAt = &now
		}

		if err := s.db.Save(&article).Error; err != nil {
			return ArticleDetail{}, apperrors.New(apperrors.CodeDatabaseUnavailable)
		}
		var tags []model.Tag
		if len(req.TagIDs) > 0 {
			if err := s.db.Find(&tags, req.TagIDs).Error; err != nil {
				return ArticleDetail{}, apperrors.New(apperrors.CodeDatabaseUnavailable)
			}
		}
		if err := s.db.Model(&article).Association("Tags").Replace(tags); err != nil {
			return ArticleDetail{}, apperrors.New(apperrors.CodeDatabaseUnavailable)
		}
		return s.Info(id)
	}

	return s.updateInMemory(parsed, req, status)
}

func (s *ArticleService) resolveCategoryID(categoryID uint64) (uint64, error) {
	if categoryID > 0 {
		return categoryID, nil
	}
	var category model.Category
	err := s.db.Where("slug = ?", "notes").First(&category).Error
	if err == nil {
		return category.ID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, apperrors.New(apperrors.CodeDatabaseUnavailable)
	}
	category = model.Category{Name: "Notes", Slug: "notes", SortOrder: 1}
	if err := s.db.Create(&category).Error; err != nil {
		return 0, apperrors.New(apperrors.CodeDatabaseUnavailable)
	}
	return category.ID, nil
}

func (s *ArticleService) Delete(id string) error {
	parsed, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return apperrors.New(apperrors.CodeInvalidParameter)
	}

	if s.db != nil {
		var article model.Article
		err := s.db.First(&article, parsed).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.New(apperrors.CodeResourceNotFound)
		}
		if err != nil {
			return apperrors.New(apperrors.CodeDatabaseUnavailable)
		}
		if err := s.db.Model(&article).Association("Tags").Clear(); err != nil {
			return apperrors.New(apperrors.CodeDatabaseUnavailable)
		}
		result := s.db.Delete(&article)
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

func (s *ArticleService) list(query ArticleListQuery, onlyPublished bool) ([]ArticleItem, pagination.CursorPage, error) {
	limit := pagination.NormalizeLimit(query.Limit)
	if s.db != nil {
		return s.listFromDB(query, limit, onlyPublished)
	}
	return s.listInMemory(query, limit, onlyPublished)
}

func (s *ArticleService) listFromDB(query ArticleListQuery, limit int, onlyPublished bool) ([]ArticleItem, pagination.CursorPage, error) {
	dbQuery := s.db.Model(&model.Article{}).Preload("Category").Preload("Tags")
	if onlyPublished {
		dbQuery = dbQuery.Where("status = ?", "published")
	} else if query.Status != "" {
		dbQuery = dbQuery.Where("status = ?", query.Status)
	}
	if query.Keyword != "" {
		dbQuery = dbQuery.Where("title LIKE ?", "%"+query.Keyword+"%")
	}
	if query.Category != "" {
		dbQuery = dbQuery.Joins("JOIN categories ON categories.id = articles.category_id").
			Where("categories.slug = ?", query.Category)
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
		if query.Category != "" && item.Category != query.Category {
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
			Status:      status,
			Category:    "uncategorized",
			Tags:        []string{},
			ViewCount:   0,
			PublishedAt: publishedAt,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		ContentMD: req.ContentMD,
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
			item.ContentMD = req.ContentMD
			item.Status = status
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

func detailFromModel(article model.Article) ArticleDetail {
	return ArticleDetail{
		ArticleItem: itemFromModel(article),
		ContentMD:   article.ContentMD,
		CategoryID:  article.CategoryID,
		TagIDs:      tagIDsFromModel(article.Tags),
	}
}

func itemFromModel(article model.Article) ArticleItem {
	publishedAt := time.Time{}
	if article.PublishedAt != nil {
		publishedAt = *article.PublishedAt
	}
	category := "uncategorized"
	if article.Category.Slug != "" {
		category = article.Category.Slug
	}
	tags := make([]string, 0, len(article.Tags))
	for _, tag := range article.Tags {
		tags = append(tags, tag.Slug)
	}
	return ArticleItem{
		ID:          article.ID,
		Title:       article.Title,
		Slug:        article.Slug,
		Summary:     article.Summary,
		Status:      article.Status,
		Category:    category,
		Tags:        tags,
		ViewCount:   article.ViewCount,
		PublishedAt: publishedAt,
		CreatedAt:   article.CreatedAt,
		UpdatedAt:   article.UpdatedAt,
	}
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
