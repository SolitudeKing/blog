package model

import "time"

type Article struct {
	ID          uint64
	Title       string
	Slug        string
	Summary     string
	ContentMD   string
	Status      string
	CategoryID  uint64
	AuthorID    uint64
	ViewCount   uint64
	PublishedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}
