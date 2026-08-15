package database

import (
	"context"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"reflect"
	"strings"
	"sync"
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
	if strings.TrimSpace(cfg.MySQLDSN) == "" {
		return nil, errors.New("MYSQL_DSN is required")
	}
	if err := registerMySQLTLSConfig(cfg.MySQLDSN); err != nil {
		return nil, err
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

	options := &redis.Options{
		Addr:     cfg.RedisAddr,
		Username: cfg.RedisUsername, // Redis 6+ ACL；留空时等价于仅密码。
		Password: cfg.RedisPassword,
		DB:       0,
	}
	if tlsConfig, enabled, err := redisTLSConfig(cfg.RedisTLS); err != nil {
		return nil, err
	} else if enabled {
		options.TLSConfig = tlsConfig
	}

	client := redis.NewClient(options)
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
		&model.RevokedRefreshToken{},
	); err != nil {
		return err
	}

	// 当前系统早期版本曾使用 category/category_id。升级时必须先补齐 topic_id，
	// 再由 AutoMigrate 建立 Article -> Topic 外键，避免迁移过程中出现悬空引用。
	// 这只是当前 Go 系统内部的结构升级兼容，不依赖或保留旧 Flask 运行代码。
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

// migrateLegacyCategories 幂等迁移当前系统早期的 category 数据，并暂留旧表和列，
// 便于升级失败时回滚或审计；它不代表继续兼容旧 Flask 博客的运行逻辑。
func migrateLegacyCategories(db *gorm.DB) error {
	if !db.Migrator().HasTable("categories") {
		return ensureArticleTopics(db, nil)
	}

	var categories []legacyCategoryRow
	if err := db.Unscoped().Table("categories").Find(&categories).Error; err != nil {
		return err
	}

	return db.Transaction(func(tx *gorm.DB) error {
		// 先把旧的默认 Notes 专题原地迁移为 NODES。categories 表会长期保留用于
		// 升级审计，因此后续每次启动都必须把它映射到 NODES，而不是重新创建 Notes。
		if err := seedDefaultTopics(tx); err != nil {
			return err
		}
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
	if isExactDefaultNotesCategory(category) {
		var nodes model.Topic
		if err := db.Unscoped().Where("slug = ?", model.TopicSlugNodes).First(&nodes).Error; err != nil {
			return 0, err
		}
		return nodes.ID, nil
	}

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

// isExactDefaultNotesCategory 只识别早期脚手架创建的完整默认行。
// 同名但描述、排序或删除状态不同的用户专题仍按普通旧专题迁移。
func isExactDefaultNotesCategory(category legacyCategoryRow) bool {
	return category.Name == "Notes" &&
		category.Slug == "notes" &&
		category.Description == "" &&
		category.SortOrder == 1 &&
		!category.DeletedAt.Valid
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
	if err := seedDefaultTopics(db); err != nil {
		return err
	}
	var nodes model.Topic
	if err := db.Where("slug = ?", model.TopicSlugNodes).First(&nodes).Error; err != nil {
		return err
	}
	return db.Table("articles").
		Where("topic_id IS NULL OR topic_id = 0").
		Update("topic_id", nodes.ID).Error
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

	if err := seedDefaultTopics(db); err != nil {
		return err
	}
	if err := seedDefaultSiteSetting(db); err != nil {
		return err
	}

	return nil
}

// seedDefaultTopics 先迁移旧脚手架中的 Notes 专题，再补齐缺失的正式专题。
// 已存在的正式专题可能已被管理员维护，因此启动时只恢复软删除状态，不覆盖其他字段。
func seedDefaultTopics(db *gorm.DB) error {
	if err := migrateExactNotesTopic(db); err != nil {
		return err
	}

	for _, initial := range model.DefaultTopics() {
		var topic model.Topic
		err := db.Unscoped().Where("slug = ?", initial.Slug).First(&topic).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := db.Create(&initial).Error; err != nil {
				return err
			}
			continue
		}
		if err != nil {
			return err
		}
		if updates := defaultTopicRestoreUpdates(topic); len(updates) > 0 {
			if err := db.Unscoped().Model(&topic).Updates(updates).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func defaultTopicRestoreUpdates(topic model.Topic) map[string]any {
	if !topic.DeletedAt.Valid {
		return nil
	}
	return map[string]any{"deleted_at": nil}
}

// migrateExactNotesTopic 只识别旧脚手架的完整默认值，并原地更新以保留专题 ID 与文章关联。
func migrateExactNotesTopic(db *gorm.DB) error {
	var legacy model.Topic
	result := db.Unscoped().Where("slug = ?", "notes").Limit(1).Find(&legacy)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return nil
	}
	if !isExactDefaultNotesTopic(legacy) {
		return nil
	}

	var existing model.Topic
	err := db.Unscoped().Where("slug = ?", model.TopicSlugNodes).First(&existing).Error
	if err == nil && existing.ID != legacy.ID {
		return errors.New("cannot migrate Notes topic because slug nodes already exists")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	nodes := model.DefaultTopics()[0]
	return db.Unscoped().Model(&legacy).Updates(map[string]any{
		"name":        nodes.Name,
		"label":       nodes.Label,
		"slug":        nodes.Slug,
		"description": nodes.Description,
		"sort_order":  nodes.SortOrder,
		"deleted_at":  nil,
	}).Error
}

// isExactDefaultNotesTopic 避免仅凭名称和 slug 覆盖管理员维护过的描述、封面或排序。
// 软删除状态不参与指纹，完整默认专题即使曾被删除也应迁移并恢复为正式 NODES。
func isExactDefaultNotesTopic(topic model.Topic) bool {
	return topic.Name == "Notes" &&
		topic.Label == "Notes" &&
		topic.Slug == "notes" &&
		topic.Description == "" &&
		topic.CoverURL == "" &&
		topic.SortOrder == 1
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

// mysqlTLSRegistry 跟踪我们已向 mysql_driver 注册过的 TLS 配置名，
// 防止同一进程内重复 RegisterTLSConfig 触发 panic。
// 外部用户通过 mysql_driver.RegisterTLSConfig 自定义的名字也会被记录。
var mysqlTLSRegistry sync.Map

// registerMySQLTLSConfig 解析 DSN 中的 tls= 参数并按需向 mysql_driver 注册 TLS 配置。
// 支持的 DSN 参数值：
//   - "true"         → 注册同名配置，使用系统根证书校验
//   - "skip-verify"  → 注册同名配置，跳过证书校验（仅开发或自签名证书场景）
//   - 其它非空值    → 假定用户已在外部通过 RegisterTLSConfig 注册过同名配置，本函数不重复注册
//   - 空 / "false"  → 不启用 TLS，直接返回
func registerMySQLTLSConfig(dsn string) error {
	parsed, err := mysqlDriver.ParseDSN(dsn)
	if err != nil {
		return err
	}
	tlsName := strings.TrimSpace(parsed.Params["tls"])
	switch tlsName {
	case "", "false":
		return nil
	case "true":
		return registerMySQLTLSConfigOnce(tlsName, &tls.Config{MinVersion: tls.VersionTLS12})
	case "skip-verify":
		return registerMySQLTLSConfigOnce(tlsName, &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: true,
		})
	default:
		// 假定用户已在外部注册过同名配置；如果未注册，mysql_driver 会在握手时报错。
		return nil
	}
}

func registerMySQLTLSConfigOnce(name string, cfg *tls.Config) error {
	if _, loaded := mysqlTLSRegistry.Load(name); loaded {
		return nil
	}
	if err := mysqlDriver.RegisterTLSConfig(name, cfg); err != nil {
		return err
	}
	mysqlTLSRegistry.Store(name, true)
	return nil
}

// redisTLSConfig 把 REDIS_TLS 字符串翻译为 redis.Options.TLSConfig 字段。
// 取值语义与 MYSQL_TLS 一致：false / 空 / true / skip-verify。
func redisTLSConfig(raw string) (*tls.Config, bool, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "false", "0", "off", "no":
		return nil, false, nil
	case "true", "1", "on":
		return &tls.Config{MinVersion: tls.VersionTLS12}, true, nil
	case "skip-verify":
		return &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: true,
		}, true, nil
	default:
		return nil, false, errors.New("REDIS_TLS only supports false / true / skip-verify")
	}
}
