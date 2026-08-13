package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"solitude-blog/server/internal/config"
)

// CorsForDev 是仅在开发环境启用的 CORS 中间件。
// 它让浏览器绕开 Vite proxy 直接请求 http://localhost:8080（API 进程），
// 方便调试 /uploads/... 静态资源或 Postman 等工具直打后端。
// 生产环境（APP_ENV=production）下退化为 no-op，避免任何 CORS 头被加到响应里。
func CorsForDev(cfg config.Config) gin.HandlerFunc {
	if !strings.EqualFold(strings.TrimSpace(cfg.AppEnv), "development") {
		return func(c *gin.Context) { c.Next() }
	}
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Version, X-Request-ID")
		c.Header("Access-Control-Expose-Headers", "X-Request-ID")
		c.Header("Access-Control-Max-Age", "600")

		// 预检请求直接放行，避免 Gin's X-API-Version 中间件把它误判为非 v1。
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
