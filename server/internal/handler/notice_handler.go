package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "solitude-blog/server/internal/errors"
	"solitude-blog/server/internal/response"
	"solitude-blog/server/internal/service"
)

type NoticeHandler struct {
	notice *service.NoticeService
}

func NewNoticeHandler(notice *service.NoticeService) *NoticeHandler {
	return &NoticeHandler{notice: notice}
}

func (h *NoticeHandler) Active(c *gin.Context) {
	item, err := h.notice.Active()
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, item)
}

func (h *NoticeHandler) ManageList(c *gin.Context) {
	query := noticeListQuery(c)
	items, page, err := h.notice.ManageList(query)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.List(c, items, page)
}

func (h *NoticeHandler) Create(c *gin.Context) {
	var req service.NoticeSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.New(apperrors.CodeMalformedJSONBody))
		return
	}
	item, err := h.notice.Create(req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, item)
}

func (h *NoticeHandler) Update(c *gin.Context) {
	var req service.NoticeSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.New(apperrors.CodeMalformedJSONBody))
		return
	}
	item, err := h.notice.Update(c.Param("id"), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, item)
}

func (h *NoticeHandler) Delete(c *gin.Context) {
	if err := h.notice.Delete(c.Param("id")); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, gin.H{"id": c.Param("id"), "deleted": true})
}

func noticeListQuery(c *gin.Context) service.NoticeListQuery {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	query := service.NoticeListQuery{
		Page:     page,
		PageSize: pageSize,
		Keyword:  c.Query("keyword"),
	}
	if enabled := c.Query("enabled"); enabled != "" {
		parsed, err := strconv.ParseBool(enabled)
		if err == nil {
			query.Enabled = &parsed
		}
	}
	return query
}
