package cache

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
)

const prefix = "blog:v1"

func ArticleDetailKey(slug string) string {
	return fmt.Sprintf("%s:article:detail:%s", prefix, slug)
}

func ArticleListKey(cursor string, limit int, filterHash string) string {
	return fmt.Sprintf("%s:article:list:%s:%d:%s", prefix, cursor, limit, filterHash)
}

func ArticleListPattern() string {
	return prefix + ":article:list:*"
}

func ArticleListFilterHash(category string, tag string, keyword string, status string) string {
	hash := sha1.Sum([]byte(category + "|" + tag + "|" + keyword + "|" + status))
	return hex.EncodeToString(hash[:])
}

func SiteSettingsKey() string {
	return prefix + ":site:settings"
}
