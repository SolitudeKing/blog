package service

import (
	"strconv"
	"strings"
	"time"

	apperrors "solitude-blog/server/internal/errors"
	"solitude-blog/server/internal/pagination"
)

type ArticleService struct {
	items []ArticleItem
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
	ContentMD string `json:"content_md"`
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

func NewArticleService() *ArticleService {
	now := time.Now().UTC()
	return &ArticleService{
		items: []ArticleItem{
			{
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
		},
	}
}

func (s *ArticleService) PublicList(query ArticleListQuery) ([]ArticleItem, pagination.CursorPage) {
	return s.list(query, true)
}

func (s *ArticleService) ManageList(query ArticleListQuery) ([]ArticleItem, pagination.CursorPage) {
	return s.list(query, false)
}

func (s *ArticleService) Detail(slug string) (ArticleDetail, error) {
	for _, item := range s.items {
		if item.Slug == slug && item.Status == "published" {
			return ArticleDetail{
				ArticleItem: item,
				ContentMD:   "# Welcome\n\nThis is the new blog API scaffold.",
			}, nil
		}
	}
	return ArticleDetail{}, apperrors.New(apperrors.CodeResourceNotFound)
}

func (s *ArticleService) Info(id string) (ArticleDetail, error) {
	parsed, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return ArticleDetail{}, apperrors.New(apperrors.CodeInvalidParameter)
	}
	for _, item := range s.items {
		if item.ID == parsed {
			return ArticleDetail{
				ArticleItem: item,
				ContentMD:   "# Draft\n\nEditable content placeholder.",
			}, nil
		}
	}
	return ArticleDetail{}, apperrors.New(apperrors.CodeResourceNotFound)
}

func (s *ArticleService) Create(req ArticleCreateRequest) (ArticleItem, error) {
	if req.Title == "" || req.Slug == "" {
		return ArticleItem{}, apperrors.New(apperrors.CodeMissingRequiredField)
	}
	now := time.Now().UTC()
	item := ArticleItem{
		ID:          uint64(len(s.items) + 1),
		Title:       req.Title,
		Slug:        req.Slug,
		Summary:     req.Summary,
		Status:      statusOrDraft(req.Status),
		Category:    "uncategorized",
		Tags:        []string{},
		PublishedAt: now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	s.items = append(s.items, item)
	return item, nil
}

func (s *ArticleService) list(query ArticleListQuery, onlyPublished bool) ([]ArticleItem, pagination.CursorPage) {
	limit := pagination.NormalizeLimit(query.Limit)
	items := make([]ArticleItem, 0, len(s.items))
	for _, item := range s.items {
		if onlyPublished && item.Status != "published" {
			continue
		}
		if query.Keyword != "" && !strings.Contains(strings.ToLower(item.Title), strings.ToLower(query.Keyword)) {
			continue
		}
		items = append(items, item)
	}
	if len(items) > limit {
		items = items[:limit]
	}
	return items, pagination.CursorPage{
		Cursor:     query.Cursor,
		NextCursor: "",
		Limit:      limit,
		HasMore:    false,
	}
}

func statusOrDraft(status string) string {
	if status == "" {
		return "draft"
	}
	return status
}
