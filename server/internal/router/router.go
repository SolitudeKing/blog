package router

import (
	"github.com/gin-gonic/gin"

	"solitude-blog/server/internal/config"
	"solitude-blog/server/internal/handler"
	"solitude-blog/server/internal/middleware"
)

type Handlers struct {
	Health  *handler.HealthHandler
	Auth    *handler.AuthHandler
	User    *handler.UserHandler
	Setting *handler.SettingHandler
	Article *handler.ArticleHandler
}

func Register(r *gin.Engine, h Handlers, cfg config.Config) {
	r.GET("healthz", h.Health.Healthz)

	r.POST("auth/login", h.Auth.Login)
	r.POST("auth/refresh", h.Auth.Refresh)
	r.POST("auth/logout", middleware.AuthRequired(), h.Auth.Logout)

	r.GET("user/info", middleware.AuthRequired(), h.User.Info)

	r.GET("setting/lobby", h.Setting.Lobby)
	r.GET("setting/detail", middleware.AuthRequired(), h.Setting.Detail)
	r.PUT("setting/update", middleware.AuthRequired(), h.Setting.Update)

	r.GET("article/list", h.Article.PublicList)
	r.GET("article/detail/:slug", h.Article.Detail)
	r.GET("article/manage-list", middleware.AuthRequired(), h.Article.ManageList)
	r.POST("article/create", middleware.AuthRequired(), h.Article.Create)
	r.GET("article/info/:id", middleware.AuthRequired(), h.Article.Info)
	r.PUT("article/update/:id", middleware.AuthRequired(), h.Article.Update)
	r.DELETE("article/delete/:id", middleware.AuthRequired(), h.Article.Delete)
}
