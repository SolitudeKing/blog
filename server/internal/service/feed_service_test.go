package service

import (
	"context"
	"strings"
	"testing"
)

func TestFeedServiceRSSWithoutDatabase(t *testing.T) {
	service := NewFeedService(nil)

	payload, err := service.RSS(context.Background(), "https://solitude.example")
	if err != nil {
		t.Fatalf("RSS() error = %v", err)
	}
	content := string(payload)
	if !strings.Contains(content, `<rss version="2.0"`) {
		t.Fatalf("RSS payload missing rss root: %s", content)
	}
	if !strings.Contains(content, `<link>https://solitude.example</link>`) {
		t.Fatalf("RSS payload missing site link: %s", content)
	}
	if !strings.Contains(content, `href="https://solitude.example/rss.xml"`) {
		t.Fatalf("RSS payload missing self link: %s", content)
	}
}

func TestFeedServiceSitemapWithoutDatabase(t *testing.T) {
	service := NewFeedService(nil)

	payload, err := service.Sitemap(context.Background(), "https://solitude.example")
	if err != nil {
		t.Fatalf("Sitemap() error = %v", err)
	}
	content := string(payload)
	if !strings.Contains(content, `<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`) {
		t.Fatalf("sitemap payload missing urlset root: %s", content)
	}
	if !strings.Contains(content, `<loc>https://solitude.example</loc>`) {
		t.Fatalf("sitemap payload missing home URL: %s", content)
	}
	if !strings.Contains(content, `<loc>https://solitude.example/archives</loc>`) {
		t.Fatalf("sitemap payload missing archives URL: %s", content)
	}
}
