package config

import (
	"os"
	"strings"
)

type Config struct {
	DatabaseURL   string
	ListenAddress string
	Upstreams     map[string]string
}

func Load(overrides map[string]string) (Config, error) {
	cfg := Config{
		DatabaseURL:   lookup(overrides, "BIFROST_DATABASE_URL", "postgres://bifrost:bifrost@127.0.0.1:5432/bifrost?sslmode=disable"),
		ListenAddress: ":" + lookup(overrides, "PORT", "8080"),
		Upstreams: map[string]string{
			"gitlab":         lookup(overrides, "BIFROST_UPSTREAM_GITLAB", "http://mock-gitlab:8080"),
			"jenkins":        lookup(overrides, "BIFROST_UPSTREAM_JENKINS", "http://mock-jenkins:8080"),
			"docs":           lookup(overrides, "BIFROST_UPSTREAM_DOCS", "http://mock-docs:8080"),
			"internal-admin": lookup(overrides, "BIFROST_UPSTREAM_INTERNAL_ADMIN", "http://mock-internal-admin:8080"),
		},
	}

	return cfg, nil
}

func lookup(overrides map[string]string, key string, fallback string) string {
	if overrides != nil {
		if value, ok := overrides[key]; ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}

	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}

	return fallback
}
