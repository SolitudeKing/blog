package model

import (
	"time"

	"gorm.io/gorm"
)

type Asset struct {
	ID          uint64         `gorm:"primaryKey" json:"id"`
	DisplayName string         `gorm:"size:255;not null" json:"display_name"`
	AltText     string         `gorm:"size:255" json:"alt_text"`
	StorageKey  string         `gorm:"size:500;not null;uniqueIndex" json:"storage_key"`
	URL         string         `gorm:"size:500;not null" json:"url"`
	ThumbURL    string         `gorm:"size:500" json:"thumb_url"`
	MimeType    string         `gorm:"size:100;not null;index" json:"mime_type"`
	Ext         string         `gorm:"size:20" json:"ext"`
	Size        uint64         `gorm:"not null;default:0" json:"size"`
	Width       uint           `json:"width"`
	Height      uint           `json:"height"`
	SHA256      string         `gorm:"size:64;not null;index" json:"sha256"`
	Status      string         `gorm:"size:32;not null;index" json:"status"`
	RefCount    uint           `gorm:"not null;default:0" json:"ref_count"`
	CreatedAt   time.Time      `gorm:"index" json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
