package model

import "time"

type SiteSetting struct {
	ID                uint64    `gorm:"primaryKey" json:"id"`
	SiteName          string    `gorm:"size:120;not null" json:"site_name"`
	Author            string    `gorm:"size:80;not null" json:"author"`
	AuthorAvatarURL   string    `gorm:"size:500" json:"author_avatar_url"`
	Essay             string    `gorm:"size:500" json:"essay"`
	ICPNumber         string    `gorm:"size:64" json:"icp_number"`
	Theme             string    `gorm:"size:32;not null" json:"theme"`
	Mode              string    `gorm:"size:32;not null" json:"mode"`
	SocialLinksJSON   string    `gorm:"type:json" json:"-"`
	ThemeElementsJSON string    `gorm:"type:json" json:"-"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
