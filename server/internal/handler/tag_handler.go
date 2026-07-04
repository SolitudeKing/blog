package handler

import (
	"github.com/gin-gonic/gin"

	apperrors "solitude-blog/server/internal/errors"
	"solitude-blog/server/internal/response"
	"solitude-blog/server/internal/service"
)

type TagHandler struct {
	tag *service.TagService
}

func NewTagHandler(tag *service.TagService) *TagHandler {
	return &TagHandler{tag: tag}
}

func (h *TagHandler) List(c *gin.Context) {
	items, err := h.tag.List()
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, items)
}

func (h *TagHandler) Create(c *gin.Context) {
	var req service.TagSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.New(apperrors.CodeMalformedJSONBody))
		return
	}
	item, err := h.tag.Create(req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, item)
}

func (h *TagHandler) Update(c *gin.Context) {
	var req service.TagSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.New(apperrors.CodeMalformedJSONBody))
		return
	}
	item, err := h.tag.Update(c.Param("id"), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, item)
}

func (h *TagHandler) Delete(c *gin.Context) {
	if err := h.tag.Delete(c.Param("id")); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, gin.H{"id": c.Param("id"), "deleted": true})
}
