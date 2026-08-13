package service

import (
	"reflect"
	"testing"

	apperrors "solitude-blog/server/internal/errors"
	"solitude-blog/server/internal/model"
)

func TestArticleServiceStartsWithoutExampleArticles(t *testing.T) {
	service := NewArticleService(nil, nil)

	items, _, err := service.PublicList(ArticleListQuery{})
	if err != nil {
		t.Fatalf("PublicList() error = %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("PublicList() length = %d, want 0", len(items))
	}
}

func TestArticleServiceRequiresTopicAndActor(t *testing.T) {
	service := NewArticleService(nil, nil)

	_, err := service.Create(ArticleCreateRequest{Title: "文章", Slug: "article"}, 7)
	assertAppErrorCode(t, err, apperrors.CodeMissingRequiredField)

	_, err = service.Create(ArticleCreateRequest{Title: "文章", Slug: "article", TopicID: 1}, 0)
	assertAppErrorCode(t, err, apperrors.CodeUnauthorized)
}

func TestArticleServiceUsesCanonicalTopicAndDeduplicatesTagIDs(t *testing.T) {
	service := NewArticleService(nil, nil)

	item, err := service.Create(ArticleCreateRequest{
		Title:   "造物记录",
		Slug:    "making",
		TopicID: 2,
		TagIDs:  []uint64{3, 3, 5},
	}, 7)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if item.Topic.Label != model.TopicLabelCode || item.Topic.Slug != model.TopicSlugCode {
		t.Fatalf("Create() topic = %#v, want CODE topic", item.Topic)
	}
	detail, err := service.Info("1")
	if err != nil {
		t.Fatalf("Info() error = %v", err)
	}
	if !reflect.DeepEqual(detail.TagIDs, []uint64{3, 5}) {
		t.Fatalf("Info() tag_ids = %#v, want [3 5]", detail.TagIDs)
	}
}

func TestArticleServiceStoresArticleCoverURL(t *testing.T) {
	service := NewArticleService(nil, nil)

	item, err := service.Create(ArticleCreateRequest{
		Title:    "带预览图的文章",
		Slug:     "with-cover",
		TopicID:  1,
		CoverURL: "  /uploads/articles/cover.webp  ",
	}, 7)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if item.CoverURL != "/uploads/articles/cover.webp" {
		t.Fatalf("Create() cover_url = %q", item.CoverURL)
	}

	detail, err := service.Update("1", ArticleUpdateRequest{
		Title:    item.Title,
		Slug:     item.Slug,
		TopicID:  item.TopicID,
		CoverURL: "https://example.com/new-cover.webp",
	}, 7)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if detail.CoverURL != "https://example.com/new-cover.webp" {
		t.Fatalf("Update() cover_url = %q", detail.CoverURL)
	}
}

func assertAppErrorCode(t *testing.T, err error, code int) {
	t.Helper()
	appErr, ok := err.(apperrors.AppError)
	if !ok {
		t.Fatalf("error = %T %v, want AppError", err, err)
	}
	if appErr.Code != code {
		t.Fatalf("error code = %d, want %d", appErr.Code, code)
	}
}
