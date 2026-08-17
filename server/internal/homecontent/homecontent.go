// Package homecontent owns the per-site home page copy contract. Values are
// theme-independent so they live as a single JSON column on site_settings
// instead of inside appearance.ThemeElements.
package homecontent

import (
	"strings"
	"unicode/utf8"
)

const (
	// HomeProfileKickerMaxRunes caps the small label rendered above the
	// author identity block on the home page (e.g. "Blog keeper · Solitude").
	HomeProfileKickerMaxRunes = 32
	// HomeHeadingPrefixMaxRunes caps the short greeting prefix rendered
	// before the author name (e.g. "你好，我是").
	HomeHeadingPrefixMaxRunes = 16
	// HomeStatusFallbackMaxRunes caps the status line shown when there is
	// no active site notice.
	HomeStatusFallbackMaxRunes = 80
	// HomeIntroHeadingMaxRunes caps the heading rendered inside the
	// home-intro aside.
	HomeIntroHeadingMaxRunes = 80
	// HomeIntroParagraphMaxRunes caps the multi-line description rendered
	// below the home-intro heading.
	HomeIntroParagraphMaxRunes = 240
	// HomeActionLabelMaxRunes caps the two short action button labels on
	// the home-intro aside.
	HomeActionLabelMaxRunes = 16
	// HomeLatestKickerMaxRunes caps the small label rendered above the
	// "Latest posts" heading.
	HomeLatestKickerMaxRunes = 32
	// HomeLatestHeadingMaxRunes caps the heading rendered above the latest
	// articles list.
	HomeLatestHeadingMaxRunes = 64
	// HomeLatestViewAllLabelMaxRunes caps the "view all" link label.
	HomeLatestViewAllLabelMaxRunes = 16
	// HomeLatestEmptyTitleMaxRunes caps the title shown when there are no
	// published articles.
	HomeLatestEmptyTitleMaxRunes = 64
	// HomeTopicsKickerMaxRunes caps the small label rendered above the
	// topics directory heading.
	HomeTopicsKickerMaxRunes = 32
	// HomeTopicsHeadingMaxRunes caps the heading rendered above the topics
	// directory links.
	HomeTopicsHeadingMaxRunes = 64
	// HomeNoticeKickerMaxRunes caps the small label rendered above the site
	// notice heading.
	HomeNoticeKickerMaxRunes = 32
	// HomeNoticeActionLabelMaxRunes caps the short action link label on the
	// site notice card.
	HomeNoticeActionLabelMaxRunes = 16
)

// HomeContent is the full set of theme-independent home page copy fields.
// Every field is required so the admin form and the public renderer can
// rely on a complete object.
type HomeContent struct {
	HomeProfileKicker          string `json:"home_profile_kicker"`
	HomeHeadingPrefix          string `json:"home_heading_prefix"`
	HomeStatusFallback         string `json:"home_status_fallback"`
	HomeIntroHeading           string `json:"home_intro_heading"`
	HomeIntroParagraph         string `json:"home_intro_paragraph"`
	HomeActionViewRecentLabel  string `json:"home_action_view_recent_label"`
	HomeActionViewArchiveLabel string `json:"home_action_view_archive_label"`
	HomeLatestKicker           string `json:"home_latest_kicker"`
	HomeLatestHeading          string `json:"home_latest_heading"`
	HomeLatestViewAllLabel     string `json:"home_latest_view_all_label"`
	HomeLatestEmptyTitle       string `json:"home_latest_empty_title"`
	HomeTopicsKicker           string `json:"home_topics_kicker"`
	HomeTopicsHeading          string `json:"home_topics_heading"`
	HomeNoticeKicker           string `json:"home_notice_kicker"`
	HomeNoticeActionLabel      string `json:"home_notice_action_label"`
}

// DefaultHomeContent returns a complete, independently allocated copy of the
// shipped home page copy. The values match the strings that used to be
// hard-coded in HomePage.vue so the migration is byte-for-byte identical
// when no admin override is set.
func DefaultHomeContent() HomeContent {
	return HomeContent{
		HomeProfileKicker:          "Blog keeper · Solitude",
		HomeHeadingPrefix:          "你好，我是",
		HomeStatusFallback:         "持续记录技术、设计与生活",
		HomeIntroHeading:           "一份持续更新的博客，也是公开的思考现场",
		HomeIntroParagraph:         "在这里记录工程实践、设计系统与构建过程。文章保留可复用的方法，也保留问题发生时的真实判断。",
		HomeActionViewRecentLabel:  "查看最近发布",
		HomeActionViewArchiveLabel: "浏览全部归档",
		HomeLatestKicker:           "Latest posts",
		HomeLatestHeading:          "最近发布的博客",
		HomeLatestViewAllLabel:     "查看全部归档",
		HomeLatestEmptyTitle:       "暂时还没有发布文章",
		HomeTopicsKicker:           "Topics",
		HomeTopicsHeading:          "从这些专题进入",
		HomeNoticeKicker:           "Site notice",
		HomeNoticeActionLabel:      "继续阅读",
	}
}

