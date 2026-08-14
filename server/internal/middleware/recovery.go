package middleware

import (
	"log/slog"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	apperrors "solitude-blog/server/internal/errors"
	"solitude-blog/server/internal/response"
)

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		// 记录 panic 时的关键上下文（method / path / user_id / request_id），
		// 便于排障。生产部署时由 slog 输出 JSON 供聚合系统采集。
		slog.Error("panic recovered",
			"error", recovered,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"query", c.Request.URL.RawQuery,
			"user_id", c.GetUint64("user_id"),
			"request_id", c.GetString(RequestIDKey),
			"stack", string(debug.Stack()),
		)
		response.Error(c, apperrors.New(apperrors.CodeInternalServerError))
		c.Abort()
	})
}
