package handler

import (
	"strings"

	"github.com/gin-gonic/gin"

	apperrors "solitude-blog/server/internal/errors"
	"solitude-blog/server/internal/response"
	"solitude-blog/server/internal/service"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.New(apperrors.CodeMalformedJSONBody))
		return
	}

	result, err := h.auth.Login(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req service.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, apperrors.New(apperrors.CodeMalformedJSONBody))
		return
	}
	if req.RefreshToken == "" {
		response.Error(c, apperrors.New(apperrors.CodeMissingRequiredField))
		return
	}

	result, err := h.auth.Refresh(req.RefreshToken)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, result)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	// 把当前 access / refresh token 的 jti 写入撤销表。
	// refreshToken 由前端登录后保存并随 logout 请求一起带回；
	// accessToken 始终由 Authorization 头携带。
	accessToken := strings.TrimSpace(strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer "))
	var body struct {
		RefreshToken string `json:"refresh_token"`
	}
	if c.Request.ContentLength > 0 {
		_ = c.ShouldBindJSON(&body)
	}
	if body.RefreshToken == "" {
		body.RefreshToken = c.PostForm("refresh_token")
	}
	if err := h.auth.LogoutRevoke(accessToken, body.RefreshToken); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, gin.H{"logged_out": true})
}
