package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"solitude-blog/server/internal/database"
)

func TestHealthzReturnsOKWhenOptionalResourcesAreDisabled(t *testing.T) {
	recorder := requestHealth(t, database.Resources{})
	if recorder.Code != http.StatusOK {
		t.Fatalf("Healthz status = %d, want %d", recorder.Code, http.StatusOK)
	}
}

func TestHealthzReturnsServiceUnavailableWhenRedisFails(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:        "127.0.0.1:0",
		DialTimeout: 20 * time.Millisecond,
		ReadTimeout: 20 * time.Millisecond,
	})
	t.Cleanup(func() { _ = client.Close() })

	recorder := requestHealth(t, database.Resources{Redis: client})
	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("Healthz status = %d, want %d; body=%s", recorder.Code, http.StatusServiceUnavailable, recorder.Body.String())
	}
	if !strings.Contains(recorder.Body.String(), `"redis":"error"`) {
		t.Fatalf("Healthz body = %s, want redis error", recorder.Body.String())
	}
}

func requestHealth(t *testing.T, resources database.Resources) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, "/healthz", nil)
	NewHealthHandler(resources).Healthz(context)
	return recorder
}
