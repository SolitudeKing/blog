package database

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"gorm.io/gorm"

	"solitude-blog/server/internal/appearance"
	"solitude-blog/server/internal/config"
	"solitude-blog/server/internal/model"
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

func TestOpenMySQLRejectsEmptyDSN(t *testing.T) {
	t.Parallel()

	if _, err := openMySQL(context.Background(), config.Config{}); err == nil {
		t.Fatal("openMySQL() error = nil, want missing DSN error")
	}
}

func TestDefaultTopicRestoreDoesNotOverwriteMaintainedFields(t *testing.T) {
	t.Parallel()

	topic := model.Topic{
		Name:        "自定义专题名",
		Label:       "CUSTOM",
		Slug:        model.TopicSlugNodes,
		Description: "后台维护的描述",
		SortOrder:   99,
	}
	if updates := defaultTopicRestoreUpdates(topic); len(updates) != 0 {
		t.Fatalf("active topic updates = %#v, want none", updates)
	}

	topic.DeletedAt = gorm.DeletedAt{Valid: true}
	updates := defaultTopicRestoreUpdates(topic)
	if len(updates) != 1 {
		t.Fatalf("deleted topic updates = %#v, want deleted_at only", updates)
	}
	if value, ok := updates["deleted_at"]; !ok || value != nil {
		t.Fatalf("deleted_at update = %#v, want nil", value)
	}
}

func TestExactDefaultNotesCategoryFingerprint(t *testing.T) {
	t.Parallel()

	exact := legacyCategoryRow{
		Name:      "Notes",
		Slug:      "notes",
		SortOrder: 1,
	}
	if !isExactDefaultNotesCategory(exact) {
		t.Fatal("exact default Notes category was not recognized")
	}

	cases := map[string]legacyCategoryRow{
		"custom name":        {Name: "我的笔记", Slug: "notes", SortOrder: 1},
		"custom slug":        {Name: "Notes", Slug: "notes-2", SortOrder: 1},
		"custom description": {Name: "Notes", Slug: "notes", Description: "自定义说明", SortOrder: 1},
		"custom order":       {Name: "Notes", Slug: "notes", SortOrder: 8},
		"deleted row": {
			Name:      "Notes",
			Slug:      "notes",
			SortOrder: 1,
			DeletedAt: gorm.DeletedAt{Valid: true},
		},
	}
	for name, category := range cases {
		category := category
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if isExactDefaultNotesCategory(category) {
				t.Fatalf("custom category %#v was treated as the scaffold default", category)
			}
		})
	}
}

func TestExactDefaultNotesTopicFingerprint(t *testing.T) {
	t.Parallel()

	exact := model.Topic{Name: "Notes", Label: "Notes", Slug: "notes", SortOrder: 1}
	if !isExactDefaultNotesTopic(exact) {
		t.Fatal("exact default Notes topic was not recognized")
	}

	customized := exact
	customized.Description = "管理员维护过的说明"
	if isExactDefaultNotesTopic(customized) {
		t.Fatal("customized Notes topic was treated as the scaffold default")
	}
	customized = exact
	customized.SortOrder = 9
	if isExactDefaultNotesTopic(customized) {
		t.Fatal("reordered Notes topic was treated as the scaffold default")
	}
}
