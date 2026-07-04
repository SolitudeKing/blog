package bootstrap

import (
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"

	"solitude-blog/server/internal/config"
	"solitude-blog/server/internal/handler"
	"solitude-blog/server/internal/middleware"
	"solitude-blog/server/internal/router"
	"solitude-blog/server/internal/service"
)

func NewApp(cfg config.Config) *gin.Engine {
	setupLogger(cfg)

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(middleware.RequestID())
	engine.Use(middleware.Recovery())
	engine.Use(middleware.APIVersion(cfg.APIVersion))

	authService := service.NewAuthService(cfg)
	settingService := service.NewSettingService()
	articleService := service.NewArticleService()

	handlers := router.Handlers{
		Health:  handler.NewHealthHandler(),
		Auth:    handler.NewAuthHandler(authService),
		User:    handler.NewUserHandler(),
		Setting: handler.NewSettingHandler(settingService),
		Article: handler.NewArticleHandler(articleService),
	}

	router.Register(engine, handlers, cfg)
	return engine
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
