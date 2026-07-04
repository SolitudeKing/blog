package handler

import (
	"github.com/gin-gonic/gin"

	apperrors "solitude-blog/server/internal/errors"
	"solitude-blog/server/internal/response"
	"solitude-blog/server/internal/service"
)

type SettingHandler struct {
	setting *service.SettingService
}

func NewSettingHandler(setting *service.SettingService) *SettingHandler {
	return &SettingHandler{setting: setting}
}

func (h *SettingHandler) Lobby(c *gin.Context) {
	item, err := h.setting.Lobby()
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, item)
}

func (h *SettingHandler) Detail(c *gin.Context) {
	item, err := h.setting.Detail()
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, item)
}

func (h *SettingHandler) Update(c *gin.Context) {
	var req service.SettingSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.New(apperrors.CodeMalformedJSONBody))
		return
	}
	item, err := h.setting.Update(req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, item)
}
