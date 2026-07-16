package model

import (
	"time"

	"gorm.io/gorm"
)

type Article struct {
	ID          uint64         `gorm:"primaryKey" json:"id"`
	Title       string         `gorm:"size:180;not null" json:"title"`
	Slug        string         `gorm:"size:220;not null;uniqueIndex" json:"slug"`
	Summary     string         `gorm:"size:500" json:"summary"`
	ContentMD   string         `gorm:"type:longtext" json:"content_md"`
	Status      string         `gorm:"size:32;not null;index" json:"status"`
	TopicID     uint64         `gorm:"index" json:"topic_id"`
	Topic       Topic          `json:"topic_detail"`
	AuthorID    uint64         `gorm:"index" json:"author_id"`
	Author      User           `json:"author_detail"`
	Tags        []Tag          `gorm:"many2many:article_tags;" json:"tag_details"`
	ViewCount   uint64         `gorm:"not null;default:0" json:"view_count"`
	PublishedAt *time.Time     `gorm:"index" json:"published_at"`
	CreatedAt   time.Time      `gorm:"index" json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
