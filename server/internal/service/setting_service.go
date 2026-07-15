package service

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"solitude-blog/server/internal/appearance"
	"solitude-blog/server/internal/cache"
	apperrors "solitude-blog/server/internal/errors"
	"solitude-blog/server/internal/model"
)

const (
	defaultSiteSettingID uint64 = 1
	siteSettingCacheTTL         = 10 * time.Minute
)

type SettingService struct {
	db      *gorm.DB
	redis   *redis.Client
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

func NewSettingService(db *gorm.DB, redisClient *redis.Client) *SettingService {
	return &SettingService{
		db:      db,
		redis:   redisClient,
		setting: defaultLobbySetting(),
	}
}

func (s *SettingService) Lobby() (LobbySetting, error) {
	return s.Detail()
}

func (s *SettingService) Detail() (LobbySetting, error) {
	if s.db != nil {
		if item, ok := s.getCachedSetting(); ok {
			return item, nil
		}
		row, err := s.loadOrCreate()
		if err != nil {
			return LobbySetting{}, err
		}
		item := settingFromModel(row)
		s.setCachedSetting(item)
		return item, nil
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
		s.invalidateSettingCache()
		return settingFromModel(row), nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.setting = cloneSetting(normalized)
	return cloneSetting(s.setting), nil
}

func (s *SettingService) getCachedSetting() (LobbySetting, bool) {
	if s.redis == nil {
		return LobbySetting{}, false
	}
	payload, err := s.redis.Get(context.Background(), cache.SiteSettingsKey()).Bytes()
	if err != nil {
		return LobbySetting{}, false
	}
	var item LobbySetting
	if err := json.Unmarshal(payload, &item); err != nil {
		return LobbySetting{}, false
	}
	item.Theme = appearance.NormalizeTheme(item.Theme)
	item.Mode = appearance.NormalizeMode(item.Mode)
	return item, true
}

func (s *SettingService) setCachedSetting(item LobbySetting) {
	if s.redis == nil {
		return
	}
	payload, err := json.Marshal(item)
	if err != nil {
		return
	}
	_ = s.redis.Set(context.Background(), cache.SiteSettingsKey(), payload, siteSettingCacheTTL).Err()
}

func (s *SettingService) invalidateSettingCache() {
	if s.redis == nil {
		return
	}
	_ = s.redis.Del(context.Background(), cache.SiteSettingsKey()).Err()
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
	if !appearance.IsValidTheme(req.Theme) || !appearance.IsValidMode(req.Mode) {
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
		Theme:       appearance.NormalizeTheme(row.Theme),
		Mode:        appearance.NormalizeMode(row.Mode),
		SocialLinks: links,
	}
}

func defaultLobbySetting() LobbySetting {
	return LobbySetting{
		SiteName: "Solitude Blog",
		Author:   "Solitude King",
		Essay:    "Keep writing, keep shipping.",
		Theme:    appearance.DefaultTheme,
		Mode:     appearance.DefaultMode,
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
