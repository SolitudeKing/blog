// Package sectioncontent owns the per-site copy contracts for the public
// sections that carry their own narrative text: archives, search and about.
// Values are theme-independent, so each section lives as its own JSON column
// on site_settings, mirroring the homecontent package.
package sectioncontent

import (
	"strings"
	"unicode/utf8"
)

func exceedsRuneLimit(value string, limit int) bool {
	return utf8.RuneCountInString(strings.TrimSpace(value)) > limit
}

func normalizeSectionText(value string, fallback string, limit int) string {
	value = strings.TrimSpace(value)
	if value == "" || utf8.RuneCountInString(value) > limit {
		return fallback
	}
	return value
}
