package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"solitude-blog/server/internal/config"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newTestRouter(middleware gin.HandlerFunc) *gin.Engine {
	engine := gin.New()
	engine.Use(middleware)
	engine.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	return engine
}

func TestCorsForDevNoopInProduction(t *testing.T) {
	t.Parallel()

	engine := newTestRouter(CorsForDev(config.Config{AppEnv: "production"}))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/healthz", nil)
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	for header, value := range rec.Header() {
		key := strings.ToLower(header)
		if strings.HasPrefix(key, "access-control-") && len(value) > 0 {
			t.Fatalf("production leaked CORS header %s=%v", header, value)
		}
	}
}

func TestCorsForDevWritesHeadersInDevelopment(t *testing.T) {
	t.Parallel()

	engine := newTestRouter(CorsForDev(config.Config{AppEnv: "development"}))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/healthz", nil)
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	wantHeaders := map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
		"Access-Control-Allow-Headers": "Content-Type, Authorization, X-API-Version, X-Request-ID",
		"Access-Control-Expose-Headers": "X-Request-ID",
	}
	for header, want := range wantHeaders {
		if got := rec.Header().Get(header); got != want {
			t.Errorf("%s = %q, want %q", header, got, want)
		}
	}
}

func TestCorsForDevHandlesOptions(t *testing.T) {
	t.Parallel()

	engine := newTestRouter(CorsForDev(config.Config{AppEnv: "development"}))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("OPTIONS", "/healthz", nil)
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("Access-Control-Allow-Origin = %q, want *", got)
	}
}
