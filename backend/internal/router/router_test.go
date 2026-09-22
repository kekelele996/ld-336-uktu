package router

import (
	"log/slog"
	"testing"

	"github.com/medasset/medasset/internal/config"
	"gorm.io/gorm"
)

// TestRouteRegistration 校验全部路由可注册且不冲突。
func TestRouteRegistration(t *testing.T) {
	cfg := &config.Config{
		ServerPort: "8080",
		RunMode:    "test",
		JWTSecret:  "test-secret",
		RateLimit:  100,
	}
	engine := New(Deps{DB: (*gorm.DB)(nil), Cfg: cfg, Log: slog.Default(), RDB: nil})
	routes := engine.Routes()
	if len(routes) < 20 {
		t.Fatalf("expected >= 20 routes, got %d", len(routes))
	}
	found := map[string]bool{}
	for _, r := range routes {
		found[r.Method+" "+r.Path] = true
	}
	for _, want := range []string{
		"GET /healthz",
		"POST /api/v1/auth/login",
		"GET /api/v1/devices",
		"POST /api/v1/purchases",
		"POST /api/v1/maintenances/plan/generate",
		"GET /api/v1/calibrations/due",
		"GET /api/v1/audits",
		"GET /api/v1/stats/overview",
	} {
		if !found[want] {
			t.Errorf("route not registered: %s", want)
		}
	}
}
