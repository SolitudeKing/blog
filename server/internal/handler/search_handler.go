package handler

import (
	"github.com/gin-gonic/gin"

	"solitude-blog/server/internal/pagination"
	"solitude-blog/server/internal/response"
	"solitude-blog/server/internal/service"
)

type SearchHandler struct {
	search *service.SearchService
}

func NewSearchHandler(search *service.SearchService) *SearchHandler {
	return &SearchHandler{search: search}
}

func (h *SearchHandler) Article(c *gin.Context) {
	page, pageSize := pagination.BindPage(c)
	items, pageInfo, err := h.search.Article(c.Request.Context(), service.ArticleSearchQuery{
		Page:     page,
		PageSize: pageSize,
		Keyword:  c.Query("keyword"),
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.List(c, items, pageInfo)
}
