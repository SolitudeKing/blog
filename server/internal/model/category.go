package model

import "time"

type Category struct {
	ID          uint64
	Name        string
	Slug        string
	Description string
	SortOrder   int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
