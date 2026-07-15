package appearance

import "strings"

const (
	ThemeMistSeaSalt = "mist-sea-salt"
	ThemeMistForest  = "mist-forest"
	DefaultTheme     = ThemeMistSeaSalt

	ModeLight   = "light"
	ModeDark    = "dark"
	DefaultMode = ModeLight
)

// IsValidTheme reports whether value is a theme that can be selected by an
// administrator. Legacy names are intentionally excluded from write-time
// validation and are handled only by NormalizeTheme when persisted data is
// read or imported.
func IsValidTheme(value string) bool {
	return value == ThemeMistSeaSalt || value == ThemeMistForest
}

// NormalizeTheme converts persisted and imported legacy values to the current
// theme contract. Unknown values safely fall back to the default theme.
func NormalizeTheme(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case ThemeMistForest, "forest":
		return ThemeMistForest
	case ThemeMistSeaSalt:
		return ThemeMistSeaSalt
	default:
		return DefaultTheme
	}
}

func IsValidMode(value string) bool {
	return value == ModeLight || value == ModeDark
}

// NormalizeMode converts persisted and imported values to the supported mode
// contract. Unknown values safely fall back to the default mode.
func NormalizeMode(value string) string {
	if strings.EqualFold(strings.TrimSpace(value), ModeDark) {
		return ModeDark
	}
	return DefaultMode
}
