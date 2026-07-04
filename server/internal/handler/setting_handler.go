package handler

import (
	"github.com/gin-gonic/gin"

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
	response.OK(c, h.setting.Lobby())
}

func (h *SettingHandler) Detail(c *gin.Context) {
	response.OK(c, h.setting.Detail())
}

func (h *SettingHandler) Update(c *gin.Context) {
	response.OK(c, h.setting.Detail())
}
