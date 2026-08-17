package sectioncontent

import (
	"strings"
	"testing"
)

func TestDefaultSearchContentIsCompleteAndIndependent(t *testing.T) {
	t.Parallel()

	first := DefaultSearchContent()
	second := DefaultSearchContent()
	for _, check := range []struct {
		name  string
		value string
	}{
		{"kicker", first.SearchKicker},
		{"heading", first.SearchHeading},
		{"intro", first.SearchIntro},
		{"placeholder", first.SearchPlaceholder},
		{"suggestion label", first.SearchSuggestionLabel},
		{"suggestion fallbacks", first.SearchSuggestionFallbacks},
		{"empty title", first.SearchEmptyTitle},
		{"empty description", first.SearchEmptyDescription},
	} {
		if check.value == "" {
			t.Fatalf("default search content has empty %s", check.name)
		}
	}
	first.SearchHeading = "changed"
	if second.SearchHeading == "changed" {
		t.Fatal("DefaultSearchContent returned shared mutable state")
	}
}

func TestIsValidSearchContent(t *testing.T) {
	t.Parallel()

	if !IsValidSearchContent(DefaultSearchContent()) {
		t.Fatal("shipped defaults must be valid")
	}
	if !IsValidSearchContent(SearchContent{}) {
		t.Fatal("empty values must be valid so they can reset to defaults")
	}

	overlong := DefaultSearchContent()
	overlong.SearchIntro = strings.Repeat("文", SearchIntroMaxRunes+1)
	if IsValidSearchContent(overlong) {
		t.Fatal("overlong intro must be rejected")
	}
	overlongFallbacks := DefaultSearchContent()
	overlongFallbacks.SearchSuggestionFallbacks = strings.Repeat("词", SearchSuggestionFallbacksMaxRunes+1)
	if IsValidSearchContent(overlongFallbacks) {
		t.Fatal("overlong suggestion fallbacks must be rejected")
	}
}

func TestNormalizeSearchContentTrimsAndFillsDefaults(t *testing.T) {
	t.Parallel()

	defaults := DefaultSearchContent()
	got := NormalizeSearchContent(SearchContent{
		SearchKicker:      "   ",
		SearchHeading:     "  自定义搜索标题  ",
		SearchPlaceholder: "",
	})

	if got.SearchKicker != defaults.SearchKicker {
		t.Fatalf("blank kicker = %q, want default", got.SearchKicker)
	}
	if got.SearchHeading != "自定义搜索标题" {
		t.Fatalf("heading = %q, want trimmed custom value", got.SearchHeading)
	}
	if got.SearchPlaceholder != defaults.SearchPlaceholder {
		t.Fatalf("empty placeholder = %q, want default", got.SearchPlaceholder)
	}
}

func TestNormalizeSearchContentFallsBackFromOverlongValues(t *testing.T) {
	t.Parallel()

	defaults := DefaultSearchContent()
	got := NormalizeSearchContent(SearchContent{
		SearchKicker:          strings.Repeat("S", SearchKickerMaxRunes+1),
		SearchEmptyTitle:      strings.Repeat("T", SearchEmptyTitleMaxRunes+1),
		SearchSuggestionLabel: strings.Repeat("L", SearchSuggestionLabelMaxRunes+1),
	})

	if got.SearchKicker != defaults.SearchKicker {
		t.Fatalf("overlong kicker = %q, want default", got.SearchKicker)
	}
	if got.SearchEmptyTitle != defaults.SearchEmptyTitle {
		t.Fatalf("overlong empty title = %q, want default", got.SearchEmptyTitle)
	}
	if got.SearchSuggestionLabel != defaults.SearchSuggestionLabel {
		t.Fatalf("overlong suggestion label = %q, want default", got.SearchSuggestionLabel)
	}
}

func TestCloneSearchContentIsIndependent(t *testing.T) {
	t.Parallel()

	original := DefaultSearchContent()
	cloned := CloneSearchContent(original)
	cloned.SearchHeading = "changed"
	if original.SearchHeading == "changed" {
		t.Fatal("CloneSearchContent returned shared mutable state")
	}
}
