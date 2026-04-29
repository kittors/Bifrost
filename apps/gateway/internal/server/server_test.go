package server_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestRequestIDMiddlewareUsesIncomingHeaderAndGeneratesWhenMissing(t *testing.T) {
	t.Parallel()

	app := server.New(server.Options{
		ReadyCheck: func(ctx context.Context) error { return nil },
		ReadyTime:  "2026-04-17T12:00:00Z",
		Upstreams:  map[string]string{},
		RequestID: func() string {
			return "req_generated_01"
		},
	})

	withHeader := httptest.NewRecorder()
	withHeaderRequest := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	withHeaderRequest.Header.Set("X-Request-Id", "req_incoming_01")
	app.Handler().ServeHTTP(withHeader, withHeaderRequest)

	if got := withHeader.Header().Get("X-Request-Id"); got != "req_incoming_01" {
		t.Fatalf("expected incoming request id to be preserved, got %q", got)
	}

	withoutHeader := httptest.NewRecorder()
	withoutHeaderRequest := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	app.Handler().ServeHTTP(withoutHeader, withoutHeaderRequest)

	if got := withoutHeader.Header().Get("X-Request-Id"); got != "req_generated_01" {
		t.Fatalf("expected generated request id, got %q", got)
	}
}

func TestRecoveryMiddlewareReturnsUnifiedAPIErrorForPanics(t *testing.T) {
	t.Parallel()

	app := server.New(server.Options{
		ReadyCheck: func(ctx context.Context) error { return nil },
		ReadyTime:  "2026-04-17T12:00:00Z",
		Upstreams:  map[string]string{},
		RequestID: func() string {
			return "req_panic_01"
		},
		RegisterRoutes: func(mux *http.ServeMux, _ *server.App) {
			mux.HandleFunc("/api/v1/test/panic", func(http.ResponseWriter, *http.Request) {
				panic("boom")
			})
		},
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/test/panic", nil)
	app.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", recorder.Code)
	}

	var payload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal panic payload: %v", err)
	}

	if payload["success"] != false {
		t.Fatalf("expected success false, got %#v", payload["success"])
	}
}

func TestAccessLogMiddlewareLogsRequests(t *testing.T) {
	t.Parallel()

	var builder strings.Builder
	logger := slog.New(slog.NewTextHandler(&builder, nil))
	app := server.New(server.Options{
		ReadyCheck: func(ctx context.Context) error { return nil },
		ReadyTime:  "2026-04-17T12:00:00Z",
		Upstreams:  map[string]string{},
		RequestID: func() string {
			return "req_log_01"
		},
		Logger: logger,
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	app.Handler().ServeHTTP(recorder, request)

	logOutput := builder.String()
	if !strings.Contains(logOutput, "req_log_01") {
		t.Fatalf("expected access log to contain request id, got %q", logOutput)
	}

	if !strings.Contains(logOutput, "/healthz") {
		t.Fatalf("expected access log to contain path, got %q", logOutput)
	}
}
