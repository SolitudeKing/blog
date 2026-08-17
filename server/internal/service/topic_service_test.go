package service

import (
	"strconv"
	"strings"
	"testing"
	"time"

	apperrors "solitude-blog/server/internal/errors"
)

func TestTopicServiceRequiresLabel(t *testing.T) {
	service := NewTopicService(nil, nil)

	_, err := service.Create(TopicSaveRequest{Name: "Engineering", Slug: "engineering"})
	if err == nil {
		t.Fatal("Create() error = nil, want missing required field error")
	}
}

func TestTopicServiceRejectsLongLabel(t *testing.T) {
	service := NewTopicService(nil, nil)

	_, err := service.Create(TopicSaveRequest{
		Name:  "Engineering",
		Label: strings.Repeat("专", 33),
		Slug:  "engineering",
	})
	if err == nil {
		t.Fatal("Create() error = nil, want invalid parameter error")
	}
}

func TestTopicServiceCreatesExpandedTopic(t *testing.T) {
	service := NewTopicService(nil, nil)

	item, err := service.Create(TopicSaveRequest{
		Name:        " Engineering ",
		Label:       " Build Log ",
		Slug:        " engineering ",
		Description: "Long-running engineering notes.",
		CoverURL:    "/uploads/topics/engineering.webp",
		SortOrder:   2,
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if item.Name != "Engineering" || item.Label != "Build Log" || item.Slug != "engineering" {
		t.Fatalf("Create() item = %#v, want trimmed required fields", item)
	}
	if item.CoverURL != "/uploads/topics/engineering.webp" {
		t.Fatalf("Create() cover_url = %q", item.CoverURL)
	}
}

func TestTopicServiceRejectsInvalidSlugFormat(t *testing.T) {
	service := NewTopicService(nil, nil)

	cases := []string{"工程", "My-Slug", "my slug", "my_slug", "-abc", "abc-"}
	for _, slug := range cases {
		_, err := service.Create(TopicSaveRequest{Name: "Engineering", Label: "L", Slug: slug})
		assertAppErrorCode(t, err, apperrors.CodeInvalidParameter)
	}
}

func TestTopicServiceRejectsOversizedSlug(t *testing.T) {
	service := NewTopicService(nil, nil)

	_, err := service.Create(TopicSaveRequest{Name: "Engineering", Label: "L", Slug: strings.Repeat("a", 120)})
	if err != nil {
		t.Fatalf("Create() error = %v, want success at 120 chars", err)
	}
	_, err = service.Create(TopicSaveRequest{Name: "Engineering", Label: "L", Slug: strings.Repeat("a", 121)})
	assertAppErrorCode(t, err, apperrors.CodeInvalidParameter)
}

func TestTopicServiceRejectsDuplicateSlugOnCreate(t *testing.T) {
	service := NewTopicService(nil, nil)

	_, err := service.Create(TopicSaveRequest{Name: "Engineering", Label: "L", Slug: "engineering"})
	if err != nil {
		t.Fatalf("first Create() error = %v", err)
	}
	_, err = service.Create(TopicSaveRequest{Name: "Engineering", Label: "L", Slug: "engineering"})
	assertAppErrorCode(t, err, apperrors.CodeDuplicateSlug)
}

func TestTopicServiceRejectsDuplicateSlugOnUpdate(t *testing.T) {
	service := NewTopicService(nil, nil)

	_, err := service.Create(TopicSaveRequest{Name: "First", Label: "F", Slug: "first-a"})
	if err != nil {
		t.Fatalf("Create() first error = %v", err)
	}
	second, err := service.Create(TopicSaveRequest{Name: "Second", Label: "S", Slug: "second-b"})
	if err != nil {
		t.Fatalf("Create() second error = %v", err)
	}

	_, err = service.Update(strconv.FormatUint(second.ID, 10), TopicSaveRequest{Name: "Second", Label: "S", Slug: "first-a"})
	assertAppErrorCode(t, err, apperrors.CodeDuplicateSlug)

	updated, err := service.Update(strconv.FormatUint(second.ID, 10), TopicSaveRequest{Name: "Second Renamed", Label: "S", Slug: "second-b"})
	if err != nil {
		t.Fatalf("Update() unchanged slug error = %v", err)
	}
	if updated.Name != "Second Renamed" || updated.Slug != "second-b" {
		t.Fatalf("Update() item = %#v", updated)
	}
}

func TestTopicServiceUpdateRejectsInvalidSlugChange(t *testing.T) {
	service := NewTopicService(nil, nil)

	item, err := service.Create(TopicSaveRequest{Name: "Engineering", Label: "L", Slug: "engineering"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	_, err = service.Update(strconv.FormatUint(item.ID, 10), TopicSaveRequest{Name: "Engineering", Label: "L", Slug: "工程"})
	assertAppErrorCode(t, err, apperrors.CodeInvalidParameter)
}

func TestTopicServiceUpdateSkipsSlugValidationWhenUnchanged(t *testing.T) {
	service := NewTopicService(nil, nil)
	now := time.Now().UTC()
	// 注入存量非法 slug 行，模拟历史数据：slug 未变化时编辑其他字段不应被拦截。
	service.items = append(service.items, TopicItem{
		ID:        99,
		Name:      "旧专题",
		Label:     "OLD",
		Slug:      "旧-slug",
		CreatedAt: now,
		UpdatedAt: now,
	})

	updated, err := service.Update("99", TopicSaveRequest{Name: "新名", Label: "OLD", Slug: "旧-slug"})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Name != "新名" || updated.Slug != "旧-slug" {
		t.Fatalf("Update() item = %#v, want name updated and slug kept", updated)
	}
}
