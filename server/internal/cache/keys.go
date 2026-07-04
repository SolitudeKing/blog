package cache

import "fmt"

const prefix = "blog:v1"

func ArticleDetailKey(slug string) string {
	return fmt.Sprintf("%s:article:detail:%s", prefix, slug)
}

func ArticleListKey(cursor string, limit int, filterHash string) string {
	return fmt.Sprintf("%s:article:list:%s:%d:%s", prefix, cursor, limit, filterHash)
}

func SiteSettingsKey() string {
	return prefix + ":site:settings"
}
