package model

import "time"

type ArticleVersion struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	ArticleID uint64    `gorm:"not null;index" json:"article_id"`
	Title     string    `gorm:"size:180;not null" json:"title"`
	Summary   string    `gorm:"size:500" json:"summary"`
	ContentMD string    `gorm:"type:longtext" json:"content_md"`
	Status    string    `gorm:"size:32;not null" json:"status"`
	CreatedBy uint64    `gorm:"index" json:"created_by"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}
