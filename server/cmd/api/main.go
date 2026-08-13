package main

import (
	"context"
	"log/slog"
	"os"

	"solitude-blog/server/internal/bootstrap"
	"solitude-blog/server/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("api server configuration failed", "error", err)
		os.Exit(1)
	}
	app, resources, err := bootstrap.NewApp(context.Background(), cfg)
	if err != nil {
		slog.Error("api server bootstrap failed", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := bootstrap.Close(resources); err != nil {
			slog.Error("api resources close failed", "error", err)
		}
	}()

	slog.Info("api server starting", "addr", cfg.HTTPAddr(), "env", cfg.AppEnv)
	if err := app.Run(cfg.HTTPAddr()); err != nil {
		slog.Error("api server stopped", "error", err)
		os.Exit(1)
	}
}
