package sectioncontent

import (
	"strings"
	"testing"
)

func TestDefaultArchiveContentIsCompleteAndIndependent(t *testing.T) {
	t.Parallel()

	first := DefaultArchiveContent()
	second := DefaultArchiveContent()
	for _, check := range []struct {
		name  string
		value string
	}{
		{"kicker", first.ArchiveKicker},
		{"heading", first.ArchiveHeading},
		{"intro", first.ArchiveIntro},
		{"empty title", first.ArchiveEmptyTitle},
		{"empty description", first.ArchiveEmptyDescription},
	} {
		if check.value == "" {
			t.Fatalf("default archive content has empty %s", check.name)
		}
	}
	first.ArchiveHeading = "changed"
	if second.ArchiveHeading == "changed" {
		t.Fatal("DefaultArchiveContent returned shared mutable state")
	}
}

func TestIsValidArchiveContent(t *testing.T) {
	t.Parallel()

	if !IsValidArchiveContent(DefaultArchiveContent()) {
		t.Fatal("shipped defaults must be valid")
	}
	if !IsValidArchiveContent(ArchiveContent{}) {
		t.Fatal("empty values must be valid so they can reset to defaults")
	}

	overlong := DefaultArchiveContent()
	overlong.ArchiveIntro = strings.Repeat("文", ArchiveIntroMaxRunes+1)
	if IsValidArchiveContent(overlong) {
		t.Fatal("overlong intro must be rejected")
	}
	overlongTitle := DefaultArchiveContent()
	overlongTitle.ArchiveEmptyTitle = strings.Repeat("题", ArchiveEmptyTitleMaxRunes+1)
	if IsValidArchiveContent(overlongTitle) {
		t.Fatal("overlong empty title must be rejected")
	}
}

func TestNormalizeArchiveContentTrimsAndFillsDefaults(t *testing.T) {
	t.Parallel()

	defaults := DefaultArchiveContent()
	got := NormalizeArchiveContent(ArchiveContent{
		ArchiveKicker:     "   ",
		ArchiveHeading:    "  自定义归档标题  ",
		ArchiveEmptyTitle: "",
	})

	if got.ArchiveKicker != defaults.ArchiveKicker {
		t.Fatalf("blank kicker = %q, want default", got.ArchiveKicker)
	}
	if got.ArchiveHeading != "自定义归档标题" {
		t.Fatalf("heading = %q, want trimmed custom value", got.ArchiveHeading)
	}
	if got.ArchiveEmptyTitle != defaults.ArchiveEmptyTitle {
		t.Fatalf("empty empty-title = %q, want default", got.ArchiveEmptyTitle)
	}
}

func TestNormalizeArchiveContentFallsBackFromOverlongValues(t *testing.T) {
	t.Parallel()

	defaults := DefaultArchiveContent()
	got := NormalizeArchiveContent(ArchiveContent{
		ArchiveKicker:           strings.Repeat("A", ArchiveKickerMaxRunes+1),
		ArchiveEmptyDescription: strings.Repeat("D", ArchiveEmptyDescriptionMaxRunes+1),
	})

	if got.ArchiveKicker != defaults.ArchiveKicker {
		t.Fatalf("overlong kicker = %q, want default", got.ArchiveKicker)
	}
	if got.ArchiveEmptyDescription != defaults.ArchiveEmptyDescription {
		t.Fatalf("overlong empty description = %q, want default", got.ArchiveEmptyDescription)
	}
}

func TestCloneArchiveContentIsIndependent(t *testing.T) {
	t.Parallel()

	original := DefaultArchiveContent()
	cloned := CloneArchiveContent(original)
	cloned.ArchiveHeading = "changed"
	if original.ArchiveHeading == "changed" {
		t.Fatal("CloneArchiveContent returned shared mutable state")
	}
}
