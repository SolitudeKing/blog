package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"reflect"
	"strings"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"solitude-blog/server/internal/appearance"
	"solitude-blog/server/internal/config"
	"solitude-blog/server/internal/model"
)

type Resources struct {
	DB    *gorm.DB
	Redis *redis.Client
}

func Open(ctx context.Context, cfg config.Config) (Resources, error) {
	var resources Resources

	db, err := openMySQL(ctx, cfg)
	if err != nil {
		return resources, err
	}
	resources.DB = db

	redisClient, err := openRedis(ctx, cfg)
	if err != nil {
		return resources, err
	}
	resources.Redis = redisClient

	return resources, nil
}

func openMySQL(ctx context.Context, cfg config.Config) (*gorm.DB, error) {
	if cfg.MySQLDSN == "" {
		slog.Warn("mysql dsn is empty, database-backed features are disabled")
		return nil, nil
	}
	if err := ensureDatabase(ctx, cfg.MySQLDSN); err != nil {
		return nil, err
	}

	db, err := gorm.Open(mysql.Open(cfg.MySQLDSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, err
	}

	if err := migrate(db); err != nil {
		return nil, err
	}
	if err := seedAdmin(db, cfg); err != nil {
		return nil, err
	}

	return db, nil
}

func ensureDatabase(ctx context.Context, dsn string) error {
	parsed, err := mysqlDriver.ParseDSN(dsn)
	if err != nil {
		return err
	}
	if parsed.DBName == "" {
		return nil
	}

	dbName := parsed.DBName
	parsed.DBName = ""
	rootDB, err := sql.Open("mysql", parsed.FormatDSN())
	if err != nil {
		return err
	}
	defer rootDB.Close()

	if err := rootDB.PingContext(ctx); err != nil {
		return err
	}

	quotedName := strings.ReplaceAll(dbName, "`", "``")
	_, err = rootDB.ExecContext(ctx, "CREATE DATABASE IF NOT EXISTS `"+quotedName+"` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci")
	return err
}

func openRedis(ctx context.Context, cfg config.Config) (*redis.Client, error) {
	if cfg.RedisAddr == "" {
		slog.Warn("redis addr is empty, redis-backed features are disabled")
		return nil, nil
	}

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       0,
	})
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return client, nil
}

func migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&model.User{},
		&model.Topic{},
		&model.Tag{},
		&model.Asset{},
		&model.SiteSetting{},
		&model.Notice{},
	); err != nil {
		return err
	}

	// Existing installations need topic_id populated before AutoMigrate adds the
	// Article -> Topic foreign key. Add only the column first, then copy legacy
	// taxonomy data without dropping categories/category_id.
	if db.Migrator().HasTable(&model.Article{}) {
		if !db.Migrator().HasColumn(&model.Article{}, "TopicID") {
			if err := db.Migrator().AddColumn(&model.Article{}, "TopicID"); err != nil {
				return err
			}
		}
		if err := migrateLegacyCategories(db); err != nil {
			return err
		}
	}

	return db.AutoMigrate(&model.Article{}, &model.ArticleVersion{})
}

type legacyCategoryRow struct {
	ID          uint64
	Name        string
	Slug        string
	Description string
	SortOrder   int
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt
}

type legacyArticleRow struct {
	CategoryID uint64
}

func (legacyArticleRow) TableName() string {
	return "articles"
}

// migrateLegacyCategories is intentionally idempotent. It keeps the legacy
// table and column in place so an upgrade can be rolled back or audited.
func migrateLegacyCategories(db *gorm.DB) error {
	if !db.Migrator().HasTable("categories") {
		return ensureArticleTopics(db, nil)
	}

	var categories []legacyCategoryRow
	if err := db.Unscoped().Table("categories").Find(&categories).Error; err != nil {
		return err
	}

	return db.Transaction(func(tx *gorm.DB) error {
		mapping := make(map[uint64]uint64, len(categories))
		for _, category := range categories {
			topicID, err := migrateLegacyCategory(tx, category)
			if err != nil {
				return err
			}
			mapping[category.ID] = topicID
		}
		return ensureArticleTopics(tx, mapping)
	})
}

func migrateLegacyCategory(db *gorm.DB, category legacyCategoryRow) (uint64, error) {
	var existing model.Topic
	err := db.Unscoped().Where("slug = ?", category.Slug).First(&existing).Error
	if err == nil {
		return existing.ID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}

	legacyIDAvailable := true
	err = db.Unscoped().First(&existing, category.ID).Error
	if err == nil {
		legacyIDAvailable = false
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}

	topic := model.Topic{
		Name:        category.Name,
		Label:       model.DefaultTopicLabel(category.Name),
		Slug:        category.Slug,
		Description: category.Description,
		SortOrder:   category.SortOrder,
		CreatedAt:   category.CreatedAt,
		UpdatedAt:   category.UpdatedAt,
		DeletedAt:   category.DeletedAt,
	}
	if legacyIDAvailable {
		topic.ID = category.ID
	}
	if err := db.Unscoped().Create(&topic).Error; err != nil {
		return 0, err
	}
	return topic.ID, nil
}

