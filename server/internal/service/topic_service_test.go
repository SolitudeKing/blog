package service

import (
	"strings"
	"testing"
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
