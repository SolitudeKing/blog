package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	apperrors "solitude-blog/server/internal/errors"
	"solitude-blog/server/internal/response"
)

func APIVersion(expected string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/healthz" || c.Request.URL.Path == "healthz" || strings.HasPrefix(c.Request.URL.Path, "/uploads/") {
			c.Next()
			return
		}

		version := c.GetHeader("X-API-Version")
		if version != expected {
			response.Error(c, apperrors.New(apperrors.CodeUnsupportedAPIVersion))
			c.Abort()
			return
		}

		c.Next()
	}
}
