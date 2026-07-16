package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"solitude-blog/server/internal/database"
	apperrors "solitude-blog/server/internal/errors"
	"solitude-blog/server/internal/response"
)

type HealthHandler struct {
	resources database.Resources
}

func NewHealthHandler(resources database.Resources) *HealthHandler {
	return &HealthHandler{resources: resources}
}

func (h *HealthHandler) Healthz(c *gin.Context) {
	failureCode := apperrors.CodeOK
	status := gin.H{
		"api":   "ok",
		"time":  time.Now().UTC().Format(time.RFC3339),
		"mysql": "disabled",
		"redis": "disabled",
	}

	if h.resources.DB != nil {
		status["mysql"] = "ok"
		if sqlDB, err := h.resources.DB.DB(); err != nil || sqlDB.PingContext(c.Request.Context()) != nil {
			status["mysql"] = "error"
			failureCode = apperrors.CodeDatabaseUnavailable
		}
	}
	if h.resources.Redis != nil {
		status["redis"] = "ok"
		if err := h.resources.Redis.Ping(c.Request.Context()).Err(); err != nil {
			status["redis"] = "error"
			if failureCode == apperrors.CodeOK {
				failureCode = apperrors.CodeCacheUnavailable
			}
		}
	}

	if failureCode != apperrors.CodeOK {
		c.JSON(http.StatusServiceUnavailable, response.Body[gin.H]{
			Code:    failureCode,
			Message: apperrors.Message(failureCode),
			Data:    status,
		})
		return
	}
	response.OK(c, status)
}
