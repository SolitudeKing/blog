package model

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

const TopicLabelMaxRunes = 32

// Topic groups articles into an editorial collection.
type Topic struct {
	ID          uint64         `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"size:80;not null" json:"name"`
	Label       string         `gorm:"size:32;not null" json:"label"`
	Slug        string         `gorm:"size:120;not null;uniqueIndex" json:"slug"`
	Description string         `gorm:"size:500" json:"description"`
	CoverURL    string         `gorm:"size:500" json:"cover_url"`
	SortOrder   int            `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// DefaultTopicLabel derives a migration-safe label when legacy data only has
// a name. Explicit labels are validated by the topic service instead.
func DefaultTopicLabel(name string) string {
	runes := []rune(strings.TrimSpace(name))
	if len(runes) == 0 {
		return "Topic"
	}
	if len(runes) > TopicLabelMaxRunes {
		runes = runes[:TopicLabelMaxRunes]
	}
	return string(runes)
}
