package bootstrap

import (
	"context"
	"testing"

	"solitude-blog/server/internal/config"
)

func TestNewAppValidatesConfigBeforeOpeningResources(t *testing.T) {
	t.Parallel()

	app, resources, err := NewApp(context.Background(), config.Config{})
	if err == nil {
		t.Fatal("NewApp() error = nil, want config validation error")
	}
	if app != nil || resources.DB != nil || resources.Redis != nil {
		t.Fatalf("NewApp() returned initialized resources after validation failure: %#v", resources)
	}
}
