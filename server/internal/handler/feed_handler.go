package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"solitude-blog/server/internal/response"
	"solitude-blog/server/internal/service"
)

type FeedHandler struct {
	feed        *service.FeedService
	siteBaseURL string
}

func NewFeedHandler(feed *service.FeedService, siteBaseURL string) *FeedHandler {
	return &FeedHandler{feed: feed, siteBaseURL: strings.TrimRight(siteBaseURL, "/")}
}

func (h *FeedHandler) RSS(c *gin.Context) {
	payload, err := h.feed.RSS(c.Request.Context(), h.baseURL(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	c.Data(http.StatusOK, "application/rss+xml; charset=utf-8", payload)
}

func (h *FeedHandler) Sitemap(c *gin.Context) {
	payload, err := h.feed.Sitemap(c.Request.Context(), h.baseURL(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	c.Data(http.StatusOK, "application/xml; charset=utf-8", payload)
}

func (h *FeedHandler) baseURL(c *gin.Context) string {
	if h.siteBaseURL != "" {
		return h.siteBaseURL
	}
	scheme := c.GetHeader("X-Forwarded-Proto")
	if scheme == "" {
		scheme = "http"
	}
	host := c.GetHeader("X-Forwarded-Host")
	if host == "" {
		host = c.Request.Host
	}
	return strings.TrimRight(scheme+"://"+host, "/")
}
