package service

import (
	"context"
	"encoding/json"
	"hash/fnv"
	"math/rand"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"solitude-blog/server/internal/cache"
)

const (
	// suggestionTopicCap / suggestionTagCap 限定两类航标各自的上限，
	// suggestionTotal 是整组航标的总量；不足时从另一池补齐。
	suggestionTopicCap = 4
	suggestionTagCap   = 3
	suggestionTotal    = 6
	// suggestionCacheTTL 只兜底过期时间：缓存键本身带 UTC 日期，
	// 跨天后自然换键，因此无需主动失效。
	suggestionCacheTTL = 24 * time.Hour
)

// SuggestionService builds the daily "search beacons" (试试这些航标) for the
// search page: a deterministic, UTC-date-seeded sample drawn from topic and
// tag names, so every visitor sees the same set within one day.
type SuggestionService struct {
	topics *TopicService
	tags   *TagService
	redis  *redis.Client
}

// SuggestionItem is a single search chip. Kind is "topic" or "tag".
type SuggestionItem struct {
	Text string `json:"text"`
	Kind string `json:"kind"`
}

func NewSuggestionService(topics *TopicService, tags *TagService, redisClient *redis.Client) *SuggestionService {
	return &SuggestionService{topics: topics, tags: tags, redis: redisClient}
}

// Suggestions returns the daily suggestion chips. The result is cached per
// UTC date in Redis; on cache miss it reuses TopicService.List / TagService.List
// so the in-memory fallback (db == nil) keeps working in tests and dev.
func (s *SuggestionService) Suggestions() ([]SuggestionItem, error) {
	date := time.Now().UTC().Format("2006-01-02")
	key := cache.SearchSuggestionsKey(date)

	if s.redis != nil {
		payload, err := s.redis.Get(context.Background(), key).Bytes()
		if err == nil {
			var items []SuggestionItem
			if json.Unmarshal(payload, &items) == nil {
				return items, nil
			}
		}
	}

	topicItems, err := s.topics.List()
	if err != nil {
		return nil, err
	}
	tagItems, err := s.tags.List()
	if err != nil {
		return nil, err
	}
	topicNames := make([]string, 0, len(topicItems))
	for _, item := range topicItems {
		if name := strings.TrimSpace(item.Name); name != "" {
			topicNames = append(topicNames, name)
		}
	}
	tagNames := make([]string, 0, len(tagItems))
	for _, item := range tagItems {
		if name := strings.TrimSpace(item.Name); name != "" {
			tagNames = append(tagNames, name)
		}
	}

	items := buildSuggestions(topicNames, tagNames, seedForDate(date))
	if s.redis != nil {
		if payload, err := json.Marshal(items); err == nil {
			_ = s.redis.Set(context.Background(), key, payload, suggestionCacheTTL).Err()
		}
	}
	return items, nil
}

// seedForDate hashes a UTC date string into a PRNG seed so the same day
// always produces the same suggestion set.
func seedForDate(date string) int64 {
	hash := fnv.New64a()
	_, _ = hash.Write([]byte(date))
	return int64(hash.Sum64())
}

// buildSuggestions draws up to suggestionTopicCap topic names and up to
// suggestionTagCap tag names, deduplicates by text across both pools (the
// frontend renders chips keyed by text), fills any shortfall from the other
// pool up to suggestionTotal and finally shuffles the combined list with the
// date seed.
func buildSuggestions(topicNames []string, tagNames []string, seed int64) []SuggestionItem {
	rng := rand.New(rand.NewSource(seed))

	topicPool := dedupeNames(topicNames)
	tagPool := dedupeNames(tagNames)
	// 专题与标签可能同名；跨池按文本去重，专题优先保留。
	topicSet := make(map[string]bool, len(topicPool))
	for _, name := range topicPool {
		topicSet[name] = true
	}
	filteredTags := make([]string, 0, len(tagPool))
	for _, name := range tagPool {
		if !topicSet[name] {
			filteredTags = append(filteredTags, name)
		}
	}
	tagPool = filteredTags

	rng.Shuffle(len(topicPool), func(i, j int) { topicPool[i], topicPool[j] = topicPool[j], topicPool[i] })
	rng.Shuffle(len(tagPool), func(i, j int) { tagPool[i], tagPool[j] = tagPool[j], tagPool[i] })

	items := make([]SuggestionItem, 0, suggestionTotal)
	appendKind := func(pool *[]string, cap int, kind string) {
		for len(items) < suggestionTotal && len(*pool) > 0 && cap > 0 {
			items = append(items, SuggestionItem{Text: (*pool)[0], Kind: kind})
			*pool = (*pool)[1:]
			cap--
		}
	}
	appendKind(&topicPool, suggestionTopicCap, "topic")
	appendKind(&tagPool, suggestionTagCap, "tag")
	// 不足总量时从另一池补齐：先标签池，再专题池。两池已按文本去重，无重复键风险。
	appendKind(&tagPool, suggestionTotal-len(items), "tag")
	appendKind(&topicPool, suggestionTotal-len(items), "topic")

	rng.Shuffle(len(items), func(i, j int) { items[i], items[j] = items[j], items[i] })
	return items
}

func dedupeNames(names []string) []string {
	seen := make(map[string]bool, len(names))
	result := make([]string, 0, len(names))
	for _, name := range names {
		if !seen[name] {
			seen[name] = true
			result = append(result, name)
		}
	}
	return result
}
