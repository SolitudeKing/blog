package handler

import (
	"github.com/gin-gonic/gin"

	"solitude-blog/server/internal/response"
	"solitude-blog/server/internal/service"
)

type DashboardHandler struct {
	dashboard *service.DashboardService
}

func NewDashboardHandler(dashboard *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashboard: dashboard}
}

func (h *DashboardHandler) Summary(c *gin.Context) {
	item, err := h.dashboard.Summary()
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, item)
}
