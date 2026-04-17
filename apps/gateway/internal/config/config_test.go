package config_test

import (
	"testing"

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
}
