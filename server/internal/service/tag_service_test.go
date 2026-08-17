package service

import (
	"strconv"
	"strings"
	"testing"
	"time"

	apperrors "solitude-blog/server/internal/errors"
)

func TestTagServiceRequiresNameAndSlug(t *testing.T) {
	service := NewTagService(nil, nil)

	_, err := service.Create(TagSaveRequest{Name: "", Slug: "x"})
	assertAppErrorCode(t, err, apperrors.CodeMissingRequiredField)
	_, err = service.Create(TagSaveRequest{Name: "x", Slug: ""})
	assertAppErrorCode(t, err, apperrors.CodeMissingRequiredField)
}

func TestTagServiceRejectsInvalidSlugFormat(t *testing.T) {
	service := NewTagService(nil, nil)

	cases := []string{"中文", "My-Slug", "my slug", "my_slug", "my.slug", "-abc", "abc-"}
	for _, slug := range cases {
		_, err := service.Create(TagSaveRequest{Name: "Tag", Slug: slug})
		assertAppErrorCode(t, err, apperrors.CodeInvalidParameter)
	}
}

func TestTagServiceRejectsOversizedSlug(t *testing.T) {
	service := NewTagService(nil, nil)

	_, err := service.Create(TagSaveRequest{Name: "Tag", Slug: strings.Repeat("a", 120)})
	if err != nil {
		t.Fatalf("Create() error = %v, want success at 120 chars", err)
	}
	_, err = service.Create(TagSaveRequest{Name: "Tag", Slug: strings.Repeat("a", 121)})
	assertAppErrorCode(t, err, apperrors.CodeInvalidParameter)
}

func TestTagServiceRejectsDuplicateSlugOnCreate(t *testing.T) {
	service := NewTagService(nil, nil)

	_, err := service.Create(TagSaveRequest{Name: "Notes", Slug: "notes"})
	if err != nil {
		t.Fatalf("first Create() error = %v", err)
	}
	_, err = service.Create(TagSaveRequest{Name: "Notes", Slug: "notes"})
	assertAppErrorCode(t, err, apperrors.CodeDuplicateSlug)
}

func TestTagServiceRejectsDuplicateSlugOnUpdate(t *testing.T) {
	service := NewTagService(nil, nil)

	_, err := service.Create(TagSaveRequest{Name: "Alpha", Slug: "alpha"})
	if err != nil {
		t.Fatalf("Create() alpha error = %v", err)
	}
	second, err := service.Create(TagSaveRequest{Name: "Beta", Slug: "beta"})
	if err != nil {
		t.Fatalf("Create() beta error = %v", err)
	}

	_, err = service.Update(strconv.FormatUint(second.ID, 10), TagSaveRequest{Name: "Beta", Slug: "alpha"})
	assertAppErrorCode(t, err, apperrors.CodeDuplicateSlug)

	updated, err := service.Update(strconv.FormatUint(second.ID, 10), TagSaveRequest{Name: "Beta Renamed", Slug: "beta"})
	if err != nil {
		t.Fatalf("Update() unchanged slug error = %v", err)
	}
	if updated.Name != "Beta Renamed" || updated.Slug != "beta" {
		t.Fatalf("Update() item = %#v", updated)
	}
}

func TestTagServiceUpdateRejectsInvalidSlugChange(t *testing.T) {
	service := NewTagService(nil, nil)

	item, err := service.Create(TagSaveRequest{Name: "Tag", Slug: "tag"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	_, err = service.Update(strconv.FormatUint(item.ID, 10), TagSaveRequest{Name: "Tag", Slug: "中文"})
	assertAppErrorCode(t, err, apperrors.CodeInvalidParameter)
}

func TestTagServiceUpdateSkipsSlugValidationWhenUnchanged(t *testing.T) {
	service := NewTagService(nil, nil)
	now := time.Now().UTC()
	// 注入存量非法 slug 行，模拟历史数据：slug 未变化时编辑其他字段不应被拦截。
	service.items = append(service.items, TagItem{
		ID:        1,
		Name:      "旧标签",
		Slug:      "旧-tag",
		CreatedAt: now,
		UpdatedAt: now,
	})

	updated, err := service.Update("1", TagSaveRequest{Name: "新名", Slug: "旧-tag"})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Name != "新名" || updated.Slug != "旧-tag" {
		t.Fatalf("Update() item = %#v, want name updated and slug kept", updated)
	}
}
