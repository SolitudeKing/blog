package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"

	"solitude-blog/server/internal/config"
	"solitude-blog/server/internal/database"
	"solitude-blog/server/internal/handler"
	"solitude-blog/server/internal/middleware"
	"solitude-blog/server/internal/router"
	"solitude-blog/server/internal/service"
	"solitude-blog/server/internal/storage"
)

func NewApp(ctx context.Context, cfg config.Config) (*gin.Engine, database.Resources, error) {
	if err := cfg.Validate(); err != nil {
		return nil, database.Resources{}, err
	}
	setupLogger(cfg)
	resources, err := database.Open(ctx, cfg)
	if err != nil {
		return nil, resources, err
	}

	objectStore, err := storage.NewFromConfig(ctx, cfg)
	if err != nil {
		return nil, resources, fmt.Errorf("init storage: %w", err)
	}
	slog.Info("object storage initialized", "driver", objectStore.Kind())

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(middleware.RequestID())
	engine.Use(middleware.Recovery())
	engine.Use(middleware.CorsForDev(cfg))
	engine.Use(middleware.APIVersion(cfg.APIVersion))

	authService := service.NewAuthService(cfg, resources.DB)
	settingService := service.NewSettingService(resources.DB, resources.Redis)
	articleService := service.NewArticleService(resources.DB, resources.Redis)
	topicService := service.NewTopicService(resources.DB, resources.Redis)
	tagService := service.NewTagService(resources.DB, resources.Redis)
	noticeService := service.NewNoticeService(resources.DB)
	dashboardService := service.NewDashboardService(resources.DB)
	assetService := service.NewAssetService(resources.DB, objectStore)
	feedService := service.NewFeedService(resources.DB)
	searchService := service.NewSearchService(resources.DB)
	suggestionService := service.NewSuggestionService(topicService, tagService, resources.Redis)

	handlers := router.Handlers{
		Health:       handler.NewHealthHandler(resources),
		Auth:         handler.NewAuthHandler(authService),
		AuthRequired: middleware.AuthRequired(authService),
		User:         handler.NewUserHandler(),
		Setting:      handler.NewSettingHandler(settingService),
		Article:      handler.NewArticleHandler(articleService),
		Topic:        handler.NewTopicHandler(topicService),
		Tag:          handler.NewTagHandler(tagService),
		Notice:       handler.NewNoticeHandler(noticeService),
		Dashboard:    handler.NewDashboardHandler(dashboardService),
		Asset:        handler.NewAssetHandler(assetService),
		Feed:         handler.NewFeedHandler(feedService, cfg.SiteBaseURL),
		Search:       handler.NewSearchHandler(searchService, suggestionService),
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
