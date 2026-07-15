package service

import (
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
			got, err := normalizeSetting(req)
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
		_, err := normalizeSetting(req)
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
}
