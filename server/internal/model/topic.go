package model

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	TopicLabelMaxRunes = 32

	TopicLabelNodes   = "NODES"
	TopicLabelCode    = "CODE"
	TopicLabelJotting = "JOTTING"

	TopicSlugNodes   = "nodes"
	TopicSlugCode    = "code"
	TopicSlugJotting = "jotting"
)

// Topic 表示文章所属的编辑专题，一篇文章只能归属一个专题。
type Topic struct {
	ID          uint64         `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"size:80;not null" json:"name"`
	Label       string         `gorm:"size:32;not null" json:"label"`
	Slug        string         `gorm:"size:120;not null;uniqueIndex" json:"slug"`
	Description string         `gorm:"size:500" json:"description"`
	CoverURL    string         `gorm:"size:500" json:"cover_url"`
	SortOrder   int            `gorm:"not null;default:0" json:"sort_order"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// DefaultTopics 返回新博客固定的三个初始专题。调用方可以安全修改返回值。
func DefaultTopics() []Topic {
	return []Topic{
		{
			Name:        "雾里拾笺",
			Label:       TopicLabelNodes,
			Slug:        TopicSlugNodes,
			Description: "收拢阅读、学习与技术实践中散落的知识微光。",
			SortOrder:   1,
		},
		{
			Name:        "微光造物",
			Label:       TopicLabelCode,
			Slug:        TopicSlugCode,
			Description: "记录灵感如何经由设计、代码与实验长成作品。",
			SortOrder:   2,
		},
		{
			Name:        "风过留痕",
			Label:       TopicLabelJotting,
			Slug:        TopicSlugJotting,
			Description: "安放日常见闻、片刻心绪与未成体系的思考。",
			SortOrder:   3,
		},
	}
}

// DefaultTopicBySlug 返回正式专题契约的副本，避免业务层重复维护 label/slug 映射。
func DefaultTopicBySlug(slug string) (Topic, bool) {
	for _, topic := range DefaultTopics() {
		if topic.Slug == slug {
			return topic, true
		}
	}
	return Topic{}, false
}

// DefaultTopicLabel 为缺少 Label 的旧专题生成迁移兼容值。
func DefaultTopicLabel(name string) string {
	runes := []rune(strings.TrimSpace(name))
	if len(runes) == 0 {
		return "Topic"
	}
	if len(runes) > TopicLabelMaxRunes {
		runes = runes[:TopicLabelMaxRunes]
	}
	return string(runes)
}
