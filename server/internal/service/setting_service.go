package service

import (
	"encoding/json"
	"errors"
	"sync"

	"gorm.io/gorm"

	apperrors "solitude-blog/server/internal/errors"
	"solitude-blog/server/internal/model"
)

const defaultSiteSettingID uint64 = 1

type SettingService struct {
	db      *gorm.DB
	mu      sync.RWMutex
	setting LobbySetting
}

type LobbySetting struct {
	SiteName    string            `json:"site_name"`
	Author      string            `json:"author"`
	Essay       string            `json:"essay"`
	Theme       string            `json:"theme"`
	Mode        string            `json:"mode"`
	SocialLinks map[string]string `json:"social_links"`
}

type SettingSaveRequest = LobbySetting

func NewSettingService(db *gorm.DB) *SettingService {
	return &SettingService{
		db:      db,
		setting: defaultLobbySetting(),
	}
}

func (s *SettingService) Lobby() (LobbySetting, error) {
	return s.Detail()
}

func (s *SettingService) Detail() (LobbySetting, error) {
	if s.db != nil {
		row, err := s.loadOrCreate()
		if err != nil {
			return LobbySetting{}, err
		}
		return settingFromModel(row), nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	return cloneSetting(s.setting), nil
}

func (s *SettingService) Update(req SettingSaveRequest) (LobbySetting, error) {
	normalized, err := normalizeSetting(req)
	if err != nil {
		return LobbySetting{}, err
	}

	if s.db != nil {
		row, err := s.loadOrCreate()
		if err != nil {
			return LobbySetting{}, err
		}
		row.SiteName = normalized.SiteName
		row.Author = normalized.Author
		row.Essay = normalized.Essay
		row.Theme = normalized.Theme
		row.Mode = normalized.Mode
		row.SocialLinksJSON = mustMarshalSocialLinks(normalized.SocialLinks)
		if err := s.db.Save(&row).Error; err != nil {
			return LobbySetting{}, apperrors.New(apperrors.CodeDatabaseUnavailable)
		}
		return settingFromModel(row), nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.setting = cloneSetting(normalized)
	return cloneSetting(s.setting), nil
}

func (s *SettingService) loadOrCreate() (model.SiteSetting, error) {
	var row model.SiteSetting
	err := s.db.First(&row, defaultSiteSettingID).Error
	if err == nil {
		return row, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return model.SiteSetting{}, apperrors.New(apperrors.CodeDatabaseUnavailable)
	}

	defaults := defaultLobbySetting()
	row = model.SiteSetting{
		ID:              defaultSiteSettingID,
		SiteName:        defaults.SiteName,
		Author:          defaults.Author,
		Essay:           defaults.Essay,
		Theme:           defaults.Theme,
		Mode:            defaults.Mode,
		SocialLinksJSON: mustMarshalSocialLinks(defaults.SocialLinks),
	}
	if err := s.db.Create(&row).Error; err != nil {
		return model.SiteSetting{}, apperrors.New(apperrors.CodeDatabaseUnavailable)
	}
	return row, nil
}

func normalizeSetting(req SettingSaveRequest) (LobbySetting, error) {
	if req.SiteName == "" || req.Author == "" {
		return LobbySetting{}, apperrors.New(apperrors.CodeMissingRequiredField)
	}
	if req.Theme == "" {
		req.Theme = "forest"
	}
	if req.Mode == "" {
		req.Mode = "light"
	}
	if req.Theme != "forest" && req.Theme != "strawberry" {
		return LobbySetting{}, apperrors.New(apperrors.CodeInvalidParameter)
	}
	if req.Mode != "light" && req.Mode != "dark" {
		return LobbySetting{}, apperrors.New(apperrors.CodeInvalidParameter)
	}
	if req.SocialLinks == nil {
		req.SocialLinks = map[string]string{}
	}
	return cloneSetting(req), nil
}

func settingFromModel(row model.SiteSetting) LobbySetting {
	links := map[string]string{}
	if row.SocialLinksJSON != "" {
		_ = json.Unmarshal([]byte(row.SocialLinksJSON), &links)
	}
	return LobbySetting{
		SiteName:    row.SiteName,
		Author:      row.Author,
		Essay:       row.Essay,
		Theme:       row.Theme,
		Mode:        row.Mode,
		SocialLinks: links,
	}
}

func defaultLobbySetting() LobbySetting {
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
			"github":   "",
		},
	}
}

func cloneSetting(setting LobbySetting) LobbySetting {
	links := map[string]string{}
	for key, value := range setting.SocialLinks {
		links[key] = value
	}
	setting.SocialLinks = links
	return setting
}

func mustMarshalSocialLinks(links map[string]string) string {
	payload, err := json.Marshal(links)
	if err != nil {
		return "{}"
	}
	return string(payload)
}
