package model

import "time"

type Tag struct {
	ID          uint64
	Name        string
	Slug        string
	Description string
	Color       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
