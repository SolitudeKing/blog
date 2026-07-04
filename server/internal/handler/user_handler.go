package handler

import (
	"github.com/gin-gonic/gin"

	"solitude-blog/server/internal/response"
)

type UserHandler struct{}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

func (h *UserHandler) Info(c *gin.Context) {
	response.OK(c, gin.H{
		"id":       c.GetUint64("user_id"),
		"username": c.GetString("username"),
		"role":     c.GetString("role"),
	})
}
