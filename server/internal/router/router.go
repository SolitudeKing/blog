package router

import (
	"github.com/gin-gonic/gin"

	"solitude-blog/server/internal/config"
	"solitude-blog/server/internal/handler"
)

type Handlers struct {
	Health       *handler.HealthHandler
	Auth         *handler.AuthHandler
	AuthRequired gin.HandlerFunc
	User         *handler.UserHandler
	Setting      *handler.SettingHandler
	Article      *handler.ArticleHandler
	Category     *handler.CategoryHandler
	Tag          *handler.TagHandler
	Notice       *handler.NoticeHandler
	Dashboard    *handler.DashboardHandler
	Asset        *handler.AssetHandler
	Feed         *handler.FeedHandler
	Search       *handler.SearchHandler
}

func Register(r *gin.Engine, h Handlers, cfg config.Config) {
	r.Static("/uploads", cfg.StorageLocalRoot)

	r.GET("healthz", h.Health.Healthz)
	r.GET("rss.xml", h.Feed.RSS)
	r.GET("sitemap.xml", h.Feed.Sitemap)
	r.GET("search/article", h.Search.Article)

	r.POST("auth/login", h.Auth.Login)
	r.POST("auth/refresh", h.Auth.Refresh)
	r.POST("auth/logout", h.AuthRequired, h.Auth.Logout)

	r.GET("user/info", h.AuthRequired, h.User.Info)

	r.GET("dashboard/summary", h.AuthRequired, h.Dashboard.Summary)

	r.GET("asset/list", h.AuthRequired, h.Asset.List)
	r.POST("asset/upload", h.AuthRequired, h.Asset.Upload)
	r.PUT("asset/update/:id", h.AuthRequired, h.Asset.Update)
	r.DELETE("asset/delete/:id", h.AuthRequired, h.Asset.Delete)
	r.GET("asset/reference-list/:id", h.AuthRequired, h.Asset.ReferenceList)

	r.GET("setting/lobby", h.Setting.Lobby)
	r.GET("setting/detail", h.AuthRequired, h.Setting.Detail)
	r.PUT("setting/update", h.AuthRequired, h.Setting.Update)

	r.GET("article/list", h.Article.PublicList)
	r.GET("article/detail/:slug", h.Article.Detail)
	r.GET("article/manage-list", h.AuthRequired, h.Article.ManageList)
	r.POST("article/create", h.AuthRequired, h.Article.Create)
	r.GET("article/info/:id", h.AuthRequired, h.Article.Info)
	r.GET("article/version-list/:id", h.AuthRequired, h.Article.VersionList)
	r.PUT("article/update/:id", h.AuthRequired, h.Article.Update)
	r.DELETE("article/delete/:id", h.AuthRequired, h.Article.Delete)

	r.GET("category/list", h.Category.List)
	r.POST("category/create", h.AuthRequired, h.Category.Create)
	r.PUT("category/update/:id", h.AuthRequired, h.Category.Update)
	r.DELETE("category/delete/:id", h.AuthRequired, h.Category.Delete)

	r.GET("tag/list", h.Tag.List)
	r.POST("tag/create", h.AuthRequired, h.Tag.Create)
	r.PUT("tag/update/:id", h.AuthRequired, h.Tag.Update)
	r.DELETE("tag/delete/:id", h.AuthRequired, h.Tag.Delete)

	r.GET("notice/active", h.Notice.Active)
	r.GET("notice/manage-list", h.AuthRequired, h.Notice.ManageList)
	r.POST("notice/create", h.AuthRequired, h.Notice.Create)
	r.PUT("notice/update/:id", h.AuthRequired, h.Notice.Update)
	r.DELETE("notice/delete/:id", h.AuthRequired, h.Notice.Delete)
}