// IsValidHomeContent reports whether every field satisfies its rune cap.
// Empty values are valid so NormalizeHomeContent can reset them to the
// shipped defaults.
func IsValidHomeContent(value HomeContent) bool {
	return !exceedsRuneLimit(value.HomeProfileKicker, HomeProfileKickerMaxRunes) &&
		!exceedsRuneLimit(value.HomeHeadingPrefix, HomeHeadingPrefixMaxRunes) &&
		!exceedsRuneLimit(value.HomeStatusFallback, HomeStatusFallbackMaxRunes) &&
		!exceedsRuneLimit(value.HomeIntroHeading, HomeIntroHeadingMaxRunes) &&
		!exceedsRuneLimit(value.HomeIntroParagraph, HomeIntroParagraphMaxRunes) &&
		!exceedsRuneLimit(value.HomeActionViewRecentLabel, HomeActionLabelMaxRunes) &&
		!exceedsRuneLimit(value.HomeActionViewArchiveLabel, HomeActionLabelMaxRunes) &&
		!exceedsRuneLimit(value.HomeLatestKicker, HomeLatestKickerMaxRunes) &&
		!exceedsRuneLimit(value.HomeLatestHeading, HomeLatestHeadingMaxRunes) &&
		!exceedsRuneLimit(value.HomeLatestViewAllLabel, HomeLatestViewAllLabelMaxRunes) &&
		!exceedsRuneLimit(value.HomeLatestEmptyTitle, HomeLatestEmptyTitleMaxRunes) &&
		!exceedsRuneLimit(value.HomeTopicsKicker, HomeTopicsKickerMaxRunes) &&
		!exceedsRuneLimit(value.HomeTopicsHeading, HomeTopicsHeadingMaxRunes) &&
		!exceedsRuneLimit(value.HomeNoticeKicker, HomeNoticeKickerMaxRunes) &&
		!exceedsRuneLimit(value.HomeNoticeActionLabel, HomeNoticeActionLabelMaxRunes)
}

// NormalizeHomeContent returns a complete HomeContent by trimming every
// field and falling back to the corresponding default when the input is
// empty or longer than the per-field rune cap.
func NormalizeHomeContent(value HomeContent) HomeContent {
	defaults := DefaultHomeContent()
	return HomeContent{
		HomeProfileKicker:          normalizeHomeText(value.HomeProfileKicker, defaults.HomeProfileKicker, HomeProfileKickerMaxRunes),
		HomeHeadingPrefix:          normalizeHomeText(value.HomeHeadingPrefix, defaults.HomeHeadingPrefix, HomeHeadingPrefixMaxRunes),
		HomeStatusFallback:         normalizeHomeText(value.HomeStatusFallback, defaults.HomeStatusFallback, HomeStatusFallbackMaxRunes),
		HomeIntroHeading:           normalizeHomeText(value.HomeIntroHeading, defaults.HomeIntroHeading, HomeIntroHeadingMaxRunes),
		HomeIntroParagraph:         normalizeHomeText(value.HomeIntroParagraph, defaults.HomeIntroParagraph, HomeIntroParagraphMaxRunes),
		HomeActionViewRecentLabel:  normalizeHomeText(value.HomeActionViewRecentLabel, defaults.HomeActionViewRecentLabel, HomeActionLabelMaxRunes),
		HomeActionViewArchiveLabel: normalizeHomeText(value.HomeActionViewArchiveLabel, defaults.HomeActionViewArchiveLabel, HomeActionLabelMaxRunes),
		HomeLatestKicker:           normalizeHomeText(value.HomeLatestKicker, defaults.HomeLatestKicker, HomeLatestKickerMaxRunes),
		HomeLatestHeading:          normalizeHomeText(value.HomeLatestHeading, defaults.HomeLatestHeading, HomeLatestHeadingMaxRunes),
		HomeLatestViewAllLabel:     normalizeHomeText(value.HomeLatestViewAllLabel, defaults.HomeLatestViewAllLabel, HomeLatestViewAllLabelMaxRunes),
		HomeLatestEmptyTitle:       normalizeHomeText(value.HomeLatestEmptyTitle, defaults.HomeLatestEmptyTitle, HomeLatestEmptyTitleMaxRunes),
		HomeTopicsKicker:           normalizeHomeText(value.HomeTopicsKicker, defaults.HomeTopicsKicker, HomeTopicsKickerMaxRunes),
		HomeTopicsHeading:          normalizeHomeText(value.HomeTopicsHeading, defaults.HomeTopicsHeading, HomeTopicsHeadingMaxRunes),
		HomeNoticeKicker:           normalizeHomeText(value.HomeNoticeKicker, defaults.HomeNoticeKicker, HomeNoticeKickerMaxRunes),
		HomeNoticeActionLabel:      normalizeHomeText(value.HomeNoticeActionLabel, defaults.HomeNoticeActionLabel, HomeNoticeActionLabelMaxRunes),
	}
}

// CloneHomeContent returns a value-level copy. All fields are strings so a
// shallow struct copy is sufficient.
func CloneHomeContent(value HomeContent) HomeContent {
	return value
}

func exceedsRuneLimit(value string, limit int) bool {
	return utf8.RuneCountInString(strings.TrimSpace(value)) > limit
}

func normalizeHomeText(value string, fallback string, limit int) string {
	value = strings.TrimSpace(value)
	if value == "" || utf8.RuneCountInString(value) > limit {
		return fallback
	}
	return value
}
