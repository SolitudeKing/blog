package model

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestDefaultTopicLabelTruncatesByRune(t *testing.T) {
	name := strings.Repeat("专题", 20)
	label := DefaultTopicLabel(name)

	if got := utf8.RuneCountInString(label); got != TopicLabelMaxRunes {
		t.Fatalf("DefaultTopicLabel() rune count = %d, want %d", got, TopicLabelMaxRunes)
	}
	if !strings.HasPrefix(name, label) {
		t.Fatalf("DefaultTopicLabel() = %q, want a rune-safe prefix of %q", label, name)
	}
}

func TestDefaultTopicLabelTrimsSpace(t *testing.T) {
	if got := DefaultTopicLabel("  Notes  "); got != "Notes" {
		t.Fatalf("DefaultTopicLabel() = %q, want Notes", got)
	}
}

func TestDefaultTopicLabelProvidesFallback(t *testing.T) {
	if got := DefaultTopicLabel("   "); got != "Topic" {
		t.Fatalf("DefaultTopicLabel() = %q, want Topic", got)
	}
}

func TestDefaultTopicsExposeStableCatalog(t *testing.T) {
	t.Parallel()

	topics := DefaultTopics()
	if len(topics) != 3 {
		t.Fatalf("DefaultTopics() length = %d, want 3", len(topics))
	}
	wants := []struct {
		name        string
		label       string
		slug        string
		description string
	}{
		{"雾里拾笺", TopicLabelNodes, TopicSlugNodes, "收拢阅读、学习与技术实践中散落的知识微光。"},
		{"微光造物", TopicLabelCode, TopicSlugCode, "记录灵感如何经由设计、代码与实验长成作品。"},
		{"风过留痕", TopicLabelJotting, TopicSlugJotting, "安放日常见闻、片刻心绪与未成体系的思考。"},
	}
	for index, want := range wants {
		got := topics[index]
		if got.Name != want.name || got.Label != want.label || got.Slug != want.slug || got.Description != want.description {
			t.Fatalf("DefaultTopics()[%d] = %#v, want %#v", index, got, want)
		}
	}
}

func TestDefaultTopicBySlug(t *testing.T) {
	t.Parallel()

	topic, ok := DefaultTopicBySlug(TopicSlugCode)
	if !ok || topic.Label != TopicLabelCode || topic.Name != "微光造物" {
		t.Fatalf("DefaultTopicBySlug(code) = %#v, %v", topic, ok)
	}
	if _, ok := DefaultTopicBySlug("unknown"); ok {
		t.Fatal("DefaultTopicBySlug(unknown) unexpectedly matched")
	}
}
