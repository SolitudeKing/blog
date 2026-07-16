package model

import "time"

type SiteSetting struct {
	ID                uint64    `gorm:"primaryKey" json:"id"`
	SiteName          string    `gorm:"size:120;not null" json:"site_name"`
	Author            string    `gorm:"size:80;not null" json:"author"`
	Essay             string    `gorm:"size:500" json:"essay"`
	Theme             string    `gorm:"size:32;not null" json:"theme"`
	Mode              string    `gorm:"size:32;not null" json:"mode"`
	SocialLinksJSON   string    `gorm:"type:json" json:"-"`
	ThemeElementsJSON string    `gorm:"type:json" json:"-"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
