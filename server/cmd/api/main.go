package main

import (
	"log/slog"
	"os"

	"solitude-blog/server/internal/bootstrap"
	"solitude-blog/server/internal/config"
)

func main() {
	cfg := config.Load()
	app := bootstrap.NewApp(cfg)

	slog.Info("api server starting", "addr", cfg.HTTPAddr(), "env", cfg.AppEnv)
	if err := app.Run(cfg.HTTPAddr()); err != nil {
		slog.Error("api server stopped", "error", err)
		os.Exit(1)
	}
}
