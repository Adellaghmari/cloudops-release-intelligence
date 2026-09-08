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
	_ = os.Unsetenv("APP_ENV")
	_ = os.Unsetenv("APP_HTTP_ADDR")
	_ = os.Unsetenv("APP_LOG_LEVEL")
	_ = os.Unsetenv("APP_SEED_LOCAL")
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
