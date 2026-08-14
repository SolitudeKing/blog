package response

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apperrors "solitude-blog/server/internal/errors"
	"solitude-blog/server/internal/pagination"
)

type Body[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

// ListBody 增长型列表的统一响应。
// 分页信息以 count / page / page_size / total / has_more 五个顶层字段承载。
type ListBody[T any] struct {
	Code     int      `json:"code"`
	Message  string   `json:"message"`
	Data     []T      `json:"data"`
	Count    int      `json:"count"`     // 当前页条目数（= len(data)）
	Page     int      `json:"page"`      // 当前页码 1-based
	PageSize int      `json:"page_size"` // 每页大小
	Total    int      `json:"total"`     // 符合条件的总条数
	HasMore  bool     `json:"has_more"`  // 是否还有下一页
}

func OK[T any](c *gin.Context, data T) {
	c.JSON(http.StatusOK, Body[T]{
		Code:    apperrors.CodeOK,
		Message: apperrors.Message(apperrors.CodeOK),
		Data:    data,
	})
}

func Created[T any](c *gin.Context, data T) {
	c.JSON(http.StatusCreated, Body[T]{
		Code:    apperrors.CodeOK,
		Message: apperrors.Message(apperrors.CodeOK),
		Data:    data,
	})
}

// List 输出增长型列表响应。count 自动从 data 长度计算；total 由调用方通过 pageInfo.Total 提供。
func List[T any](c *gin.Context, data []T, pageInfo pagination.ListPage) {
	c.JSON(http.StatusOK, ListBody[T]{
		Code:     apperrors.CodeOK,
		Message:  apperrors.Message(apperrors.CodeOK),
		Data:     data,
		Count:    len(data),
		Page:     pageInfo.Page,
		PageSize: pageInfo.PageSize,
		Total:    pageInfo.Total,
		HasMore:  pageInfo.HasMore,
	})
}

func Error(c *gin.Context, err error) {
	appErr, ok := err.(apperrors.AppError)
	if !ok {
		appErr = apperrors.New(apperrors.CodeInternalServerError)
	}

	c.JSON(appErr.HTTPStatus, Body[any]{
		Code:    appErr.Code,
		Message: appErr.Message,
		Data:    nil,
	})
}
