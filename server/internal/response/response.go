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

type ListBody[T any] struct {
	Code    int                   `json:"code"`
	Message string                `json:"message"`
	Data    []T                   `json:"data"`
	Page    pagination.CursorPage `json:"page"`
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

func List[T any](c *gin.Context, data []T, page pagination.CursorPage) {
	c.JSON(http.StatusOK, ListBody[T]{
		Code:    apperrors.CodeOK,
		Message: apperrors.Message(apperrors.CodeOK),
		Data:    data,
		Page:    page,
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
