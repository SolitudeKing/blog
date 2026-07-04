package handler

import (
	"github.com/gin-gonic/gin"

	apperrors "solitude-blog/server/internal/errors"
	"solitude-blog/server/internal/response"
	"solitude-blog/server/internal/service"
)

type CategoryHandler struct {
	category *service.CategoryService
}

func NewCategoryHandler(category *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{category: category}
}

func (h *CategoryHandler) List(c *gin.Context) {
	items, err := h.category.List()
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, items)
}

func (h *CategoryHandler) Create(c *gin.Context) {
	var req service.CategorySaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.New(apperrors.CodeMalformedJSONBody))
		return
	}
	item, err := h.category.Create(req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, item)
}

func (h *CategoryHandler) Update(c *gin.Context) {
	var req service.CategorySaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.New(apperrors.CodeMalformedJSONBody))
		return
	}
	item, err := h.category.Update(c.Param("id"), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, item)
}

func (h *CategoryHandler) Delete(c *gin.Context) {
	if err := h.category.Delete(c.Param("id")); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, gin.H{"id": c.Param("id"), "deleted": true})
}
