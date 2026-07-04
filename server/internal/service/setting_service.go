package service

type SettingService struct{}

type LobbySetting struct {
	SiteName    string            `json:"site_name"`
	Author      string            `json:"author"`
	Essay       string            `json:"essay"`
	Theme       string            `json:"theme"`
	Mode        string            `json:"mode"`
	SocialLinks map[string]string `json:"social_links"`
}

func NewSettingService() *SettingService {
	return &SettingService{}
}

func (s *SettingService) Lobby() LobbySetting {
	return LobbySetting{
		SiteName: "Solitude Blog",
		Author:   "Solitude King",
		Essay:    "Keep writing, keep shipping.",
		Theme:    "forest",
		Mode:     "light",
		SocialLinks: map[string]string{
			"gitee":    "",
			"bilibili": "",
			"douyin":   "",
		},
	}
}

func (s *SettingService) Detail() LobbySetting {
	return s.Lobby()
}
