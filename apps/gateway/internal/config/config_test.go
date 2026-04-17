package config_test

import (
	"testing"
	"time"

	"github.com/kittors/bifrost/apps/gateway/internal/config"
)

func TestLoadDefaults(t *testing.T) {
	t.Parallel()

	cfg, err := config.Load(nil)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.ListenAddress != ":8080" {
		t.Fatalf("expected default listen address :8080, got %s", cfg.ListenAddress)
	}

	if cfg.DatabaseURL != "postgres://bifrost:bifrost@127.0.0.1:5432/bifrost?sslmode=disable" {
		t.Fatalf("unexpected default database URL: %s", cfg.DatabaseURL)
	}

	if cfg.Upstreams["gitlab"] != "http://mock-gitlab:8080" {
		t.Fatalf("unexpected gitlab upstream: %s", cfg.Upstreams["gitlab"])
	}

	if cfg.AccessTokenTTL != 15*time.Minute {
		t.Fatalf("expected default access token ttl 15m, got %s", cfg.AccessTokenTTL)
	}

	if cfg.RefreshTokenTTL != 7*24*time.Hour {
		t.Fatalf("expected default refresh token ttl 168h, got %s", cfg.RefreshTokenTTL)
	}

	if cfg.TokenSecret == "" {
		t.Fatal("expected default token secret")
	}
}

func TestLoadFromEnvironment(t *testing.T) {
	t.Parallel()

	cfg, err := config.Load(map[string]string{
		"PORT":                            "19090",
		"BIFROST_DATABASE_URL":            "postgres://example",
		"BIFROST_UPSTREAM_GITLAB":         "http://gitlab.internal:8080",
		"BIFROST_UPSTREAM_JENKINS":        "http://jenkins.internal:8080",
		"BIFROST_UPSTREAM_DOCS":           "http://docs.internal:8080",
		"BIFROST_UPSTREAM_INTERNAL_ADMIN": "http://internal-admin.internal:8080",
		"BIFROST_TOKEN_SECRET":            "test-secret-32-bytes-for-config",
		"BIFROST_ACCESS_TOKEN_TTL":        "20m",
		"BIFROST_REFRESH_TOKEN_TTL":       "240h",
	})
	if err != nil {
		t.Fatalf("load config from environment: %v", err)
	}

	if cfg.ListenAddress != ":19090" {
		t.Fatalf("expected listen address :19090, got %s", cfg.ListenAddress)
	}

	if cfg.DatabaseURL != "postgres://example" {
		t.Fatalf("unexpected database URL: %s", cfg.DatabaseURL)
	}

	if cfg.Upstreams["internal-admin"] != "http://internal-admin.internal:8080" {
		t.Fatalf("unexpected internal-admin upstream: %s", cfg.Upstreams["internal-admin"])
	}

	if cfg.TokenSecret != "test-secret-32-bytes-for-config" {
		t.Fatalf("unexpected token secret: %s", cfg.TokenSecret)
	}

	if cfg.AccessTokenTTL != 20*time.Minute {
		t.Fatalf("expected access token ttl 20m, got %s", cfg.AccessTokenTTL)
	}

	if cfg.RefreshTokenTTL != 240*time.Hour {
		t.Fatalf("expected refresh token ttl 240h, got %s", cfg.RefreshTokenTTL)
	}
}
