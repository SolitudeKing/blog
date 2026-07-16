package cache

import (
	"strings"
	"testing"
)

func TestArticleKeysUseTopicAwareSchemaVersion(t *testing.T) {
	keys := []string{
		ArticleDetailKey("welcome"),
		ArticleListKey("", 20, ArticleListFilterHash("notes", "go", "", "published")),
		ArticleListPattern(),
	}
	for _, key := range keys {
		if !strings.HasPrefix(key, "blog:v2:article:") {
			t.Fatalf("article cache key %q does not use v2 schema", key)
		}
	}
	if key := SiteSettingsKey(); !strings.HasPrefix(key, "blog:v1:") {
		t.Fatalf("site settings key %q unexpectedly changed schema", key)
	}
}
