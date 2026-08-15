package homecontent

import (
	"strings"
	"testing"
)

func TestDefaultHomeContentIsCompleteAndIndependent(t *testing.T) {
	t.Parallel()

	first := DefaultHomeContent()
	second := DefaultHomeContent()
	for _, check := range []struct {
		name  string
		value string
	}{
		{"profile kicker", first.HomeProfileKicker},
		{"heading prefix", first.HomeHeadingPrefix},
		{"status fallback", first.HomeStatusFallback},
		{"intro heading", first.HomeIntroHeading},
		{"intro paragraph", first.HomeIntroParagraph},
		{"action view recent", first.HomeActionViewRecentLabel},
		{"action view archive", first.HomeActionViewArchiveLabel},
		{"latest kicker", first.HomeLatestKicker},
		{"latest heading", first.HomeLatestHeading},
		{"latest view all", first.HomeLatestViewAllLabel},
		{"latest empty title", first.HomeLatestEmptyTitle},
	} {
		if check.value == "" {
			t.Fatalf("default home content has empty %s", check.name)
		}
	}
	first.HomeIntroHeading = "changed"
	if second.HomeIntroHeading == "changed" {
		t.Fatal("DefaultHomeContent returned shared mutable state")
	}
}

func TestIsValidHomeContent(t *testing.T) {
	t.Parallel()

	if !IsValidHomeContent(DefaultHomeContent()) {
		t.Fatal("shipped defaults must be valid")
	}
	if !IsValidHomeContent(HomeContent{}) {
		t.Fatal("empty values must be valid so they can reset to defaults")
	}

	overlong := DefaultHomeContent()
	overlong.HomeIntroParagraph = strings.Repeat("文", HomeIntroParagraphMaxRunes+1)
	if IsValidHomeContent(overlong) {
		t.Fatal("overlong intro paragraph must be rejected")
	}
	overlongAction := DefaultHomeContent()
	overlongAction.HomeActionViewRecentLabel = strings.Repeat("查", HomeActionLabelMaxRunes+1)
	if IsValidHomeContent(overlongAction) {
		t.Fatal("overlong action label must be rejected")
	}
}

func TestNormalizeHomeContentTrimsAndFillsDefaults(t *testing.T) {
	t.Parallel()

	defaults := DefaultHomeContent()
	got := NormalizeHomeContent(HomeContent{
		HomeProfileKicker:          "   ",
		HomeIntroHeading:           "  自定义介绍标题  ",
		HomeActionViewRecentLabel:  "",
		HomeLatestEmptyTitle:       "    ",
	})

	if got.HomeProfileKicker != defaults.HomeProfileKicker {
		t.Fatalf("blank profile kicker = %q, want default", got.HomeProfileKicker)
	}
	if got.HomeIntroHeading != "自定义介绍标题" {
		t.Fatalf("intro heading = %q, want trimmed custom value", got.HomeIntroHeading)
	}
	if got.HomeActionViewRecentLabel != defaults.HomeActionViewRecentLabel {
		t.Fatalf("empty action label = %q, want default", got.HomeActionViewRecentLabel)
	}
	if got.HomeLatestEmptyTitle != defaults.HomeLatestEmptyTitle {
		t.Fatalf("blank latest empty title = %q, want default", got.HomeLatestEmptyTitle)
	}
}

func TestNormalizeHomeContentFallsBackFromOverlongValues(t *testing.T) {
	t.Parallel()

	defaults := DefaultHomeContent()
	got := NormalizeHomeContent(HomeContent{
		HomeProfileKicker:  strings.Repeat("B", HomeProfileKickerMaxRunes+1),
		HomeIntroHeading:   strings.Repeat("H", HomeIntroHeadingMaxRunes+1),
		HomeLatestKicker:   strings.Repeat("K", HomeLatestKickerMaxRunes+1),
		HomeLatestHeading:  strings.Repeat("T", HomeLatestHeadingMaxRunes+1),
	})

	if got.HomeProfileKicker != defaults.HomeProfileKicker {
		t.Fatalf("overlong profile kicker = %q, want default", got.HomeProfileKicker)
	}
	if got.HomeIntroHeading != defaults.HomeIntroHeading {
		t.Fatalf("overlong intro heading = %q, want default", got.HomeIntroHeading)
	}
	if got.HomeLatestKicker != defaults.HomeLatestKicker {
		t.Fatalf("overlong latest kicker = %q, want default", got.HomeLatestKicker)
	}
	if got.HomeLatestHeading != defaults.HomeLatestHeading {
		t.Fatalf("overlong latest heading = %q, want default", got.HomeLatestHeading)
	}
}

func TestCloneHomeContentIsIndependent(t *testing.T) {
	t.Parallel()

	original := DefaultHomeContent()
	cloned := CloneHomeContent(original)
	cloned.HomeIntroHeading = "changed"
	if original.HomeIntroHeading == "changed" {
		t.Fatal("CloneHomeContent returned shared mutable state")
	}
}