func ensureArticleTopics(db *gorm.DB, legacyMapping map[uint64]uint64) error {
	if !db.Migrator().HasTable(&model.Article{}) || !db.Migrator().HasColumn(&model.Article{}, "TopicID") {
		return nil
	}
	if legacyMapping != nil && db.Migrator().HasColumn(&legacyArticleRow{}, "CategoryID") {
		for categoryID, topicID := range legacyMapping {
			if err := db.Table("articles").
				Where("category_id = ? AND (topic_id IS NULL OR topic_id = 0)", categoryID).
				Update("topic_id", topicID).Error; err != nil {
				return err
			}
		}
	}
	if err := seedDefaultTopic(db); err != nil {
		return err
	}
	var notes model.Topic
	if err := db.Where("slug = ?", "notes").First(&notes).Error; err != nil {
		return err
	}
	return db.Table("articles").
		Where("topic_id IS NULL OR topic_id = 0").
		Update("topic_id", notes.ID).Error
}

func seedAdmin(db *gorm.DB, cfg config.Config) error {
	var count int64
	if err := db.Model(&model.User{}).Where("username = ?", cfg.AdminUsername).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(cfg.AdminPassword), bcrypt.DefaultCost)
		if err != nil {
			return err
		}

		user := model.User{
			Username:     cfg.AdminUsername,
			PasswordHash: string(passwordHash),
			Nickname:     cfg.AdminNickname,
			Role:         "owner",
			Status:       "active",
		}
		if err := db.Create(&user).Error; err != nil {
			return err
		}
	}

	if err := seedDefaultTopic(db); err != nil {
		return err
	}
	if err := seedDefaultTags(db); err != nil {
		return err
	}
	if err := seedDefaultSiteSetting(db); err != nil {
		return err
	}

	return nil
}

func seedDefaultTopic(db *gorm.DB) error {
	var topic model.Topic
	err := db.Unscoped().Where("slug = ?", "notes").First(&topic).Error
	if err == nil {
		if topic.DeletedAt.Valid {
			return db.Unscoped().Model(&topic).Update("deleted_at", nil).Error
		}
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return db.Create(&model.Topic{Name: "Notes", Label: "Notes", Slug: "notes", SortOrder: 1}).Error
}

func seedDefaultTags(db *gorm.DB) error {
	defaultTags := []model.Tag{
		{Name: "Go", Slug: "go", Color: "#5f8d62"},
		{Name: "Vue", Slug: "vue", Color: "#557ea8"},
	}
	for _, tag := range defaultTags {
		var count int64
		if err := db.Model(&model.Tag{}).Where("slug = ?", tag.Slug).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		if err := db.Create(&tag).Error; err != nil {
			return err
		}
	}
	return nil
}

func seedDefaultSiteSetting(db *gorm.DB) error {
	var row model.SiteSetting
	err := db.First(&row, 1).Error
	if err == nil {
		theme := appearance.NormalizeTheme(row.Theme)
		mode := appearance.NormalizeMode(row.Mode)
		updates := map[string]any{}
		if theme != row.Theme {
			updates["theme"] = theme
		}
		if mode != row.Mode {
			updates["mode"] = mode
		}
		themeElementsJSON, changed, err := normalizeStoredThemeElements(row.ThemeElementsJSON)
		if err != nil {
			return err
		}
		if changed {
			updates["theme_elements_json"] = themeElementsJSON
		}
		if len(updates) == 0 {
			return nil
		}
		return db.Model(&row).Updates(updates).Error
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	themeElementsJSON, err := marshalThemeElements(appearance.DefaultThemeElements())
	if err != nil {
		return err
	}
	return db.Create(&model.SiteSetting{
		ID:                1,
		SiteName:          "Solitude Blog",
		Author:            "Solitude King",
		Essay:             "Keep writing, keep shipping.",
		Theme:             appearance.DefaultTheme,
		Mode:              appearance.DefaultMode,
		ThemeElementsJSON: themeElementsJSON,
		SocialLinksJSON: `{
			"gitee": "",
			"bilibili": "",
			"douyin": "",
			"github": ""
		}`,
	}).Error
}

func normalizeStoredThemeElements(raw string) (string, bool, error) {
	stored := appearance.ThemeElementMap{}
	if raw == "" || json.Unmarshal([]byte(raw), &stored) != nil {
		payload, err := marshalThemeElements(appearance.DefaultThemeElements())
		return payload, true, err
	}
	normalized := appearance.NormalizeThemeElements(stored)
	if reflect.DeepEqual(stored, normalized) {
		return raw, false, nil
	}
	payload, err := marshalThemeElements(normalized)
	return payload, true, err
}

func marshalThemeElements(elements appearance.ThemeElementMap) (string, error) {
	payload, err := json.Marshal(elements)
	if err != nil {
		return "", err
	}
	return string(payload), nil
}

func Close(resources Resources) error {
	var closeErr error
	if resources.Redis != nil {
		closeErr = errors.Join(closeErr, resources.Redis.Close())
	}
	if resources.DB != nil {
		sqlDB, err := resources.DB.DB()
		if err != nil {
			closeErr = errors.Join(closeErr, err)
		} else {
			closeErr = errors.Join(closeErr, sqlDB.Close())
		}
	}
	return closeErr
}
