package cache

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
)

const (
	prefix        = "blog:v1"
	articlePrefix = "blog:v3"
)

func ArticleDetailKey(slug string) string {
	return fmt.Sprintf("%s:article:detail:%s", articlePrefix, slug)
}

func ArticleDetailPattern() string {
	return articlePrefix + ":article:detail:*"
}

func ArticleListKey(page int, pageSize int, filterHash string) string {
	return fmt.Sprintf("%s:article:list:%d:%d:%s", articlePrefix, page, pageSize, filterHash)
}

func ArticleListPattern() string {
	return articlePrefix + ":article:list:*"
}

func ArticleListFilterHash(topic string, tag string, keyword string, status string) string {
	hash := sha1.Sum([]byte(topic + "|" + tag + "|" + keyword + "|" + status))
	return hex.EncodeToString(hash[:])
}

func SiteSettingsKey() string {
	return prefix + ":site:settings"
}
