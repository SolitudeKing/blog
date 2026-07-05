package service

import (
	"context"
	"encoding/xml"
	"net/url"
	"strings"
	"time"

	"gorm.io/gorm"

	apperrors "solitude-blog/server/internal/errors"
	"solitude-blog/server/internal/model"
)

const feedArticleLimit = 50

type FeedService struct {
	db *gorm.DB
}

type RSSFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Atom    string     `xml:"xmlns:atom,attr"`
	Channel RSSChannel `xml:"channel"`
}

type RSSChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Language    string    `xml:"language"`
	LastBuild   string    `xml:"lastBuildDate"`
	AtomLink    AtomLink  `xml:"atom:link"`
	Items       []RSSItem `xml:"item"`
}

type AtomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	GUID        string `xml:"guid"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

type URLSet struct {
	XMLName xml.Name     `xml:"urlset"`
	XMLNS   string       `xml:"xmlns,attr"`
	URLs    []SitemapURL `xml:"url"`
}

type SitemapURL struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod,omitempty"`
	ChangeFreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
}

func NewFeedService(db *gorm.DB) *FeedService {
	return &FeedService{db: db}
}

func (s *FeedService) RSS(ctx context.Context, baseURL string) ([]byte, error) {
	setting, articles, err := s.loadFeedData(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	items := make([]RSSItem, 0, len(articles))
	for _, article := range articles {
		publishedAt := article.CreatedAt
		if article.PublishedAt != nil {
			publishedAt = *article.PublishedAt
		}
		items = append(items, RSSItem{
			Title:       article.Title,
			Link:        joinURL(baseURL, "/articles/"+article.Slug),
			GUID:        joinURL(baseURL, "/articles/"+article.Slug),
			Description: article.Summary,
			PubDate:     publishedAt.UTC().Format(time.RFC1123Z),
		})
		if article.UpdatedAt.After(now) {
			continue
		}
		if article.UpdatedAt.After(setting.UpdatedAt) {
			setting.UpdatedAt = article.UpdatedAt
		}
	}

	feed := RSSFeed{
		Version: "2.0",
		Atom:    "http://www.w3.org/2005/Atom",
		Channel: RSSChannel{
			Title:       setting.SiteName,
			Link:        baseURL,
			Description: setting.Essay,
			Language:    "zh-CN",
			LastBuild:   setting.UpdatedAt.UTC().Format(time.RFC1123Z),
			AtomLink: AtomLink{
				Href: joinURL(baseURL, "/rss.xml"),
				Rel:  "self",
				Type: "application/rss+xml",
			},
			Items: items,
		},
	}
	return marshalXML(feed)
}

func (s *FeedService) Sitemap(ctx context.Context, baseURL string) ([]byte, error) {
	setting, articles, err := s.loadFeedData(ctx)
	if err != nil {
		return nil, err
	}

	urls := []SitemapURL{
		{Loc: baseURL, LastMod: formatSitemapTime(setting.UpdatedAt), ChangeFreq: "daily", Priority: "1.0"},
		{Loc: joinURL(baseURL, "/archives"), LastMod: formatSitemapTime(setting.UpdatedAt), ChangeFreq: "weekly", Priority: "0.7"},
	}
	for _, article := range articles {
		lastMod := article.UpdatedAt
		if article.PublishedAt != nil && article.PublishedAt.After(lastMod) {
			lastMod = *article.PublishedAt
		}
		urls = append(urls, SitemapURL{
			Loc:        joinURL(baseURL, "/articles/"+article.Slug),
			LastMod:    formatSitemapTime(lastMod),
			ChangeFreq: "monthly",
			Priority:   "0.8",
		})
	}

	return marshalXML(URLSet{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	})
}

func (s *FeedService) loadFeedData(ctx context.Context) (model.SiteSetting, []model.Article, error) {
	setting := model.SiteSetting{
		SiteName:  "Solitude Blog",
		Author:    "Solitude King",
		Essay:     "Keep writing, keep shipping.",
		UpdatedAt: time.Now().UTC(),
	}
	if s.db == nil {
		return setting, []model.Article{}, nil
	}

	_ = s.db.WithContext(ctx).First(&setting, 1).Error

	var articles []model.Article
	err := s.db.WithContext(ctx).
		Where("status = ?", "published").
		Order("published_at DESC, created_at DESC, id DESC").
		Limit(feedArticleLimit).
		Find(&articles).Error
	if err != nil {
		return model.SiteSetting{}, nil, apperrors.New(apperrors.CodeDatabaseUnavailable)
	}
	return setting, articles, nil
}

func marshalXML(value any) ([]byte, error) {
	payload, err := xml.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, apperrors.New(apperrors.CodeInternalServerError)
	}
	return append([]byte(xml.Header), payload...), nil
}

func joinURL(baseURL string, path string) string {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return strings.TrimRight(baseURL, "/") + path
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/") + path
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String()
}

func formatSitemapTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format("2006-01-02")
}
