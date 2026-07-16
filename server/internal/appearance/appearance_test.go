package appearance

import (
	"strings"
	"testing"
)

func TestNormalizeTheme(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
		want  string
	}{
		{name: "sea salt", value: ThemeMistSeaSalt, want: ThemeMistSeaSalt},
		{name: "forest", value: ThemeMistForest, want: ThemeMistForest},
		{name: "legacy forest", value: "forest", want: ThemeMistForest},
		{name: "legacy violet", value: "mist-violet", want: ThemeMistSeaSalt},
		{name: "legacy strawberry", value: "strawberry", want: ThemeMistSeaSalt},
		{name: "empty", value: "", want: ThemeMistSeaSalt},
		{name: "unknown", value: "unknown", want: ThemeMistSeaSalt},
		{name: "normalized legacy spacing", value: "  FOREST  ", want: ThemeMistForest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := NormalizeTheme(tt.value); got != tt.want {
				t.Fatalf("NormalizeTheme(%q) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestThemeValidationRejectsLegacyAndUnknownValues(t *testing.T) {
	t.Parallel()

	for _, value := range []string{ThemeMistSeaSalt, ThemeMistForest} {
		if !IsValidTheme(value) {
			t.Errorf("IsValidTheme(%q) = false, want true", value)
		}
	}
	for _, value := range []string{"forest", "mist-violet", "strawberry", "", "unknown"} {
		if IsValidTheme(value) {
			t.Errorf("IsValidTheme(%q) = true, want false", value)
		}
	}
}

func TestNormalizeAndValidateMode(t *testing.T) {
	t.Parallel()

	if !IsValidMode(ModeLight) || !IsValidMode(ModeDark) {
		t.Fatal("light and dark must be valid modes")
	}
	for _, value := range []string{"", "system", "dark ", "LIGHT"} {
		if IsValidMode(value) {
			t.Errorf("IsValidMode(%q) = true, want false", value)
		}
	}

	tests := map[string]string{
		ModeLight: ModeLight,
		ModeDark:  ModeDark,
		" DARK ":  ModeDark,
		"":        ModeLight,
		"system":  ModeLight,
	}
	for value, want := range tests {
		if got := NormalizeMode(value); got != want {
			t.Errorf("NormalizeMode(%q) = %q, want %q", value, got, want)
		}
	}
}

func TestDefaultThemeElementsAreCompleteAndIndependent(t *testing.T) {
	t.Parallel()

	first := DefaultThemeElements()
	second := DefaultThemeElements()
	for _, theme := range []string{ThemeMistSeaSalt, ThemeMistForest} {
		element, ok := first[theme]
		if !ok {
			t.Fatalf("default theme elements missing %q", theme)
		}
		if element.HomeLatestEmptyDescription == "" || element.HomeLatestEndText == "" {
			t.Fatalf("default theme elements for %q are incomplete: %#v", theme, element)
		}
	}

	first[ThemeMistSeaSalt] = ThemeElements{HomeLatestEndText: "changed"}
	if second[ThemeMistSeaSalt].HomeLatestEndText == "changed" {
		t.Fatal("DefaultThemeElements returned shared mutable state")
	}
}

func TestNormalizeThemeElementsFillsPartialValues(t *testing.T) {
	t.Parallel()

	defaults := DefaultThemeElements()
	normalized := NormalizeThemeElements(ThemeElementMap{
		ThemeMistSeaSalt: {
			HomeLatestEmptyDescription: "   ",
			HomeLatestEndText:          "  自定义结束文案  ",
		},
		"unknown-theme": {
			HomeLatestEndText: "ignored",
		},
	})

	seaSalt := normalized[ThemeMistSeaSalt]
	if seaSalt.HomeLatestEmptyDescription != defaults[ThemeMistSeaSalt].HomeLatestEmptyDescription {
		t.Fatalf("empty description = %q, want default", seaSalt.HomeLatestEmptyDescription)
	}
	if seaSalt.HomeLatestEndText != "自定义结束文案" {
		t.Fatalf("end text = %q, want trimmed custom value", seaSalt.HomeLatestEndText)
	}
	if normalized[ThemeMistForest] != defaults[ThemeMistForest] {
		t.Fatalf("missing forest = %#v, want %#v", normalized[ThemeMistForest], defaults[ThemeMistForest])
	}
	if _, ok := normalized["unknown-theme"]; ok {
		t.Fatal("NormalizeThemeElements retained an unknown theme")
	}
}

func TestNormalizeThemeElementsFallsBackFromOverlongPersistedValues(t *testing.T) {
	t.Parallel()

	defaults := DefaultThemeElements()
	normalized := NormalizeThemeElements(ThemeElementMap{
		ThemeMistForest: {
			HomeLatestEmptyDescription: strings.Repeat("林", HomeLatestEmptyDescriptionMaxRunes+1),
			HomeLatestEndText:          strings.Repeat("径", HomeLatestEndTextMaxRunes+1),
		},
	})
	if normalized[ThemeMistForest] != defaults[ThemeMistForest] {
		t.Fatalf("overlong persisted values = %#v, want defaults %#v", normalized[ThemeMistForest], defaults[ThemeMistForest])
	}
}

func TestThemeElementValidation(t *testing.T) {
	t.Parallel()

	valid := ThemeElementMap{
		ThemeMistSeaSalt: {
			HomeLatestEmptyDescription: strings.Repeat("海", HomeLatestEmptyDescriptionMaxRunes),
			HomeLatestEndText:          strings.Repeat("潮", HomeLatestEndTextMaxRunes),
		},
	}
	if !IsValidThemeElements(valid) {
		t.Fatal("values at the rune limits must be valid")
	}
	if !IsValidThemeElements(ThemeElementMap{ThemeMistForest: {}}) {
		t.Fatal("empty values must be valid so they can reset to defaults")
	}

	invalid := []ThemeElementMap{
		{"unknown-theme": {}},
		{ThemeMistSeaSalt: {HomeLatestEmptyDescription: strings.Repeat("海", HomeLatestEmptyDescriptionMaxRunes+1)}},
		{ThemeMistForest: {HomeLatestEndText: strings.Repeat("林", HomeLatestEndTextMaxRunes+1)}},
	}
	for _, elements := range invalid {
		if IsValidThemeElements(elements) {
			t.Fatalf("IsValidThemeElements(%#v) = true, want false", elements)
		}
	}
}

func TestCloneThemeElementsDoesNotShareMap(t *testing.T) {
	t.Parallel()

	original := DefaultThemeElements()
	cloned := CloneThemeElements(original)
	cloned[ThemeMistForest] = ThemeElements{HomeLatestEndText: "changed"}
	if original[ThemeMistForest].HomeLatestEndText == "changed" {
		t.Fatal("CloneThemeElements returned shared mutable state")
	}
}
