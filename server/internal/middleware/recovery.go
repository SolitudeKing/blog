package middleware

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	apperrors "solitude-blog/server/internal/errors"
	"solitude-blog/server/internal/response"
)

func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		slog.Error("panic recovered", "error", recovered)
		response.Error(c, apperrors.New(apperrors.CodeInternalServerError))
		c.Abort()
	})
}
