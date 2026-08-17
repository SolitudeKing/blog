package sectioncontent

const (
	// ArchiveKickerMaxRunes caps the small label rendered above the archive
	// heading (e.g. "Archive"). The public page appends the year range after
	// it, e.g. "Archive · 2024—2026".
	ArchiveKickerMaxRunes = 32
	// ArchiveHeadingMaxRunes caps the heading rendered inside the archive hero.
	ArchiveHeadingMaxRunes = 64
	// ArchiveIntroMaxRunes caps the multi-line description rendered below the
	// archive heading.
	ArchiveIntroMaxRunes = 240
	// ArchiveEmptyTitleMaxRunes caps the title shown when there are no
	// published articles.
	ArchiveEmptyTitleMaxRunes = 64
	// ArchiveEmptyDescriptionMaxRunes caps the description shown together
	// with the archive empty title.
	ArchiveEmptyDescriptionMaxRunes = 160
)

// ArchiveContent is the full set of theme-independent archive page copy
// fields. Every field is required so the admin form and the public renderer
// can rely on a complete object.
type ArchiveContent struct {
	ArchiveKicker           string `json:"archive_kicker"`
	ArchiveHeading          string `json:"archive_heading"`
	ArchiveIntro            string `json:"archive_intro"`
	ArchiveEmptyTitle       string `json:"archive_empty_title"`
	ArchiveEmptyDescription string `json:"archive_empty_description"`
}

// DefaultArchiveContent returns a complete, independently allocated copy of
// the shipped archive page copy. The values match the strings that used to be
// hard-coded in ArchivesPage.vue so the migration is byte-for-byte identical
// when no admin override is set.
func DefaultArchiveContent() ArchiveContent {
	return ArchiveContent{
		ArchiveKicker:           "Archive",
		ArchiveHeading:          "所有足迹，都有刻度",
		ArchiveIntro:            "从最近一次发布向过去回望。这里按年份与月份整理已经公开的文章，让每一段记录都能被重新抵达。",
		ArchiveEmptyTitle:       "还没有归档内容",
		ArchiveEmptyDescription: "发布文章后会按年/月自动汇总到这里。",
	}
}

// IsValidArchiveContent reports whether every field satisfies its rune cap.
// Empty values are valid so NormalizeArchiveContent can reset them to the
// shipped defaults.
func IsValidArchiveContent(value ArchiveContent) bool {
	return !exceedsRuneLimit(value.ArchiveKicker, ArchiveKickerMaxRunes) &&
		!exceedsRuneLimit(value.ArchiveHeading, ArchiveHeadingMaxRunes) &&
		!exceedsRuneLimit(value.ArchiveIntro, ArchiveIntroMaxRunes) &&
		!exceedsRuneLimit(value.ArchiveEmptyTitle, ArchiveEmptyTitleMaxRunes) &&
		!exceedsRuneLimit(value.ArchiveEmptyDescription, ArchiveEmptyDescriptionMaxRunes)
}

// NormalizeArchiveContent returns a complete ArchiveContent by trimming every
// field and falling back to the corresponding default when the input is
// empty or longer than the per-field rune cap.
func NormalizeArchiveContent(value ArchiveContent) ArchiveContent {
	defaults := DefaultArchiveContent()
	return ArchiveContent{
		ArchiveKicker:           normalizeSectionText(value.ArchiveKicker, defaults.ArchiveKicker, ArchiveKickerMaxRunes),
		ArchiveHeading:          normalizeSectionText(value.ArchiveHeading, defaults.ArchiveHeading, ArchiveHeadingMaxRunes),
		ArchiveIntro:            normalizeSectionText(value.ArchiveIntro, defaults.ArchiveIntro, ArchiveIntroMaxRunes),
		ArchiveEmptyTitle:       normalizeSectionText(value.ArchiveEmptyTitle, defaults.ArchiveEmptyTitle, ArchiveEmptyTitleMaxRunes),
		ArchiveEmptyDescription: normalizeSectionText(value.ArchiveEmptyDescription, defaults.ArchiveEmptyDescription, ArchiveEmptyDescriptionMaxRunes),
	}
}

// CloneArchiveContent returns a value-level copy. All fields are strings so a
// shallow struct copy is sufficient.
func CloneArchiveContent(value ArchiveContent) ArchiveContent {
	return value
}
