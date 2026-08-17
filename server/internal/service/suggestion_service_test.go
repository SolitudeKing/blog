package service

import (
	"reflect"
	"testing"
)

func TestBuildSuggestionsIsDeterministicForSeed(t *testing.T) {
	t.Parallel()

	topics := []string{"雾里拾笺", "微光造物", "风过留痕", "拾遗集"}
	tags := []string{"Vue", "设计系统", "写作", "TypeScript"}

	first := buildSuggestions(topics, tags, seedForDate("2026-08-17"))
	second := buildSuggestions(topics, tags, seedForDate("2026-08-17"))
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("same date produced different suggestions: %#v vs %#v", first, second)
	}
	if len(first) != suggestionTotal {
		t.Fatalf("suggestion count = %d, want %d", len(first), suggestionTotal)
	}
}

func TestBuildSuggestionsRespectsCapsAndFillsShortfall(t *testing.T) {
	t.Parallel()

	// 专题远多于上限：专题取满 4（优先），标签填余量至总量 6。
	manyTopics := []string{"A", "B", "C", "D", "E", "F"}
	manyTags := []string{"G", "H", "I", "J"}
	items := buildSuggestions(manyTopics, manyTags, 42)
	if len(items) != suggestionTotal {
		t.Fatalf("suggestion count = %d, want %d", len(items), suggestionTotal)
	}
	topicCount := 0
	tagCount := 0
	for _, item := range items {
		switch item.Kind {
		case "topic":
			topicCount++
		case "tag":
			tagCount++
		}
	}
	if topicCount != suggestionTopicCap {
		t.Fatalf("topic count = %d, want %d", topicCount, suggestionTopicCap)
	}
	if tagCount != suggestionTotal-suggestionTopicCap {
		t.Fatalf("tag count = %d, want %d", tagCount, suggestionTotal-suggestionTopicCap)
	}

	// 标签不足时从专题池补齐，总量仍为 6。
	fewTags := buildSuggestions(manyTopics, []string{"G"}, 42)
	if len(fewTags) != suggestionTotal {
		t.Fatalf("suggestion count with few tags = %d, want %d", len(fewTags), suggestionTotal)
	}
	if got := kindCount(fewTags, "topic"); got != suggestionTotal-1 {
		t.Fatalf("topic count with few tags = %d, want %d", got, suggestionTotal-1)
	}

	// 专题不足时从标签池补齐；名字总量只有 5，结果应耗尽两池。
	fewTopics := buildSuggestions([]string{"A"}, manyTags, 42)
	if len(fewTopics) != 5 {
		t.Fatalf("suggestion count with few topics = %d, want 5", len(fewTopics))
	}
	if got := kindCount(fewTopics, "topic"); got != 1 {
		t.Fatalf("topic count with few topics = %d, want 1", got)
	}
	if got := kindCount(fewTopics, "tag"); got != 4 {
		t.Fatalf("tag count with few topics = %d, want 4", got)
	}
}

func TestBuildSuggestionsDeduplicatesAcrossPools(t *testing.T) {
	t.Parallel()

	// 专题与标签同名时保留专题，且结果内无重复文本。
	topics := []string{"Vue", "A", "B", "C"}
	tags := []string{"Vue", "Vue", "D", "E"}
	items := buildSuggestions(topics, tags, 7)
	seen := map[string]int{}
	for _, item := range items {
		seen[item.Text]++
		if item.Text == "Vue" && item.Kind != "topic" {
			t.Fatalf("duplicate text %q was kept as kind %q", item.Text, item.Kind)
		}
		if seen[item.Text] > 1 {
			t.Fatalf("duplicate suggestion text %q in %#v", item.Text, items)
		}
	}
}

func TestBuildSuggestionsEmptyPoolsReturnEmptySlice(t *testing.T) {
	t.Parallel()

	items := buildSuggestions(nil, nil, 1)
	if items == nil {
		t.Fatal("buildSuggestions returned nil, want empty slice")
	}
	if len(items) != 0 {
		t.Fatalf("suggestion count = %d, want 0", len(items))
	}
}

func TestSuggestionsSameDayIsStableWithInMemoryTaxonomy(t *testing.T) {
	t.Parallel()

	topics := NewTopicService(nil, nil)
	tags := NewTagService(nil, nil)
	service := NewSuggestionService(topics, tags, nil)

	first, err := service.Suggestions()
	if err != nil {
		t.Fatalf("Suggestions returned error: %v", err)
	}
	second, err := service.Suggestions()
	if err != nil {
		t.Fatalf("Suggestions returned error: %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("same-day Suggestions differ: %#v vs %#v", first, second)
	}
	// 内存模式下只有 3 个种子专题、没有标签，结果是专题子集，不保证凑满 6。
	if len(first) == 0 {
		t.Fatal("Suggestions returned empty result for seeded topics")
	}
	for _, item := range first {
		if item.Kind != "topic" {
			t.Fatalf("in-memory suggestion kind = %q, want topic", item.Kind)
		}
	}
}

func kindCount(items []SuggestionItem, kind string) int {
	count := 0
	for _, item := range items {
		if item.Kind == kind {
			count++
		}
	}
	return count
}
