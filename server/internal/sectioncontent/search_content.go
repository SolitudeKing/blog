package sectioncontent

const (
	// SearchKickerMaxRunes caps the small label rendered above the search
	// heading (e.g. "Search the current").
	SearchKickerMaxRunes = 32
	// SearchHeadingMaxRunes caps the heading rendered inside the search hero.
	SearchHeadingMaxRunes = 64
	// SearchIntroMaxRunes caps the multi-line description rendered below the
	// search heading.
	SearchIntroMaxRunes = 240
	// SearchPlaceholderMaxRunes caps the placeholder of the keyword input.
	SearchPlaceholderMaxRunes = 64
	// SearchSuggestionLabelMaxRunes caps the label rendered above the daily
	// suggestion chips (e.g. "试试这些航标").
	SearchSuggestionLabelMaxRunes = 32
	// SearchSuggestionFallbacksMaxRunes caps the newline-joined fallback
	// keywords used when the daily suggestion endpoint returns nothing.
	SearchSuggestionFallbacksMaxRunes = 160
	// SearchEmptyTitleMaxRunes caps the title shown when a search has no hits.
	SearchEmptyTitleMaxRunes = 64
	// SearchEmptyDescriptionMaxRunes caps the description shown together
	// with the search empty title.
	SearchEmptyDescriptionMaxRunes = 160
)

// SearchContent is the full set of theme-independent search page copy
// fields. Every field is required so the admin form and the public renderer
// can rely on a complete object.
type SearchContent struct {
	SearchKicker              string `json:"search_kicker"`
	SearchHeading             string `json:"search_heading"`
	SearchIntro               string `json:"search_intro"`
	SearchPlaceholder         string `json:"search_placeholder"`
	SearchSuggestionLabel     string `json:"search_suggestion_label"`
	SearchSuggestionFallbacks string `json:"search_suggestion_fallbacks"`
	SearchEmptyTitle          string `json:"search_empty_title"`
	SearchEmptyDescription    string `json:"search_empty_description"`
}

// DefaultSearchContent returns a complete, independently allocated copy of
// the shipped search page copy. The values match the strings that used to be
// hard-coded in SearchPage.vue so the migration is byte-for-byte identical
// when no admin override is set.
func DefaultSearchContent() SearchContent {
	return SearchContent{
		SearchKicker:              "Search the current",
		SearchHeading:             "打捞一段想法",
		SearchIntro:               "输入一个词，沿着标题、摘要、正文、专题与标签寻找。也可以从常用航标开始，看看它会把你带去哪里。",
		SearchPlaceholder:         "例如：设计系统、写作、Vue……",
		SearchSuggestionLabel:     "试试这些航标",
		SearchSuggestionFallbacks: "设计\n代码\n写作",
		SearchEmptyTitle:          "这片水域还没有记录",
		SearchEmptyDescription:    "可以尝试缩短关键词，或换一个角度重新出发。",
	}
}

// IsValidSearchContent reports whether every field satisfies its rune cap.
// Empty values are valid so NormalizeSearchContent can reset them to the
// shipped defaults.
func IsValidSearchContent(value SearchContent) bool {
	return !exceedsRuneLimit(value.SearchKicker, SearchKickerMaxRunes) &&
		!exceedsRuneLimit(value.SearchHeading, SearchHeadingMaxRunes) &&
		!exceedsRuneLimit(value.SearchIntro, SearchIntroMaxRunes) &&
		!exceedsRuneLimit(value.SearchPlaceholder, SearchPlaceholderMaxRunes) &&
		!exceedsRuneLimit(value.SearchSuggestionLabel, SearchSuggestionLabelMaxRunes) &&
		!exceedsRuneLimit(value.SearchSuggestionFallbacks, SearchSuggestionFallbacksMaxRunes) &&
		!exceedsRuneLimit(value.SearchEmptyTitle, SearchEmptyTitleMaxRunes) &&
		!exceedsRuneLimit(value.SearchEmptyDescription, SearchEmptyDescriptionMaxRunes)
}

// NormalizeSearchContent returns a complete SearchContent by trimming every
// field and falling back to the corresponding default when the input is
// empty or longer than the per-field rune cap.
func NormalizeSearchContent(value SearchContent) SearchContent {
	defaults := DefaultSearchContent()
	return SearchContent{
		SearchKicker:              normalizeSectionText(value.SearchKicker, defaults.SearchKicker, SearchKickerMaxRunes),
		SearchHeading:             normalizeSectionText(value.SearchHeading, defaults.SearchHeading, SearchHeadingMaxRunes),
		SearchIntro:               normalizeSectionText(value.SearchIntro, defaults.SearchIntro, SearchIntroMaxRunes),
		SearchPlaceholder:         normalizeSectionText(value.SearchPlaceholder, defaults.SearchPlaceholder, SearchPlaceholderMaxRunes),
		SearchSuggestionLabel:     normalizeSectionText(value.SearchSuggestionLabel, defaults.SearchSuggestionLabel, SearchSuggestionLabelMaxRunes),
		SearchSuggestionFallbacks: normalizeSectionText(value.SearchSuggestionFallbacks, defaults.SearchSuggestionFallbacks, SearchSuggestionFallbacksMaxRunes),
		SearchEmptyTitle:          normalizeSectionText(value.SearchEmptyTitle, defaults.SearchEmptyTitle, SearchEmptyTitleMaxRunes),
		SearchEmptyDescription:    normalizeSectionText(value.SearchEmptyDescription, defaults.SearchEmptyDescription, SearchEmptyDescriptionMaxRunes),
	}
}

// CloneSearchContent returns a value-level copy. All fields are strings so a
// shallow struct copy is sufficient.
func CloneSearchContent(value SearchContent) SearchContent {
	return value
}
