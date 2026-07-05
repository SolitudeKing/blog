package handler

import (
	"strconv"

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
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	items, page, err := h.search.Article(c.Request.Context(), service.ArticleSearchQuery{
		Cursor:  c.Query("cursor"),
		Limit:   pagination.NormalizeLimit(limit),
		Keyword: c.Query("keyword"),
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.List(c, items, page)
}
