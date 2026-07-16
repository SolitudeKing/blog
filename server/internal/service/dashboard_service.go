package service

import (
	"errors"
	"time"

	"gorm.io/gorm"

	apperrors "solitude-blog/server/internal/errors"
	"solitude-blog/server/internal/model"
)

type DashboardService struct {
	db *gorm.DB
}

type DashboardSummary struct {
	ArticleCounts  DashboardArticleCounts  `json:"article_counts"`
	TaxonomyCounts DashboardTaxonomyCounts `json:"taxonomy_counts"`
	NoticeCounts   DashboardNoticeCounts   `json:"notice_counts"`
	AssetCount     int64                   `json:"asset_count"`
	TotalViews     uint64                  `json:"total_views"`
	RecentArticles []DashboardArticleItem  `json:"recent_articles"`
	TopArticles    []DashboardTopArticle   `json:"top_articles"`
	TopicStats     []DashboardTopicStat    `json:"topic_stats"`
	ActiveNotice   *DashboardNoticeItem    `json:"active_notice"`
	GeneratedAt    time.Time               `json:"generated_at"`
}

type DashboardArticleCounts struct {
	Total     int64 `json:"total"`
	Published int64 `json:"published"`
	Draft     int64 `json:"draft"`
	Private   int64 `json:"private"`
	Archived  int64 `json:"archived"`
}

type DashboardTaxonomyCounts struct {
	Topics int64 `json:"topics"`
	Tags   int64 `json:"tags"`
}

type DashboardNoticeCounts struct {
	Total   int64 `json:"total"`
	Enabled int64 `json:"enabled"`
}

