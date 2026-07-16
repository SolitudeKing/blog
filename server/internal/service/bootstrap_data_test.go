package service

import (
	"testing"

	"solitude-blog/server/internal/model"
)

func TestInMemoryCatalogContainsOnlyFormalTopics(t *testing.T) {
	service := NewTopicService(nil, nil)
	items, err := service.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("List() length = %d, want 3", len(items))
	}
	labels := []string{model.TopicLabelNodes, model.TopicLabelCode, model.TopicLabelJotting}
	for index, label := range labels {
		if items[index].Label != label {
			t.Fatalf("List()[%d].Label = %q, want %q", index, items[index].Label, label)
		}
	}
}

func TestInMemoryExampleCollectionsStartEmpty(t *testing.T) {
	tags, err := NewTagService(nil, nil).List()
	if err != nil || len(tags) != 0 {
		t.Fatalf("tag List() = %#v, %v; want empty", tags, err)
	}

	notice, err := NewNoticeService(nil).Active()
	if err != nil || notice != nil {
		t.Fatalf("notice Active() = %#v, %v; want nil", notice, err)
	}

	summary, err := NewDashboardService(nil).Summary()
	if err != nil {
		t.Fatalf("dashboard Summary() error = %v", err)
	}
	if summary.ArticleCounts.Total != 0 || summary.TaxonomyCounts.Topics != 0 || len(summary.RecentArticles) != 0 || summary.ActiveNotice != nil {
		t.Fatalf("dashboard Summary() = %#v, want zero data", summary)
	}
}

func TestFormalTopicIdentityCannotDriftOrBeDeleted(t *testing.T) {
	service := NewTopicService(nil, nil)

	_, err := service.Update("1", TopicSaveRequest{
		Name:      "仍可维护展示名",
		Label:     model.TopicLabelNodes,
		Slug:      "renamed-nodes",
		SortOrder: 1,
	})
	if err == nil {
		t.Fatal("Update() allowed a formal topic slug to drift")
	}
	if err := service.Delete("1"); err == nil {
		t.Fatal("Delete() allowed a formal topic to be deleted")
	}
}
