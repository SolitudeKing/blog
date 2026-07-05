package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"solitude-blog/server/internal/config"
	"solitude-blog/server/internal/model"
)

type exportPackage struct {
	Version     int              `json:"version"`
	GeneratedAt string           `json:"generated_at"`
	Settings    exportedSetting  `json:"settings"`
	Categories  []exportCategory `json:"categories"`
	Tags        []exportTag      `json:"tags"`
	Articles    []exportArticle  `json:"articles"`
	Assets      []exportAsset    `json:"assets"`
	Notices     []exportNotice   `json:"notices"`
}

type exportedSetting struct {
	SiteName     string            `json:"site_name"`
	Author       string            `json:"author"`
	Essay        string            `json:"essay"`
	Theme        string            `json:"theme"`
	Mode         string            `json:"mode"`
	SocialLinks  map[string]string `json:"social_links"`
	AboutContent string            `json:"about_content"`
}

type exportCategory struct {
	LegacyID    string `json:"legacy_id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type exportTag struct {
	LegacyID    string `json:"legacy_id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Color       string `json:"color"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type exportArticle struct {
	LegacyID         string   `json:"legacy_id"`
	Title            string   `json:"title"`
	Slug             string   `json:"slug"`
	Summary          string   `json:"summary"`
	ContentMD        string   `json:"content_md"`
	Status           string   `json:"status"`
	CategoryLegacyID string   `json:"category_legacy_id"`
	TagLegacyIDs     []string `json:"tag_legacy_ids"`
	ViewCount        uint64   `json:"view_count"`
	PublishedAt      string   `json:"published_at"`
	CreatedAt        string   `json:"created_at"`
	UpdatedAt        string   `json:"updated_at"`
}

