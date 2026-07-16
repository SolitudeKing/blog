package service

import (
	"reflect"
	"strings"
	"testing"

	"solitude-blog/server/internal/appearance"
	apperrors "solitude-blog/server/internal/errors"
	"solitude-blog/server/internal/model"
)

func TestNormalizeSettingAcceptsSupportedAppearance(t *testing.T) {
	t.Parallel()

	for _, theme := range []string{appearance.ThemeMistSeaSalt, appearance.ThemeMistForest} {
		for _, mode := range []string{appearance.ModeLight, appearance.ModeDark} {
			req := SettingSaveRequest{
				SiteName: "Solitude Blog",
				Author:   "Solitude King",
				Theme:    theme,
				Mode:     mode,
			}
			got, err := normalizeSetting(req, nil)
			if err != nil {
				t.Fatalf("normalizeSetting(%q, %q) returned error: %v", theme, mode, err)
			}
			if got.Theme != theme || got.Mode != mode {
				t.Fatalf("normalizeSetting(%q, %q) = (%q, %q)", theme, mode, got.Theme, got.Mode)
			}
			if got.SocialLinks == nil {
				t.Fatal("normalizeSetting must initialize social links")
			}
		}
	}
}

func TestNormalizeSettingRejectsInvalidAppearance(t *testing.T) {
	t.Parallel()

	tests := []SettingSaveRequest{
		{SiteName: "Blog", Author: "Author", Theme: "forest", Mode: appearance.ModeLight},
		{SiteName: "Blog", Author: "Author", Theme: "mist-violet", Mode: appearance.ModeLight},
		{SiteName: "Blog", Author: "Author", Theme: "", Mode: appearance.ModeLight},
		{SiteName: "Blog", Author: "Author", Theme: appearance.ThemeMistSeaSalt, Mode: "system"},
		{SiteName: "Blog", Author: "Author", Theme: appearance.ThemeMistSeaSalt, Mode: ""},
	}

	for _, req := range tests {
		_, err := normalizeSetting(req, nil)
		if err == nil {
			t.Fatalf("normalizeSetting(%q, %q) returned nil error", req.Theme, req.Mode)
		}
		appErr, ok := err.(apperrors.AppError)
		if !ok || appErr.Code != apperrors.CodeInvalidParameter {
			t.Fatalf("normalizeSetting(%q, %q) error = %#v, want invalid parameter", req.Theme, req.Mode, err)
		}
	}
}

func TestSettingFromModelNormalizesLegacyAppearance(t *testing.T) {
	t.Parallel()

	item := settingFromModel(model.SiteSetting{
		SiteName: "Blog",
		Author:   "Author",
		Theme:    "forest",
		Mode:     "system",
	})

	if item.Theme != appearance.ThemeMistForest {
		t.Fatalf("theme = %q, want %q", item.Theme, appearance.ThemeMistForest)
	}
	if item.Mode != appearance.DefaultMode {
		t.Fatalf("mode = %q, want %q", item.Mode, appearance.DefaultMode)
	}
	if !reflect.DeepEqual(item.ThemeElements, appearance.DefaultThemeElements()) {
		t.Fatalf("theme elements = %#v, want defaults", item.ThemeElements)
	}
}

func TestNormalizeSettingMergesProvidedThemesAndResetsEmptyFields(t *testing.T) {
	t.Parallel()

	current := appearance.DefaultThemeElements()
	current[appearance.ThemeMistForest] = appearance.ThemeElements{
		HomeLatestEmptyDescription: "保留的青森空状态",
		HomeLatestEndText:          "保留的青森结束文案",
	}
	provided := appearance.ThemeElementMap{
		appearance.ThemeMistSeaSalt: {
			HomeLatestEmptyDescription: "",
			HomeLatestEndText:          "新的海盐结束文案",
		},
	}
	got, err := normalizeSetting(SettingSaveRequest{
		SiteName:      "Blog",
		Author:        "Author",
		Theme:         appearance.ThemeMistSeaSalt,
		Mode:          appearance.ModeLight,
		ThemeElements: &provided,
	}, current)
	if err != nil {
		t.Fatalf("normalizeSetting returned error: %v", err)
	}

	defaults := appearance.DefaultThemeElements()
	seaSalt := got.ThemeElements[appearance.ThemeMistSeaSalt]
	if seaSalt.HomeLatestEmptyDescription != defaults[appearance.ThemeMistSeaSalt].HomeLatestEmptyDescription {
		t.Fatalf("empty provided field = %q, want default", seaSalt.HomeLatestEmptyDescription)
	}
	if seaSalt.HomeLatestEndText != "新的海盐结束文案" {
		t.Fatalf("provided end text = %q", seaSalt.HomeLatestEndText)
	}
	if got.ThemeElements[appearance.ThemeMistForest] != current[appearance.ThemeMistForest] {
		t.Fatalf("omitted forest theme = %#v, want preserved %#v", got.ThemeElements[appearance.ThemeMistForest], current[appearance.ThemeMistForest])
	}
}

