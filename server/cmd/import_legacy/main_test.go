package main

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"solitude-blog/server/internal/appearance"
)

func TestMergeImportedThemeElementsPreservesCurrentWhenOmitted(t *testing.T) {
	t.Parallel()

	current := appearance.DefaultThemeElements()
	current[appearance.ThemeMistForest] = appearance.ThemeElements{
		HomeLatestEmptyDescription: "保留的林间空状态",
		HomeLatestEndText:          "保留的林径终点",
	}
	got, err := mergeImportedThemeElements(mustMarshalThemeElements(t, current), nil)
	if err != nil {
		t.Fatalf("mergeImportedThemeElements returned error: %v", err)
	}
	if !reflect.DeepEqual(got, current) {
		t.Fatalf("merged theme elements = %#v, want preserved %#v", got, current)
	}
}

func TestMergeImportedThemeElementsUpdatesOnlyProvidedThemes(t *testing.T) {
	t.Parallel()

	current := appearance.DefaultThemeElements()
	current[appearance.ThemeMistForest] = appearance.ThemeElements{
		HomeLatestEmptyDescription: "保留的林间空状态",
		HomeLatestEndText:          "保留的林径终点",
	}
	incoming := appearance.ThemeElementMap{
		appearance.ThemeMistSeaSalt: {
			HomeLatestEmptyDescription: "",
			HomeLatestEndText:          "新的潮汐终点",
		},
	}
	got, err := mergeImportedThemeElements(mustMarshalThemeElements(t, current), &incoming)
	if err != nil {
		t.Fatalf("mergeImportedThemeElements returned error: %v", err)
	}

	defaults := appearance.DefaultThemeElements()
	seaSalt := got[appearance.ThemeMistSeaSalt]
	if seaSalt.HomeLatestEmptyDescription != defaults[appearance.ThemeMistSeaSalt].HomeLatestEmptyDescription {
		t.Fatalf("empty provided field = %q, want default", seaSalt.HomeLatestEmptyDescription)
	}
	if seaSalt.HomeLatestEndText != "新的潮汐终点" {
		t.Fatalf("provided end text = %q", seaSalt.HomeLatestEndText)
	}
	if got[appearance.ThemeMistForest] != current[appearance.ThemeMistForest] {
		t.Fatalf("omitted forest theme = %#v, want preserved %#v", got[appearance.ThemeMistForest], current[appearance.ThemeMistForest])
	}
}

func TestMergeImportedThemeElementsRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	tests := []appearance.ThemeElementMap{
		{"unknown-theme": {}},
		{appearance.ThemeMistSeaSalt: {
			HomeLatestEmptyDescription: strings.Repeat("潮", appearance.HomeLatestEmptyDescriptionMaxRunes+1),
		}},
		{appearance.ThemeMistForest: {
			HomeLatestEndText: strings.Repeat("林", appearance.HomeLatestEndTextMaxRunes+1),
		}},
	}
	for _, incoming := range tests {
		if _, err := mergeImportedThemeElements("", &incoming); err == nil {
			t.Fatalf("mergeImportedThemeElements(%#v) returned nil error", incoming)
		}
	}
}

func TestMergeImportedThemeElementsDefaultsMalformedCurrentJSON(t *testing.T) {
	t.Parallel()

	got, err := mergeImportedThemeElements("{not-json", nil)
	if err != nil {
		t.Fatalf("mergeImportedThemeElements returned error: %v", err)
	}
	if !reflect.DeepEqual(got, appearance.DefaultThemeElements()) {
		t.Fatalf("merged theme elements = %#v, want defaults", got)
	}
}

func mustMarshalThemeElements(t *testing.T, elements appearance.ThemeElementMap) string {
	t.Helper()
	payload, err := json.Marshal(elements)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}
	return string(payload)
}
