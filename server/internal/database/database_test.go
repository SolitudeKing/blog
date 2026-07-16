package database

import (
	"encoding/json"
	"reflect"
	"testing"

	"solitude-blog/server/internal/appearance"
)

func TestNormalizeStoredThemeElementsBackfillsEmptyAndMalformedJSON(t *testing.T) {
	t.Parallel()

	for _, raw := range []string{"", "{not-json"} {
		payload, changed, err := normalizeStoredThemeElements(raw)
		if err != nil {
			t.Fatalf("normalizeStoredThemeElements(%q) returned error: %v", raw, err)
		}
		if !changed {
			t.Fatalf("normalizeStoredThemeElements(%q) changed = false, want true", raw)
		}
		if got := decodeThemeElements(t, payload); !reflect.DeepEqual(got, appearance.DefaultThemeElements()) {
			t.Fatalf("normalizeStoredThemeElements(%q) = %#v, want defaults", raw, got)
		}
	}
}

func TestNormalizeStoredThemeElementsCompletesPartialJSON(t *testing.T) {
	t.Parallel()

	partial := appearance.ThemeElementMap{
		appearance.ThemeMistSeaSalt: {
			HomeLatestEndText: "保留的潮汐终点",
		},
	}
	payload, changed, err := normalizeStoredThemeElements(mustMarshalThemeElements(t, partial))
	if err != nil {
		t.Fatalf("normalizeStoredThemeElements returned error: %v", err)
	}
	if !changed {
		t.Fatal("partial theme elements changed = false, want true")
	}

	got := decodeThemeElements(t, payload)
	want := appearance.DefaultThemeElements()
	seaSalt := want[appearance.ThemeMistSeaSalt]
	seaSalt.HomeLatestEndText = "保留的潮汐终点"
	want[appearance.ThemeMistSeaSalt] = seaSalt
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("normalized partial theme elements = %#v, want %#v", got, want)
	}
}

func TestNormalizeStoredThemeElementsLeavesCompleteValidJSONUntouched(t *testing.T) {
	t.Parallel()

	complete := appearance.DefaultThemeElements()
	complete[appearance.ThemeMistForest] = appearance.ThemeElements{
		HomeLatestEmptyDescription: "自定义林间空状态",
		HomeLatestEndText:          "自定义林径终点",
	}
	raw := mustMarshalThemeElements(t, complete)
	payload, changed, err := normalizeStoredThemeElements(raw)
	if err != nil {
		t.Fatalf("normalizeStoredThemeElements returned error: %v", err)
	}
	if changed {
		t.Fatal("complete valid theme elements changed = true, want false")
	}
	if payload != raw {
		t.Fatalf("payload = %q, want original %q", payload, raw)
	}
}

func decodeThemeElements(t *testing.T, payload string) appearance.ThemeElementMap {
	t.Helper()
	var elements appearance.ThemeElementMap
	if err := json.Unmarshal([]byte(payload), &elements); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}
	return elements
}

func mustMarshalThemeElements(t *testing.T, elements appearance.ThemeElementMap) string {
	t.Helper()
	payload, err := json.Marshal(elements)
	if err != nil {
		t.Fatalf("json.Marshal returned error: %v", err)
	}
	return string(payload)
}
