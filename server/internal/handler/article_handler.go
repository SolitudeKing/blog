package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "solitude-blog/server/internal/errors"
	"solitude-blog/server/internal/pagination"
	"solitude-blog/server/internal/response"
	"solitude-blog/server/internal/service"
)

type ArticleHandler struct {
	article *service.ArticleService
}

func NewArticleHandler(article *service.ArticleService) *ArticleHandler {
	return &ArticleHandler{article: article}
}

func (h *ArticleHandler) PublicList(c *gin.Context) {
	query := articleListQuery(c)
	items, page, err := h.article.PublicList(query)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.List(c, items, page)
}

func (h *ArticleHandler) Detail(c *gin.Context) {
	item, err := h.article.Detail(c.Param("slug"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, item)
}

func (h *ArticleHandler) ManageList(c *gin.Context) {
	query := articleListQuery(c)
	items, page, err := h.article.ManageList(query)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.List(c, items, page)
}

func (h *ArticleHandler) Create(c *gin.Context) {
	var req service.ArticleCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.New(apperrors.CodeMalformedJSONBody))
		return
	}
	item, err := h.article.Create(req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, item)
}

func (h *ArticleHandler) Info(c *gin.Context) {
	item, err := h.article.Info(c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, item)
}

func (h *ArticleHandler) VersionList(c *gin.Context) {
	items, err := h.article.VersionList(c.Param("id"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, items)
}

func (h *ArticleHandler) Update(c *gin.Context) {
	var req service.ArticleUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.New(apperrors.CodeMalformedJSONBody))
		return
	}
	item, err := h.article.Update(c.Param("id"), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, item)
}

func (h *ArticleHandler) Delete(c *gin.Context) {
	if err := h.article.Delete(c.Param("id")); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, gin.H{"id": c.Param("id"), "deleted": true})
}

func articleListQuery(c *gin.Context) service.ArticleListQuery {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	return service.ArticleListQuery{
		Cursor:   c.Query("cursor"),
		Limit:    pagination.NormalizeLimit(limit),
		Category: c.Query("category"),
		Tag:      c.Query("tag"),
		Keyword:  c.Query("keyword"),
		Status:   c.Query("status"),
	}
}
