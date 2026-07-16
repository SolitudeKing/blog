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
