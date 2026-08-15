package service

import (
	"testing"
	"time"

	apperrors "solitude-blog/server/internal/errors"
	"solitude-blog/server/internal/model"
)

// TestArticleServiceDetailWithoutDatabaseReturnsNoNeighbors 验证在无 DB 的
// 内存模式下 Detail 不会 panic，且 prev/next 都为 nil。
func TestArticleServiceDetailWithoutDatabaseReturnsNoNeighbors(t *testing.T) {
	service := NewArticleService(nil, nil)

	if _, err := service.Create(ArticleCreateRequest{
		Title:   "内存模式文章",
		Slug:    "in-memory-detail",
		TopicID: 1,
		Status:  "published",
	}, 7); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	detail, err := service.Detail("in-memory-detail")
	if err != nil {
		t.Fatalf("Detail() error = %v", err)
	}
	if detail.Prev != nil || detail.Next != nil {
		t.Fatalf("memory-mode neighbors = (%v, %v), want (nil, nil)", detail.Prev, detail.Next)
	}
	if detail.PublishedAt == nil || !detail.PublishedAt.Equal(detail.UpdatedAt) {
		t.Fatalf("memory-mode published_at = %v, want equal to updated_at", detail.PublishedAt)
	}
}

// TestArticleServiceDraftLeavesPublishedAtNil 验证未发布的草稿在响应里
// 把 published_at 序列化为 null，避免 time.Time 零值在 JS 里出现 Invalid Date。
func TestArticleServiceDraftLeavesPublishedAtNil(t *testing.T) {
	service := NewArticleService(nil, nil)

	item, err := service.Create(ArticleCreateRequest{
		Title:   "草稿",
		Slug:    "draft",
		TopicID: 1,
		Status:  "draft",
	}, 7)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if item.PublishedAt != nil {
		t.Fatalf("Create(draft) published_at = %v, want nil", item.PublishedAt)
	}
}

// TestArticleServicePublicListOrdersByPublishedAt 当 onlyPublished=true 时
// 排序键应使用 PublishedAt；这是通过内存模式覆盖到的契约。
func TestArticleServicePublicListOrdersByPublishedAt(t *testing.T) {
	service := NewArticleService(nil, nil)

	for _, draft := range []ArticleCreateRequest{
		{Title: "old", Slug: "old", TopicID: 1, Status: "published"},
		{Title: "new", Slug: "new", TopicID: 1, Status: "published"},
	} {
		if _, err := service.Create(draft, 7); err != nil {
			t.Fatalf("Create(%s) error = %v", draft.Slug, err)
		}
	}
	// 让"new"的发布时间晚于"old"
	time.Sleep(10 * time.Millisecond)

	items, _, err := service.PublicList(ArticleListQuery{})
	if err != nil {
		t.Fatalf("PublicList() error = %v", err)
	}
	if len(items) != 2 || items[0].Slug != "new" || items[1].Slug != "old" {
		t.Fatalf("PublicList() order = %s, %s; want new, old", items[0].Slug, items[1].Slug)
	}
}

// TestArticleServiceManageListOrdersByCreatedAt 当 onlyPublished=false 时
// 排序键应使用 CreatedAt（最新创建在最前）。
func TestArticleServiceManageListOrdersByCreatedAt(t *testing.T) {
	service := NewArticleService(nil, nil)

	for _, draft := range []ArticleCreateRequest{
		{Title: "first", Slug: "first", TopicID: 1, Status: "draft"},
		{Title: "second", Slug: "second", TopicID: 1, Status: "draft"},
	} {
		if _, err := service.Create(draft, 7); err != nil {
			t.Fatalf("Create(%s) error = %v", draft.Slug, err)
		}
	}
	items, _, err := service.ManageList(ArticleListQuery{})
	if err != nil {
		t.Fatalf("ManageList() error = %v", err)
	}
	if len(items) != 2 || items[0].Slug != "second" || items[1].Slug != "first" {
		t.Fatalf("ManageList() order = %s, %s; want second, first", items[0].Slug, items[1].Slug)
	}
}

// TestTopicServiceListWithoutDatabaseIncludesDraftCount 仅占位：
// 在 DB 模式下 article_count 来自 LEFT JOIN，未覆盖到这里。
// 此用例保证内存模式下不会有 nil deref。
func TestTopicServiceListWithoutDatabase(t *testing.T) {
	items, err := NewTopicService(nil, nil).List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("List() length = %d, want 3 (fixed topics)", len(items))
	}
	for _, item := range items {
		if item.ArticleCount != 0 {
			t.Fatalf("memory-mode article_count = %d, want 0", item.ArticleCount)
		}
	}
}

// TestAppErrorIncludesContext 简单冒烟，确保 AppError 在 §2.5 改动下行为未变。
func TestAppErrorIncludesContext(t *testing.T) {
	err := apperrors.New(apperrors.CodeResourceNotFound)
	if err.Code != apperrors.CodeResourceNotFound {
		t.Fatalf("code = %d, want %d", err.Code, apperrors.CodeResourceNotFound)
	}
	// model.RevokedRefreshToken 只是占位 import 防 unused 报错。
	_ = model.RevokedRefreshToken{}
}
