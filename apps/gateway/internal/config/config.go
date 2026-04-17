package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	DatabaseURL     string
	ListenAddress   string
	Upstreams       map[string]string
	TokenSecret     string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

func Load(overrides map[string]string) (Config, error) {
	accessTokenTTL, err := lookupDuration(overrides, "BIFROST_ACCESS_TOKEN_TTL", 15*time.Minute)
	if err != nil {
		return Config{}, err
	}

	refreshTokenTTL, err := lookupDuration(overrides, "BIFROST_REFRESH_TOKEN_TTL", 7*24*time.Hour)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		DatabaseURL:     lookup(overrides, "BIFROST_DATABASE_URL", "postgres://bifrost:bifrost@127.0.0.1:5432/bifrost?sslmode=disable"),
		ListenAddress:   ":" + lookup(overrides, "PORT", "8080"),
		TokenSecret:     lookup(overrides, "BIFROST_TOKEN_SECRET", "dev-only-bifrost-token-secret-change-me-32bytes"),
		AccessTokenTTL:  accessTokenTTL,
		RefreshTokenTTL: refreshTokenTTL,
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

func lookupDuration(overrides map[string]string, key string, fallback time.Duration) (time.Duration, error) {
	value := lookup(overrides, key, "")
	if value == "" {
		return fallback, nil
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s duration: %w", key, err)
	}

	return duration, nil
}
