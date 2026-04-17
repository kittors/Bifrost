package server_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kittors/bifrost/apps/gateway/internal/server"
)

func TestHealthzAndReadyz(t *testing.T) {
	t.Parallel()

	app := server.New(server.Options{
		ReadyCheck: func(ctx context.Context) error {
			return nil
		},
		ReadyTime: "2026-04-17T12:00:00Z",
		Upstreams: map[string]string{
			"gitlab": "http://mock-gitlab:8080",
		},
	})

	healthRecorder := httptest.NewRecorder()
	healthRequest := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	app.Handler().ServeHTTP(healthRecorder, healthRequest)

	if healthRecorder.Code != http.StatusOK {
		t.Fatalf("expected healthz 200, got %d", healthRecorder.Code)
	}

	readyRecorder := httptest.NewRecorder()
	readyRequest := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	app.Handler().ServeHTTP(readyRecorder, readyRequest)

	if readyRecorder.Code != http.StatusOK {
		t.Fatalf("expected readyz 200, got %d", readyRecorder.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(readyRecorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal readyz payload: %v", err)
	}

	if payload["status"] != "ready" {
		t.Fatalf("expected ready status, got %#v", payload["status"])
	}
}

func TestReadyzReturnsServiceUnavailableWhenDependencyFails(t *testing.T) {
	t.Parallel()

	app := server.New(server.Options{
		ReadyCheck: func(ctx context.Context) error {
			return errors.New("database unavailable")
		},
		ReadyTime: "2026-04-17T12:00:00Z",
		Upstreams: map[string]string{},
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	app.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected readyz 503, got %d", recorder.Code)
	}
}
