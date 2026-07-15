package appearance

import "testing"

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
