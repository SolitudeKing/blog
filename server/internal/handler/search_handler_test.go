package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"solitude-blog/server/internal/service"
)

func TestSearchSuggestionsReturnsDailyChips(t *testing.T) {
	t.Parallel()

	handler := NewSearchHandler(
		service.NewSearchService(nil),
		service.NewSuggestionService(service.NewTopicService(nil, nil), service.NewTagService(nil, nil), nil),
	)

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/search/suggestions", nil)
	handler.Suggestions(context)

	if recorder.Code != http.StatusOK {
		t.Fatalf("Suggestions status = %d, want %d; body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var body struct {
		Code int                      `json:"code"`
		Data []service.SuggestionItem `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v; body=%s", err, recorder.Body.String())
	}
	if body.Code != 0 {
		t.Fatalf("response code = %d, want 0", body.Code)
	}
	if len(body.Data) == 0 {
		t.Fatal("suggestions data is empty, want seeded topic chips")
	}
	for _, item := range body.Data {
		if item.Kind != "topic" && item.Kind != "tag" {
			t.Fatalf("suggestion kind = %q, want topic or tag", item.Kind)
		}
		if item.Text == "" {
			t.Fatal("suggestion text is empty")
		}
	}
}

func TestSettingLobbyIncludesSectionContentGroups(t *testing.T) {
	t.Parallel()

	handler := NewSettingHandler(service.NewSettingService(nil, nil))

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/setting/lobby", nil)
	handler.Lobby(context)

	if recorder.Code != http.StatusOK {
		t.Fatalf("Lobby status = %d, want %d; body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var body struct {
		Code int `json:"code"`
		Data struct {
			ArchiveContent struct {
				ArchiveHeading string `json:"archive_heading"`
			} `json:"archive_content"`
			SearchContent struct {
				SearchHeading string `json:"search_heading"`
			} `json:"search_content"`
			AboutContent struct {
				AboutHeading string `json:"about_heading"`
			} `json:"about_content"`
			HomeContent struct {
				HomeTopicsHeading string `json:"home_topics_heading"`
			} `json:"home_content"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v; body=%s", err, recorder.Body.String())
	}
	if body.Data.ArchiveContent.ArchiveHeading == "" {
		t.Fatal("archive_content missing or empty in lobby payload")
	}
	if body.Data.SearchContent.SearchHeading == "" {
		t.Fatal("search_content missing or empty in lobby payload")
	}
	if body.Data.AboutContent.AboutHeading == "" {
		t.Fatal("about_content missing or empty in lobby payload")
	}
	if body.Data.HomeContent.HomeTopicsHeading == "" {
		t.Fatal("home_content.home_topics_heading missing or empty in lobby payload")
	}
}
