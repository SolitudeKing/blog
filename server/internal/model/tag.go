package model

import (
	"time"

	"gorm.io/gorm"
)

type Tag struct {
	ID          uint64         `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"size:80;not null" json:"name"`
	Slug        string         `gorm:"size:120;not null;uniqueIndex" json:"slug"`
	Description string         `gorm:"size:255" json:"description"`
	Color       string         `gorm:"size:32" json:"color"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
