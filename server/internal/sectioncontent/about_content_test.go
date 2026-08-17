package sectioncontent

import (
	"strings"
	"testing"
)

func TestDefaultAboutContentIsCompleteAndIndependent(t *testing.T) {
	t.Parallel()

	first := DefaultAboutContent()
	second := DefaultAboutContent()
	for _, check := range []struct {
		name  string
		value string
	}{
		{"kicker", first.AboutKicker},
		{"heading", first.AboutHeading},
		{"lead", first.AboutLead},
		{"signature", first.AboutSignature},
		{"contact label", first.AboutContactLabel},
		{"reading label", first.AboutReadingLabel},
		{"principles kicker", first.AboutPrinciplesKicker},
		{"principles heading", first.AboutPrinciplesHeading},
		{"principles intro", first.AboutPrinciplesIntro},
		{"principle 1 title", first.Principle1Title},
		{"principle 1 description", first.Principle1Description},
		{"principle 2 title", first.Principle2Title},
		{"principle 2 description", first.Principle2Description},
		{"principle 3 title", first.Principle3Title},
		{"principle 3 description", first.Principle3Description},
		{"contact kicker", first.AboutContactKicker},
		{"contact heading with links", first.AboutContactHeadingWith},
		{"contact heading empty", first.AboutContactHeadingEmpty},
		{"contact intro with links", first.AboutContactIntroWith},
		{"contact intro empty", first.AboutContactIntroEmpty},
		{"contact empty cta", first.AboutContactEmptyCta},
		{"portrait line 1", first.AboutPortraitLine1},
		{"portrait line 2", first.AboutPortraitLine2},
	} {
		if check.value == "" {
			t.Fatalf("default about content has empty %s", check.name)
		}
	}
	first.AboutHeading = "changed"
	if second.AboutHeading == "changed" {
		t.Fatal("DefaultAboutContent returned shared mutable state")
	}
}

func TestIsValidAboutContent(t *testing.T) {
	t.Parallel()

	if !IsValidAboutContent(DefaultAboutContent()) {
		t.Fatal("shipped defaults must be valid")
	}
	if !IsValidAboutContent(AboutContent{}) {
		t.Fatal("empty values must be valid so they can reset to defaults")
	}

	overlong := DefaultAboutContent()
	overlong.AboutLead = strings.Repeat("文", AboutLeadMaxRunes+1)
	if IsValidAboutContent(overlong) {
		t.Fatal("overlong lead must be rejected")
	}
	overlongPrinciple := DefaultAboutContent()
	overlongPrinciple.Principle2Description = strings.Repeat("描", AboutPrincipleDescriptionMaxRunes+1)
	if IsValidAboutContent(overlongPrinciple) {
		t.Fatal("overlong principle description must be rejected")
	}
	overlongPortrait := DefaultAboutContent()
	overlongPortrait.AboutPortraitLine2 = strings.Repeat("题", AboutPortraitLine2MaxRunes+1)
	if IsValidAboutContent(overlongPortrait) {
		t.Fatal("overlong portrait line must be rejected")
	}
}

func TestNormalizeAboutContentTrimsAndFillsDefaults(t *testing.T) {
	t.Parallel()

	defaults := DefaultAboutContent()
	got := NormalizeAboutContent(AboutContent{
		AboutKicker:        "   ",
		AboutHeading:       "  自定义关于标题  ",
		Principle1Title:    "",
		AboutPortraitLine2: "    ",
	})

	if got.AboutKicker != defaults.AboutKicker {
		t.Fatalf("blank kicker = %q, want default", got.AboutKicker)
	}
	if got.AboutHeading != "自定义关于标题" {
		t.Fatalf("heading = %q, want trimmed custom value", got.AboutHeading)
	}
	if got.Principle1Title != defaults.Principle1Title {
		t.Fatalf("empty principle title = %q, want default", got.Principle1Title)
	}
	if got.AboutPortraitLine2 != defaults.AboutPortraitLine2 {
		t.Fatalf("blank portrait line = %q, want default", got.AboutPortraitLine2)
	}
}

func TestNormalizeAboutContentFallsBackFromOverlongValues(t *testing.T) {
	t.Parallel()

	defaults := DefaultAboutContent()
	got := NormalizeAboutContent(AboutContent{
		AboutSignature:          strings.Repeat("S", AboutSignatureMaxRunes+1),
		AboutContactHeadingWith: strings.Repeat("H", AboutContactHeadingMaxRunes+1),
		Principle3Title:         strings.Repeat("P", AboutPrincipleTitleMaxRunes+1),
	})

	if got.AboutSignature != defaults.AboutSignature {
		t.Fatalf("overlong signature = %q, want default", got.AboutSignature)
	}
	if got.AboutContactHeadingWith != defaults.AboutContactHeadingWith {
		t.Fatalf("overlong contact heading = %q, want default", got.AboutContactHeadingWith)
	}
	if got.Principle3Title != defaults.Principle3Title {
		t.Fatalf("overlong principle title = %q, want default", got.Principle3Title)
	}
}

func TestCloneAboutContentIsIndependent(t *testing.T) {
	t.Parallel()

	original := DefaultAboutContent()
	cloned := CloneAboutContent(original)
	cloned.AboutHeading = "changed"
	if original.AboutHeading == "changed" {
		t.Fatal("CloneAboutContent returned shared mutable state")
	}
}
