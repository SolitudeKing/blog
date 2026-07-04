package bootstrap

import (
	"context"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"

	"solitude-blog/server/internal/config"
	"solitude-blog/server/internal/database"
	"solitude-blog/server/internal/handler"
	"solitude-blog/server/internal/middleware"
	"solitude-blog/server/internal/router"
	"solitude-blog/server/internal/service"
)

func NewApp(ctx context.Context, cfg config.Config) (*gin.Engine, database.Resources, error) {
	setupLogger(cfg)
	resources, err := database.Open(ctx, cfg)
	if err != nil {
		return nil, resources, err
	}

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(middleware.RequestID())
	engine.Use(middleware.Recovery())
	engine.Use(middleware.APIVersion(cfg.APIVersion))

	authService := service.NewAuthService(cfg, resources.DB)
	settingService := service.NewSettingService()
	articleService := service.NewArticleService(resources.DB)
	categoryService := service.NewCategoryService(resources.DB)
	tagService := service.NewTagService(resources.DB)

	handlers := router.Handlers{
		Health:       handler.NewHealthHandler(resources),
		Auth:         handler.NewAuthHandler(authService),
		AuthRequired: middleware.AuthRequired(authService),
		User:         handler.NewUserHandler(),
		Setting:      handler.NewSettingHandler(settingService),
		Article:      handler.NewArticleHandler(articleService),
		Category:     handler.NewCategoryHandler(categoryService),
		Tag:          handler.NewTagHandler(tagService),
	}

	router.Register(engine, handlers, cfg)
	return engine, resources, nil
}

func Close(resources database.Resources) error {
	return database.Close(resources)
}

func setupLogger(cfg config.Config) {
	level := slog.LevelInfo
	if cfg.AppEnv == "development" {
		level = slog.LevelDebug
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
	slog.SetDefault(logger)
}
