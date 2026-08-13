package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	apperrors "solitude-blog/server/internal/errors"
	"solitude-blog/server/internal/response"
	"solitude-blog/server/internal/service"
)

type AssetHandler struct {
	asset *service.AssetService
}

func NewAssetHandler(asset *service.AssetService) *AssetHandler {
	return &AssetHandler{asset: asset}
}

func (h *AssetHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	items, pageInfo, err := h.asset.List(service.AssetListQuery{
		Page:     page,
		PageSize: pageSize,
		Keyword:  c.Query("keyword"),
		Mime:     c.Query("mime"),
	})
	if err != nil {
		response.Error(c, err)
		return
	}
	response.List(c, items, pageInfo)
}

func (h *AssetHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Error(c, apperrors.New(apperrors.CodeMissingRequiredField))
		return
	}
	item, err := h.asset.Upload(file, c.PostForm("display_name"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, item)
}

func (h *AssetHandler) Update(c *gin.Context) {
	var req service.AssetUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.New(apperrors.CodeMalformedJSONBody))
		return
	}
	item, err := h.asset.Update(c.Param("id"), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, item)
}

func (h *AssetHandler) Delete(c *gin.Context) {
	if err := h.asset.Delete(c.Param("id")); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, gin.H{"id": c.Param("id"), "deleted": true})
}

func (h *AssetHandler) ReferenceList(c *gin.Context) {
	page, _ := strconv.Atoi(c.Query("page"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	items, pageInfo, err := h.asset.ReferenceList(c.Param("id"), page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.List(c, items, pageInfo)
}
