package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("APP_HTTP_ADDR", "")
	t.Setenv("APP_LOG_LEVEL", "")
	t.Setenv("APP_VERSION", "")
	t.Setenv("APP_CORS_ORIGINS", "")
	t.Setenv("APP_SEED_LOCAL", "")
	t.Setenv("APP_ALLOW_EVENT_INGEST", "")
	t.Setenv("APP_ALLOW_DEMO_RESET", "")
	_ = os.Unsetenv("APP_ENV")
	_ = os.Unsetenv("APP_HTTP_ADDR")
	_ = os.Unsetenv("APP_LOG_LEVEL")
	_ = os.Unsetenv("APP_SEED_LOCAL")
	_ = os.Unsetenv("APP_ALLOW_EVENT_INGEST")
	_ = os.Unsetenv("APP_ALLOW_DEMO_RESET")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.HTTPAddr != ":8080" {
		t.Fatalf("addr=%s", cfg.HTTPAddr)
	}
	if cfg.ServiceName != "cloudops-api" {
		t.Fatalf("service=%s", cfg.ServiceName)
	}
	if !cfg.SeedLocalData {
		t.Fatal("local default should seed")
	}
	if !cfg.AllowEventIngest {
		t.Fatal("local default should allow event ingestion")
	}
	if !cfg.AllowDemoReset {
		t.Fatal("local default should allow bounded demo reset")
	}
}

func TestLoadRejectsBadLogLevel(t *testing.T) {
	t.Setenv("APP_LOG_LEVEL", "verbose")
	if _, err := Load(); err == nil {
		t.Fatal("expected error")
	}
}

func TestCORSSplit(t *testing.T) {
	t.Setenv("APP_LOG_LEVEL", "info")
	t.Setenv("APP_CORS_ORIGINS", "http://localhost:4200, https://example.com")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(cfg.CORSOrigins) != 2 {
		t.Fatalf("origins=%v", cfg.CORSOrigins)
	}
}

func TestEventIngestIsSafeByDefaultOutsideLocal(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("APP_ALLOW_EVENT_INGEST", "")
	t.Setenv("APP_ALLOW_DEMO_RESET", "")
	_ = os.Unsetenv("APP_ALLOW_EVENT_INGEST")
	_ = os.Unsetenv("APP_ALLOW_DEMO_RESET")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.AllowEventIngest {
		t.Fatal("non-local runtime must not allow unauthenticated event ingestion by default")
	}
	if cfg.AllowDemoReset {
		t.Fatal("non-local runtime must not allow demo reset by default")
	}
}

func TestDemoResetRequiresExplicitNonLocalOverride(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("APP_ALLOW_DEMO_RESET", "true")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.AllowDemoReset {
		t.Fatal("explicit controlled-runtime reset override was ignored")
	}
}

func TestEventIngestCanBeExplicitlyEnabledForControlledRuntime(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("APP_ALLOW_EVENT_INGEST", "true")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.AllowEventIngest {
		t.Fatal("explicit controlled-runtime override was ignored")
	}
}
