package handler

import (
	"github.com/gin-gonic/gin"

	apperrors "solitude-blog/server/internal/errors"
	"solitude-blog/server/internal/response"
	"solitude-blog/server/internal/service"
)

type TopicHandler struct {
	topic *service.TopicService
}

func NewTopicHandler(topic *service.TopicService) *TopicHandler {
	return &TopicHandler{topic: topic}
}

func (h *TopicHandler) List(c *gin.Context) {
	items, err := h.topic.List()
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, items)
}

func (h *TopicHandler) Create(c *gin.Context) {
	var req service.TopicSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.New(apperrors.CodeMalformedJSONBody))
		return
	}
	item, err := h.topic.Create(req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, item)
}

func (h *TopicHandler) Update(c *gin.Context) {
	var req service.TopicSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.New(apperrors.CodeMalformedJSONBody))
		return
	}
	item, err := h.topic.Update(c.Param("id"), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, item)
}

func (h *TopicHandler) Delete(c *gin.Context) {
	if err := h.topic.Delete(c.Param("id")); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, gin.H{"id": c.Param("id"), "deleted": true})
}
