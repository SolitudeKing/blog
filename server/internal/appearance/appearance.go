package appearance

import (
	"strings"
	"unicode/utf8"
)

const (
	ThemeMistSeaSalt = "mist-sea-salt"
	ThemeMistForest  = "mist-forest"
	DefaultTheme     = ThemeMistSeaSalt

	ModeLight   = "light"
	ModeDark    = "dark"
	DefaultMode = ModeLight

	HomeLatestEmptyDescriptionMaxRunes = 160
	HomeLatestEndTextMaxRunes          = 80
)

type ThemeElements struct {
	HomeLatestEmptyDescription string `json:"home_latest_empty_description"`
	HomeLatestEndText          string `json:"home_latest_end_text"`
}

type ThemeElementMap map[string]ThemeElements

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

// DefaultThemeElements returns a complete, independently allocated element
// map for every supported site theme.
func DefaultThemeElements() ThemeElementMap {
	return ThemeElementMap{
		ThemeMistSeaSalt: {
			HomeLatestEmptyDescription: "第一篇文章正在潮汐之外酝酿。",
			HomeLatestEndText:          "已经读到潮汐尽头",
		},
		ThemeMistForest: {
			HomeLatestEmptyDescription: "第一篇文章正在林雾之间酝酿。",
			HomeLatestEndText:          "已经走到林径尽头",
		},
	}
}

// IsValidThemeElements validates administrator-provided values. Empty and
// missing values are valid because NormalizeThemeElements fills them with the
// theme defaults. Unknown themes and overlong values are rejected.
func IsValidThemeElements(elements ThemeElementMap) bool {
	for theme, element := range elements {
		if !IsValidTheme(theme) {
			return false
		}
		if exceedsRuneLimit(element.HomeLatestEmptyDescription, HomeLatestEmptyDescriptionMaxRunes) {
			return false
		}
		if exceedsRuneLimit(element.HomeLatestEndText, HomeLatestEndTextMaxRunes) {
			return false
		}
	}
	return true
}

// NormalizeThemeElements repairs persisted, imported, and partial element
// maps. Unknown themes are discarded and invalid values fall back to the
// corresponding supported theme default.
func NormalizeThemeElements(elements ThemeElementMap) ThemeElementMap {
	defaults := DefaultThemeElements()
	normalized := make(ThemeElementMap, len(defaults))
	for theme, fallback := range defaults {
		value := elements[theme]
		normalized[theme] = ThemeElements{
			HomeLatestEmptyDescription: normalizeElementText(
				value.HomeLatestEmptyDescription,
				fallback.HomeLatestEmptyDescription,
				HomeLatestEmptyDescriptionMaxRunes,
			),
			HomeLatestEndText: normalizeElementText(
				value.HomeLatestEndText,
				fallback.HomeLatestEndText,
				HomeLatestEndTextMaxRunes,
			),
		}
	}
	return normalized
}

func CloneThemeElements(elements ThemeElementMap) ThemeElementMap {
	cloned := make(ThemeElementMap, len(elements))
	for theme, element := range elements {
		cloned[theme] = element
	}
	return cloned
}

func exceedsRuneLimit(value string, limit int) bool {
	return utf8.RuneCountInString(strings.TrimSpace(value)) > limit
}

func normalizeElementText(value string, fallback string, limit int) string {
	value = strings.TrimSpace(value)
	if value == "" || utf8.RuneCountInString(value) > limit {
		return fallback
	}
	return value
}
