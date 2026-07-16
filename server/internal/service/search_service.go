package service

import (
	"context"
	"strings"
	"unicode/utf8"

	"gorm.io/gorm"

	apperrors "solitude-blog/server/internal/errors"
	"solitude-blog/server/internal/model"
	"solitude-blog/server/internal/pagination"
)

type SearchService struct {
	db *gorm.DB
}

type ArticleSearchQuery struct {
	Cursor  string
	Limit   int
	Keyword string
}

type ArticleSearchItem struct {
	ArticleItem
	Snippet       string   `json:"snippet"`
	MatchedFields []string `json:"matched_fields"`
}

func NewSearchService(db *gorm.DB) *SearchService {
	return &SearchService{db: db}
}

func (s *SearchService) Article(ctx context.Context, query ArticleSearchQuery) ([]ArticleSearchItem, pagination.CursorPage, error) {
	limit := pagination.NormalizeLimit(query.Limit)
	keyword := strings.TrimSpace(query.Keyword)
	if keyword == "" {
		return []ArticleSearchItem{}, pagination.CursorPage{Cursor: query.Cursor, Limit: limit}, nil
	}
	if s.db == nil {
		return []ArticleSearchItem{}, pagination.CursorPage{Cursor: query.Cursor, Limit: limit}, nil
	}

	dbQuery := s.db.WithContext(ctx).Model(&model.Article{}).Preload("Topic").Preload("Tags").
		Where("articles.status = ?", "published")
	like := "%" + keyword + "%"
	dbQuery = dbQuery.Where(
		`articles.title LIKE ? OR articles.summary LIKE ? OR articles.content_md LIKE ?
		OR EXISTS (
			SELECT 1 FROM topics
			WHERE topics.id = articles.topic_id
			AND (topics.name LIKE ? OR topics.label LIKE ? OR topics.slug LIKE ?)
		)
		OR EXISTS (
			SELECT 1 FROM article_tags
			JOIN tags ON tags.id = article_tags.tag_id
			WHERE article_tags.article_id = articles.id
			AND (tags.name LIKE ? OR tags.slug LIKE ?)
		)`,
		like,
		like,
		like,
		like,
		like,
		like,
		like,
		like,
	)
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
	err := dbQuery.Order("articles.created_at DESC, articles.id DESC").Limit(limit + 1).Find(&rows).Error
	if err != nil {
		return nil, pagination.CursorPage{}, apperrors.New(apperrors.CodeDatabaseUnavailable)
	}

	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	items := make([]ArticleSearchItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, ArticleSearchItem{
			ArticleItem:   itemFromModel(row),
			Snippet:       searchSnippet(row, keyword),
			MatchedFields: matchedArticleFields(row, keyword),
		})
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

func searchSnippet(article model.Article, keyword string) string {
	candidates := []string{article.Summary, article.ContentMD, article.Title}
	for _, candidate := range candidates {
		if strings.Contains(strings.ToLower(candidate), strings.ToLower(keyword)) {
			return compactSnippet(candidate, keyword, 120)
		}
	}
	if article.Summary != "" {
		return compactSnippet(article.Summary, "", 120)
	}
	return compactSnippet(article.ContentMD, "", 120)
}

func matchedArticleFields(article model.Article, keyword string) []string {
	lowerKeyword := strings.ToLower(keyword)
	fields := []string{}
	if containsFold(article.Title, lowerKeyword) {
		fields = append(fields, "title")
	}
	if containsFold(article.Summary, lowerKeyword) {
		fields = append(fields, "summary")
	}
	if containsFold(article.ContentMD, lowerKeyword) {
		fields = append(fields, "content")
	}
	if containsFold(article.Topic.Name, lowerKeyword) || containsFold(article.Topic.Label, lowerKeyword) || containsFold(article.Topic.Slug, lowerKeyword) {
		fields = append(fields, "topic")
	}
	for _, tag := range article.Tags {
		if containsFold(tag.Name, lowerKeyword) || containsFold(tag.Slug, lowerKeyword) {
			fields = append(fields, "tag")
			break
		}
	}
	return fields
}

func compactSnippet(value string, keyword string, maxRunes int) string {
	plain := stripMarkdown(value)
	runes := []rune(plain)
	if len(runes) <= maxRunes {
		return strings.TrimSpace(plain)
	}
	start := 0
	if keyword != "" {
		index := strings.Index(strings.ToLower(plain), strings.ToLower(keyword))
		if index > 0 {
			prefixRunes := utf8.RuneCountInString(plain[:index])
			start = prefixRunes - maxRunes/3
			if start < 0 {
				start = 0
			}
		}
	}
	end := start + maxRunes
	if end > len(runes) {
		end = len(runes)
		start = end - maxRunes
		if start < 0 {
			start = 0
		}
	}
	snippet := strings.TrimSpace(string(runes[start:end]))
	if start > 0 {
		snippet = "..." + snippet
	}
	if end < len(runes) {
		snippet += "..."
	}
	return snippet
}

func stripMarkdown(value string) string {
	replacer := strings.NewReplacer("#", " ", "`", " ", "*", " ", "_", " ", ">", " ", "[", " ", "]", " ", "(", " ", ")", " ", "!", " ")
	return strings.Join(strings.Fields(replacer.Replace(value)), " ")
}

func containsFold(value string, lowerKeyword string) bool {
	return strings.Contains(strings.ToLower(value), lowerKeyword)
}
