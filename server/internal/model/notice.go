package model

import (
	"time"

	"gorm.io/gorm"
)

type Notice struct {
	ID        uint64         `gorm:"primaryKey" json:"id"`
	Title     string         `gorm:"size:120;not null" json:"title"`
	Content   string         `gorm:"type:text" json:"content"`
	Enabled   bool           `gorm:"not null;default:false;index" json:"enabled"`
	SortOrder int            `gorm:"not null;default:0;index" json:"sort_order"`
	StartsAt  *time.Time     `gorm:"index" json:"starts_at"`
	EndsAt    *time.Time     `gorm:"index" json:"ends_at"`
	CreatedAt time.Time      `gorm:"index" json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