type exportAsset struct {
	LegacyID    string `json:"legacy_id"`
	DisplayName string `json:"display_name"`
	AltText     string `json:"alt_text"`
	StorageKey  string `json:"storage_key"`
	URL         string `json:"url"`
	ThumbURL    string `json:"thumb_url"`
	MimeType    string `json:"mime_type"`
	Ext         string `json:"ext"`
	Size        uint64 `json:"size"`
	SHA256      string `json:"sha256"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	FilePath    string `json:"file_path"`
}

type exportNotice struct {
	LegacyID  string `json:"legacy_id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Enabled   bool   `json:"enabled"`
	SortOrder int    `json:"sort_order"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type importReport struct {
	StartedAt  string            `json:"started_at"`
	FinishedAt string            `json:"finished_at"`
	DryRun     bool              `json:"dry_run"`
	Input      string            `json:"input"`
	Counts     map[string]int    `json:"counts"`
	Warnings   []string          `json:"warnings"`
	Failures   []string          `json:"failures"`
	Mappings   map[string]uint64 `json:"mappings,omitempty"`
}

func main() {
	input := flag.String("input", filepath.FromSlash("../migration-output/legacy-export/legacy-export.json"), "Legacy export JSON path.")
	storageRoot := flag.String("storage-root", "", "Storage root for imported assets. Defaults to STORAGE_LOCAL_ROOT.")
	reportPath := flag.String("report", filepath.FromSlash("../migration-output/legacy-import-report.json"), "Import report path.")
	dryRun := flag.Bool("dry-run", false, "Validate package and print planned changes without writing to MySQL.")
	flag.Parse()

	report := importReport{
		StartedAt: time.Now().UTC().Format(time.RFC3339),
		DryRun:    *dryRun,
		Input:     *input,
		Counts:    map[string]int{},
		Mappings:  map[string]uint64{},
	}

	pkg, err := loadPackage(*input)
	if err != nil {
		exitWithReport(report, *reportPath, fmt.Errorf("load package: %w", err))
	}
	report.Counts["categories_planned"] = len(pkg.Categories)
	report.Counts["tags_planned"] = len(pkg.Tags)
	report.Counts["articles_planned"] = len(pkg.Articles)
	report.Counts["assets_planned"] = len(pkg.Assets)
	report.Counts["notices_planned"] = len(pkg.Notices)

	cfg := config.Load()
	if *storageRoot == "" {
		*storageRoot = cfg.StorageLocalRoot
	}
	if cfg.MySQLDSN == "" && !*dryRun {
		exitWithReport(report, *reportPath, errors.New("MYSQL_DSN is required unless --dry-run is set"))
	}

	if *dryRun {
		report.FinishedAt = time.Now().UTC().Format(time.RFC3339)
		writeReport(*reportPath, report)
		printReport(report)
		return
	}

	db, err := gorm.Open(mysql.Open(cfg.MySQLDSN), &gorm.Config{})
	if err != nil {
		exitWithReport(report, *reportPath, fmt.Errorf("open mysql: %w", err))
	}
	if err := db.AutoMigrate(
		&model.User{},
		&model.Category{},
		&model.Tag{},
		&model.Article{},
		&model.Asset{},
		&model.SiteSetting{},
		&model.Notice{},
	); err != nil {
		exitWithReport(report, *reportPath, fmt.Errorf("auto migrate: %w", err))
	}

	if err := importPackage(db, pkg, filepath.Dir(*input), *storageRoot, &report); err != nil {
		exitWithReport(report, *reportPath, err)
	}

	report.FinishedAt = time.Now().UTC().Format(time.RFC3339)
	writeReport(*reportPath, report)
	printReport(report)
}

func loadPackage(path string) (exportPackage, error) {
	file, err := os.Open(path)
	if err != nil {
		return exportPackage{}, err
	}
	defer file.Close()

	var pkg exportPackage
	if err := json.NewDecoder(file).Decode(&pkg); err != nil {
		return exportPackage{}, err
	}
	if pkg.Version != 1 {
		return exportPackage{}, fmt.Errorf("unsupported export version %d", pkg.Version)
	}
	return pkg, nil
}

func importPackage(db *gorm.DB, pkg exportPackage, packageDir string, storageRoot string, report *importReport) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := upsertSetting(tx, pkg.Settings, report); err != nil {
			return err
		}
		categoryMap, err := upsertCategories(tx, pkg.Categories, report)
		if err != nil {
			return err
		}
		tagMap, err := upsertTags(tx, pkg.Tags, report)
		if err != nil {
			return err
		}
		if err := upsertAssets(tx, pkg.Assets, packageDir, storageRoot, report); err != nil {
			return err
		}
		if err := upsertArticles(tx, pkg.Articles, categoryMap, tagMap, report); err != nil {
			return err
		}
		if err := upsertNotices(tx, pkg.Notices, report); err != nil {
			return err
		}
		return nil
	})
}

func upsertSetting(tx *gorm.DB, setting exportedSetting, report *importReport) error {
	links, err := json.Marshal(setting.SocialLinks)
	if err != nil {
		return fmt.Errorf("marshal social links: %w", err)
	}
	if setting.SiteName == "" {
		setting.SiteName = "Solitude Blog"
	}
	if setting.Author == "" {
		setting.Author = "Solitude King"
	}
	if setting.Theme == "" {
		setting.Theme = "forest"
	}
	if setting.Mode == "" {
		setting.Mode = "light"
	}
	row := model.SiteSetting{
		ID:              1,
		SiteName:        setting.SiteName,
		Author:          setting.Author,
		Essay:           setting.Essay,
		Theme:           setting.Theme,
		Mode:            setting.Mode,
		SocialLinksJSON: string(links),
	}
	if err := tx.Save(&row).Error; err != nil {
		return fmt.Errorf("upsert setting: %w", err)
	}
	report.Counts["settings_imported"] = 1
	return nil
}

func upsertCategories(tx *gorm.DB, categories []exportCategory, report *importReport) (map[string]uint64, error) {
	result := map[string]uint64{}
	for _, item := range categories {
		if item.Name == "" || item.Slug == "" {
			report.Warnings = append(report.Warnings, "skip category with empty name or slug")
			continue
		}
		row := model.Category{}
		err := tx.Where("slug = ?", item.Slug).First(&row).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			row = model.Category{
				Name:        item.Name,
				Slug:        item.Slug,
				Description: item.Description,
				SortOrder:   item.SortOrder,
				CreatedAt:   parseImportTime(item.CreatedAt),
				UpdatedAt:   parseImportTime(item.UpdatedAt),
			}
			err = tx.Create(&row).Error
		} else if err == nil {
			row.Name = item.Name
			row.Description = item.Description
			row.SortOrder = item.SortOrder
			err = tx.Save(&row).Error
		}
		if err != nil {
			return nil, fmt.Errorf("upsert category %s: %w", item.Slug, err)
		}
		result[item.LegacyID] = row.ID
		report.Mappings["category:"+item.LegacyID] = row.ID
		report.Counts["categories_imported"]++
	}
	return result, nil
}

func upsertTags(tx *gorm.DB, tags []exportTag, report *importReport) (map[string]uint64, error) {
	result := map[string]uint64{}
	for _, item := range tags {
		if item.Name == "" || item.Slug == "" {
			report.Warnings = append(report.Warnings, "skip tag with empty name or slug")
			continue
		}
		row := model.Tag{}
		err := tx.Where("slug = ?", item.Slug).First(&row).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			row = model.Tag{
				Name:        item.Name,
				Slug:        item.Slug,
				Description: item.Description,
				Color:       item.Color,
				CreatedAt:   parseImportTime(item.CreatedAt),
				UpdatedAt:   parseImportTime(item.UpdatedAt),
			}
			err = tx.Create(&row).Error
		} else if err == nil {
			row.Name = item.Name
			row.Description = item.Description
			row.Color = item.Color
			err = tx.Save(&row).Error
		}
		if err != nil {
			return nil, fmt.Errorf("upsert tag %s: %w", item.Slug, err)
		}
		result[item.LegacyID] = row.ID
		report.Mappings["tag:"+item.LegacyID] = row.ID
		report.Counts["tags_imported"]++
	}
	return result, nil
}

func upsertAssets(tx *gorm.DB, assets []exportAsset, packageDir string, storageRoot string, report *importReport) error {
	for _, item := range assets {
		if item.StorageKey == "" || item.FilePath == "" {
			report.Warnings = append(report.Warnings, "skip asset with empty storage key or file path")
			continue
		}
		source := filepath.Join(packageDir, filepath.FromSlash(item.FilePath))
		target := filepath.Join(storageRoot, filepath.FromSlash(item.StorageKey))
		hash, size, err := copyAsset(source, target)
		if err != nil {
			return fmt.Errorf("copy asset %s: %w", item.StorageKey, err)
		}
		if item.SHA256 != "" && !strings.EqualFold(item.SHA256, hash) {
			report.Warnings = append(report.Warnings, fmt.Sprintf("asset checksum changed: %s", item.StorageKey))
		}
		row := model.Asset{}
		err = tx.Where("storage_key = ?", item.StorageKey).First(&row).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			row = model.Asset{
				DisplayName: item.DisplayName,
				AltText:     item.AltText,
				StorageKey:  item.StorageKey,
				URL:         item.URL,
				ThumbURL:    item.ThumbURL,
				MimeType:    item.MimeType,
				Ext:         item.Ext,
				Size:        size,
				SHA256:      hash,
				Status:      normalizeAssetStatus(item.Status),
				CreatedAt:   parseImportTime(item.CreatedAt),
				UpdatedAt:   parseImportTime(item.UpdatedAt),
			}
			err = tx.Create(&row).Error
		} else if err == nil {
			row.DisplayName = item.DisplayName
			row.AltText = item.AltText
			row.URL = item.URL
			row.ThumbURL = item.ThumbURL
			row.MimeType = item.MimeType
			row.Ext = item.Ext
			row.Size = size
			row.SHA256 = hash
			row.Status = normalizeAssetStatus(item.Status)
			err = tx.Save(&row).Error
		}
		if err != nil {
			return fmt.Errorf("upsert asset %s: %w", item.StorageKey, err)
		}
		report.Counts["assets_imported"]++
	}
	return nil
}

func upsertArticles(
	tx *gorm.DB,
	articles []exportArticle,
	categoryMap map[string]uint64,
	tagMap map[string]uint64,
	report *importReport,
) error {
	for _, item := range articles {
		if item.Title == "" || item.Slug == "" {
			report.Warnings = append(report.Warnings, "skip article with empty title or slug")
			continue
		}
		categoryID := categoryMap[item.CategoryLegacyID]
		if categoryID == 0 {
			categoryID = ensureFallbackCategory(tx, report)
		}
		tagIDs := make([]uint64, 0, len(item.TagLegacyIDs))
		for _, legacyID := range item.TagLegacyIDs {
			if id := tagMap[legacyID]; id > 0 {
				tagIDs = append(tagIDs, id)
			} else {
				report.Warnings = append(report.Warnings, fmt.Sprintf("article %s missing tag mapping %s", item.Slug, legacyID))
			}
		}

		article := model.Article{}
		err := tx.Where("slug = ?", item.Slug).First(&article).Error
		publishedAt := parseOptionalImportTime(item.PublishedAt)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			article = model.Article{
				Title:       item.Title,
				Slug:        item.Slug,
				Summary:     item.Summary,
				ContentMD:   item.ContentMD,
				Status:      normalizeArticleStatus(item.Status),
				CategoryID:  categoryID,
				AuthorID:    1,
				ViewCount:   item.ViewCount,
				PublishedAt: publishedAt,
				CreatedAt:   parseImportTime(item.CreatedAt),
				UpdatedAt:   parseImportTime(item.UpdatedAt),
			}
			err = tx.Create(&article).Error
		} else if err == nil {
			article.Title = item.Title
			article.Summary = item.Summary
			article.ContentMD = item.ContentMD
			article.Status = normalizeArticleStatus(item.Status)
			article.CategoryID = categoryID
			article.ViewCount = item.ViewCount
			article.PublishedAt = publishedAt
			err = tx.Save(&article).Error
		}
		if err != nil {
			return fmt.Errorf("upsert article %s: %w", item.Slug, err)
		}

		var tags []model.Tag
		if len(tagIDs) > 0 {
			if err := tx.Find(&tags, tagIDs).Error; err != nil {
				return fmt.Errorf("load tags for article %s: %w", item.Slug, err)
			}
		}
		if err := tx.Model(&article).Association("Tags").Replace(tags); err != nil {
			return fmt.Errorf("replace tags for article %s: %w", item.Slug, err)
		}
		report.Mappings["article:"+item.LegacyID] = article.ID
		report.Counts["articles_imported"]++
	}
	return nil
}

func upsertNotices(tx *gorm.DB, notices []exportNotice, report *importReport) error {
	for _, item := range notices {
		if item.Title == "" {
			report.Warnings = append(report.Warnings, "skip notice with empty title")
			continue
		}
		row := model.Notice{}
		err := tx.Where("title = ?", item.Title).First(&row).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			row = model.Notice{
				Title:     item.Title,
				Content:   item.Content,
				Enabled:   item.Enabled,
				SortOrder: item.SortOrder,
				CreatedAt: parseImportTime(item.CreatedAt),
				UpdatedAt: parseImportTime(item.UpdatedAt),
			}
			err = tx.Create(&row).Error
		} else if err == nil {
			row.Content = item.Content
			row.Enabled = item.Enabled
			row.SortOrder = item.SortOrder
			err = tx.Save(&row).Error
		}
		if err != nil {
			return fmt.Errorf("upsert notice %s: %w", item.Title, err)
		}
		report.Mappings["notice:"+item.LegacyID] = row.ID
		report.Counts["notices_imported"]++
	}
	return nil
}

func ensureFallbackCategory(tx *gorm.DB, report *importReport) uint64 {
	row := model.Category{}
	err := tx.Where("slug = ?", "legacy").First(&row).Error
	if err == nil {
		return row.ID
	}
	row = model.Category{Name: "Legacy", Slug: "legacy", SortOrder: 999}
	if err := tx.Create(&row).Error; err != nil {
		report.Warnings = append(report.Warnings, "fallback category create failed: "+err.Error())
		return 0
	}
	return row.ID
}

func copyAsset(source string, target string) (string, uint64, error) {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return "", 0, err
	}
	sourceFile, err := os.Open(source)
	if err != nil {
		return "", 0, err
	}
	defer sourceFile.Close()

	targetFile, err := os.Create(target)
	if err != nil {
		return "", 0, err
	}
	defer targetFile.Close()

	hash := sha256.New()
	size, err := io.Copy(io.MultiWriter(targetFile, hash), sourceFile)
	if err != nil {
		return "", 0, err
	}
	return hex.EncodeToString(hash.Sum(nil)), uint64(size), nil
}

func normalizeArticleStatus(status string) string {
	switch status {
	case "published", "private", "archived", "draft":
		return status
	default:
		return "draft"
	}
}

func normalizeAssetStatus(status string) string {
	if status == "temporary" {
		return "temporary"
	}
	return "active"
}

func parseImportTime(value string) time.Time {
	parsed := parseOptionalImportTime(value)
	if parsed == nil {
		return time.Now().UTC()
	}
	return *parsed
}

func parseOptionalImportTime(value string) *time.Time {
	if value == "" {
		return nil
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return &parsed
	}
	if parsed, err := time.Parse("2006-01-02 15:04:05", value); err == nil {
		return &parsed
	}
	return nil
}

func writeReport(path string, report importReport) {
	report.FinishedAt = time.Now().UTC().Format(time.RFC3339)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		slog.Warn("create report directory failed", "error", err)
		return
	}
	file, err := os.Create(path)
	if err != nil {
		slog.Warn("create report failed", "error", err)
		return
	}
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		slog.Warn("write report failed", "error", err)
	}
}

func printReport(report importReport) {
	payload, _ := json.MarshalIndent(report.Counts, "", "  ")
	fmt.Println(string(payload))
	if len(report.Warnings) > 0 {
		fmt.Printf("warnings: %d\n", len(report.Warnings))
	}
	if len(report.Failures) > 0 {
		fmt.Printf("failures: %d\n", len(report.Failures))
	}
}

func exitWithReport(report importReport, reportPath string, err error) {
	report.Failures = append(report.Failures, err.Error())
	writeReport(reportPath, report)
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