type DashboardArticleItem struct {
	ID        uint64    `json:"id"`
	Title     string    `json:"title"`
	Slug      string    `json:"slug"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DashboardTopArticle struct {
	ID        uint64    `json:"id"`
	Title     string    `json:"title"`
	Slug      string    `json:"slug"`
	ViewCount uint64    `json:"view_count"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DashboardTopicStat struct {
	ID           uint64 `json:"id"`
	Name         string `json:"name"`
	Label        string `json:"label"`
	Slug         string `json:"slug"`
	ArticleCount int64  `json:"article_count"`
}

type DashboardNoticeItem struct {
	ID        uint64     `json:"id"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	StartsAt  *time.Time `json:"starts_at"`
	EndsAt    *time.Time `json:"ends_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func NewDashboardService(db *gorm.DB) *DashboardService {
	return &DashboardService{db: db}
}

func (s *DashboardService) Summary() (DashboardSummary, error) {
	if s.db == nil {
		return DashboardSummary{
			RecentArticles: []DashboardArticleItem{},
			TopArticles:    []DashboardTopArticle{},
			TopicStats:     []DashboardTopicStat{},
			GeneratedAt:    time.Now().UTC(),
		}, nil
	}

	summary := DashboardSummary{GeneratedAt: time.Now().UTC()}
	if err := s.fillArticleCounts(&summary); err != nil {
		return DashboardSummary{}, err
	}
	if err := s.fillTaxonomyCounts(&summary); err != nil {
		return DashboardSummary{}, err
	}
	if err := s.fillNoticeCounts(&summary); err != nil {
		return DashboardSummary{}, err
	}
	if err := s.fillAssetCount(&summary); err != nil {
		return DashboardSummary{}, err
	}
	if err := s.fillRecentArticles(&summary); err != nil {
		return DashboardSummary{}, err
	}
	if err := s.fillTopArticles(&summary); err != nil {
		return DashboardSummary{}, err
	}
	if err := s.fillTopicStats(&summary); err != nil {
		return DashboardSummary{}, err
	}
	if err := s.fillActiveNotice(&summary); err != nil {
		return DashboardSummary{}, err
	}
	return summary, nil
}

func (s *DashboardService) fillArticleCounts(summary *DashboardSummary) error {
	if err := s.db.Model(&model.Article{}).Count(&summary.ArticleCounts.Total).Error; err != nil {
		return apperrors.New(apperrors.CodeDatabaseUnavailable)
	}
	if err := s.db.Model(&model.Article{}).Where("status = ?", "published").Count(&summary.ArticleCounts.Published).Error; err != nil {
		return apperrors.New(apperrors.CodeDatabaseUnavailable)
	}
	if err := s.db.Model(&model.Article{}).Where("status = ?", "draft").Count(&summary.ArticleCounts.Draft).Error; err != nil {
		return apperrors.New(apperrors.CodeDatabaseUnavailable)
	}
	if err := s.db.Model(&model.Article{}).Where("status = ?", "private").Count(&summary.ArticleCounts.Private).Error; err != nil {
		return apperrors.New(apperrors.CodeDatabaseUnavailable)
	}
	if err := s.db.Model(&model.Article{}).Where("status = ?", "archived").Count(&summary.ArticleCounts.Archived).Error; err != nil {
		return apperrors.New(apperrors.CodeDatabaseUnavailable)
	}
	var totalViews uint64
	if err := s.db.Model(&model.Article{}).Select("COALESCE(SUM(view_count), 0)").Scan(&totalViews).Error; err != nil {
		return apperrors.New(apperrors.CodeDatabaseUnavailable)
	}
	summary.TotalViews = totalViews
	return nil
}

func (s *DashboardService) fillTaxonomyCounts(summary *DashboardSummary) error {
	if err := s.db.Model(&model.Topic{}).Count(&summary.TaxonomyCounts.Topics).Error; err != nil {
		return apperrors.New(apperrors.CodeDatabaseUnavailable)
	}
	if err := s.db.Model(&model.Tag{}).Count(&summary.TaxonomyCounts.Tags).Error; err != nil {
		return apperrors.New(apperrors.CodeDatabaseUnavailable)
	}
	return nil
}

func (s *DashboardService) fillNoticeCounts(summary *DashboardSummary) error {
	if err := s.db.Model(&model.Notice{}).Count(&summary.NoticeCounts.Total).Error; err != nil {
		return apperrors.New(apperrors.CodeDatabaseUnavailable)
	}
	if err := s.db.Model(&model.Notice{}).Where("enabled = ?", true).Count(&summary.NoticeCounts.Enabled).Error; err != nil {
		return apperrors.New(apperrors.CodeDatabaseUnavailable)
	}
	return nil
}

func (s *DashboardService) fillAssetCount(summary *DashboardSummary) error {
	if err := s.db.Model(&model.Asset{}).Count(&summary.AssetCount).Error; err != nil {
		return apperrors.New(apperrors.CodeDatabaseUnavailable)
	}
	return nil
}

func (s *DashboardService) fillRecentArticles(summary *DashboardSummary) error {
	var rows []model.Article
	if err := s.db.Select("id", "title", "slug", "status", "updated_at").Order("updated_at DESC, id DESC").Limit(5).Find(&rows).Error; err != nil {
		return apperrors.New(apperrors.CodeDatabaseUnavailable)
	}
	summary.RecentArticles = make([]DashboardArticleItem, 0, len(rows))
	for _, row := range rows {
		summary.RecentArticles = append(summary.RecentArticles, DashboardArticleItem{
			ID:        row.ID,
			Title:     row.Title,
			Slug:      row.Slug,
			Status:    row.Status,
			UpdatedAt: row.UpdatedAt,
		})
	}
	return nil
}

func (s *DashboardService) fillTopArticles(summary *DashboardSummary) error {
	var rows []model.Article
	if err := s.db.Select("id", "title", "slug", "view_count", "updated_at").
		Where("status = ?", "published").
		Order("view_count DESC, updated_at DESC, id DESC").
		Limit(5).
		Find(&rows).Error; err != nil {
		return apperrors.New(apperrors.CodeDatabaseUnavailable)
	}
	summary.TopArticles = make([]DashboardTopArticle, 0, len(rows))
	for _, row := range rows {
		summary.TopArticles = append(summary.TopArticles, DashboardTopArticle{
			ID:        row.ID,
			Title:     row.Title,
			Slug:      row.Slug,
			ViewCount: row.ViewCount,
			UpdatedAt: row.UpdatedAt,
		})
	}
	return nil
}

func (s *DashboardService) fillTopicStats(summary *DashboardSummary) error {
	type topicStatRow struct {
		ID           uint64
		Name         string
		Label        string
		Slug         string
		ArticleCount int64
	}
	var rows []topicStatRow
	if err := s.db.Table("topics").
		Select("topics.id, topics.name, topics.label, topics.slug, COUNT(articles.id) AS article_count").
		Joins("LEFT JOIN articles ON articles.topic_id = topics.id AND articles.deleted_at IS NULL").
		Where("topics.deleted_at IS NULL").
		Group("topics.id, topics.name, topics.label, topics.slug").
		Order("article_count DESC, topics.sort_order ASC, topics.id ASC").
		Limit(8).
		Scan(&rows).Error; err != nil {
		return apperrors.New(apperrors.CodeDatabaseUnavailable)
	}
	summary.TopicStats = make([]DashboardTopicStat, 0, len(rows))
	for _, row := range rows {
		summary.TopicStats = append(summary.TopicStats, DashboardTopicStat{
			ID:           row.ID,
			Name:         row.Name,
			Label:        row.Label,
			Slug:         row.Slug,
			ArticleCount: row.ArticleCount,
		})
	}
	return nil
}

func (s *DashboardService) fillActiveNotice(summary *DashboardSummary) error {
	now := time.Now().UTC()
	var row model.Notice
	err := s.db.
		Where("enabled = ?", true).
		Where("(starts_at IS NULL OR starts_at <= ?) AND (ends_at IS NULL OR ends_at > ?)", now, now).
		Order("sort_order ASC, created_at DESC, id DESC").
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return apperrors.New(apperrors.CodeDatabaseUnavailable)
	}
	summary.ActiveNotice = &DashboardNoticeItem{
		ID:        row.ID,
		Title:     row.Title,
		Content:   row.Content,
		StartsAt:  row.StartsAt,
		EndsAt:    row.EndsAt,
		UpdatedAt: row.UpdatedAt,
	}
	return nil
}