func TestNormalizeSettingRejectsInvalidThemeElements(t *testing.T) {
	t.Parallel()

	tests := []appearance.ThemeElementMap{
		{"unknown-theme": {}},
		{appearance.ThemeMistForest: {
			HomeLatestEmptyDescription: strings.Repeat("林", appearance.HomeLatestEmptyDescriptionMaxRunes+1),
		}},
		{appearance.ThemeMistSeaSalt: {
			HomeLatestEndText: strings.Repeat("潮", appearance.HomeLatestEndTextMaxRunes+1),
		}},
	}
	for _, elements := range tests {
		_, err := normalizeSetting(SettingSaveRequest{
			SiteName:      "Blog",
			Author:        "Author",
			Theme:         appearance.ThemeMistSeaSalt,
			Mode:          appearance.ModeLight,
			ThemeElements: &elements,
		}, appearance.DefaultThemeElements())
		appErr, ok := err.(apperrors.AppError)
		if !ok || appErr.Code != apperrors.CodeInvalidParameter {
			t.Fatalf("normalizeSetting(%#v) error = %#v, want invalid parameter", elements, err)
		}
	}
}

func TestUpdatePreservesThemeElementsWhenRequestOmitsThem(t *testing.T) {
	t.Parallel()

	settings := NewSettingService(nil, nil)
	custom := appearance.DefaultThemeElements()
	custom[appearance.ThemeMistForest] = appearance.ThemeElements{
		HomeLatestEmptyDescription: "林间新文章正在酝酿。",
		HomeLatestEndText:          "自定义林径终点",
	}
	if _, err := settings.Update(SettingSaveRequest{
		SiteName:      "Blog",
		Author:        "Author",
		Theme:         appearance.ThemeMistForest,
		Mode:          appearance.ModeDark,
		ThemeElements: &custom,
	}); err != nil {
		t.Fatalf("initial Update returned error: %v", err)
	}

	got, err := settings.Update(SettingSaveRequest{
		SiteName: "Renamed Blog",
		Author:   "Author",
		Theme:    appearance.ThemeMistForest,
		Mode:     appearance.ModeLight,
	})
	if err != nil {
		t.Fatalf("legacy Update returned error: %v", err)
	}
	if !reflect.DeepEqual(got.ThemeElements, custom) {
		t.Fatalf("legacy Update theme elements = %#v, want preserved %#v", got.ThemeElements, custom)
	}
}

func TestSettingFromModelFallsBackFromMalformedThemeElementsJSON(t *testing.T) {
	t.Parallel()

	item := settingFromModel(model.SiteSetting{
		SiteName:          "Blog",
		Author:            "Author",
		Theme:             appearance.ThemeMistSeaSalt,
		Mode:              appearance.ModeLight,
		ThemeElementsJSON: "{not-json",
	})
	if !reflect.DeepEqual(item.ThemeElements, appearance.DefaultThemeElements()) {
		t.Fatalf("theme elements = %#v, want defaults", item.ThemeElements)
	}
}

func TestCloneSettingDeepCopiesThemeElements(t *testing.T) {
	t.Parallel()

	original := defaultLobbySetting()
	cloned := cloneSetting(original)
	cloned.ThemeElements[appearance.ThemeMistSeaSalt] = appearance.ThemeElements{
		HomeLatestEndText: "changed",
	}
	if original.ThemeElements[appearance.ThemeMistSeaSalt].HomeLatestEndText == "changed" {
		t.Fatal("cloneSetting shared its theme element map")
	}
}
