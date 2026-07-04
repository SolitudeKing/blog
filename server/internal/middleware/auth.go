package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	apperrors "solitude-blog/server/internal/errors"
	"solitude-blog/server/internal/response"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") || len(strings.TrimPrefix(header, "Bearer ")) == 0 {
			response.Error(c, apperrors.New(apperrors.CodeUnauthorized))
			c.Abort()
			return
		}

		c.Set("user_id", "1")
		c.Next()
	}
}
