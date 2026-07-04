package database

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"strings"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

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
	return db.AutoMigrate(
		&model.User{},
		&model.Category{},
		&model.Tag{},
		&model.Article{},
		&model.SiteSetting{},
		&model.Notice{},
	)
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

	if err := seedDefaultCategory(db); err != nil {
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

func seedDefaultCategory(db *gorm.DB) error {
	var categoryCount int64
	if err := db.Model(&model.Category{}).Where("slug = ?", "notes").Count(&categoryCount).Error; err != nil {
		return err
	}
	if categoryCount > 0 {
		return nil
	}
	return db.Create(&model.Category{Name: "Notes", Slug: "notes", SortOrder: 1}).Error
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
	var count int64
	if err := db.Model(&model.SiteSetting{}).Where("id = ?", 1).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return db.Create(&model.SiteSetting{
		ID:       1,
		SiteName: "Solitude Blog",
		Author:   "Solitude King",
		Essay:    "Keep writing, keep shipping.",
		Theme:    "forest",
		Mode:     "light",
		SocialLinksJSON: `{
			"gitee": "",
			"bilibili": "",
			"douyin": "",
			"github": ""
		}`,
	}).Error
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
