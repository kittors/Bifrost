package server_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kittors/bifrost/apps/gateway/internal/auth"
	"github.com/kittors/bifrost/apps/gateway/internal/contracts"
	"github.com/kittors/bifrost/apps/gateway/internal/server"
)

func TestServiceProxyMapsUpstreamTimeout(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		time.Sleep(50 * time.Millisecond)
		writer.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	stub := &stubAuthService{
		resolveProxyResult: auth.ResolveProxyRequestResult{
			ServiceID:   "service_gitlab",
			ServiceKey:  "gitlab",
			UpstreamURL: upstream.URL,
			UserID:      "user_alice",
		},
	}
	app := server.New(server.Options{
		ReadyCheck:   func(context.Context) error { return nil },
		ReadyTime:    "2026-04-17T12:00:00Z",
		Upstreams:    map[string]string{},
		RequestID:    func() string { return "req_proxy_timeout" },
		ProxyTimeout: 1 * time.Millisecond,
		AuthService:  stub,
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/s/gitlab/slow", nil)
	request.Header.Set("Authorization", "Bearer access-token")
	app.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusGatewayTimeout {
		t.Fatalf("expected 504, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	assertAPIErrorCode(t, recorder.Body.String(), string(contracts.ErrorCodeGatewayUpstreamTimeout))
}

func TestServiceProxyMapsUpstreamUnavailable(t *testing.T) {
	t.Parallel()

	stub := &stubAuthService{
		resolveProxyResult: auth.ResolveProxyRequestResult{
			ServiceID:   "service_gitlab",
			ServiceKey:  "gitlab",
			UpstreamURL: "http://127.0.0.1:1",
			UserID:      "user_alice",
		},
	}
	app := server.New(server.Options{
		ReadyCheck:  func(context.Context) error { return nil },
		ReadyTime:   "2026-04-17T12:00:00Z",
		Upstreams:   map[string]string{},
		RequestID:   func() string { return "req_proxy_bad_upstream" },
		AuthService: stub,
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/s/gitlab/whoami", nil)
	request.Header.Set("Authorization", "Bearer access-token")
	app.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	assertAPIErrorCode(t, recorder.Body.String(), string(contracts.ErrorCodeGatewayBadUpstream))
	if stub.recordProxyAccessEventInput.Type != "service.access.upstream_error" {
		t.Fatalf("expected upstream error audit event, got %#v", stub.recordProxyAccessEventInput)
	}
	if stub.recordProxyAccessEventInput.Result != "failure" {
		t.Fatalf("expected failure audit result, got %q", stub.recordProxyAccessEventInput.Result)
	}
}

func TestServiceProxyRejectsOversizedBody(t *testing.T) {
	t.Parallel()

	stub := &stubAuthService{
		resolveProxyResult: auth.ResolveProxyRequestResult{
			ServiceID:   "service_gitlab",
			ServiceKey:  "gitlab",
			UpstreamURL: "http://127.0.0.1:1",
			UserID:      "user_alice",
		},
	}
	app := server.New(server.Options{
		ReadyCheck:        func(context.Context) error { return nil },
		ReadyTime:         "2026-04-17T12:00:00Z",
		Upstreams:         map[string]string{},
		RequestID:         func() string { return "req_proxy_too_large" },
		MaxProxyBodyBytes: 4,
		AuthService:       stub,
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/s/gitlab/echo", strings.NewReader("too-large"))
	request.Header.Set("Authorization", "Bearer access-token")
	app.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	assertAPIErrorCode(t, recorder.Body.String(), string(contracts.ErrorCodeGatewayRequestTooLarge))
}

func TestServiceProxyMapsAuthorizationErrors(t *testing.T) {
	t.Parallel()

	stub := &stubAuthService{
		resolveProxyError: &auth.ServiceError{
			StatusCode:  http.StatusForbidden,
			Code:        contracts.ErrorCodePolicyAccessDenied,
			Message:     "user is not allowed to access service",
			UserMessage: "你没有访问该服务的权限",
		},
	}
	app := server.New(server.Options{
		ReadyCheck:  func(context.Context) error { return nil },
		ReadyTime:   "2026-04-17T12:00:00Z",
		Upstreams:   map[string]string{},
		RequestID:   func() string { return "req_proxy_denied" },
		AuthService: stub,
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/s/gitlab/whoami", nil)
	request.Header.Set("Authorization", "Bearer access-token")
	app.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	assertAPIErrorCode(t, recorder.Body.String(), string(contracts.ErrorCodePolicyAccessDenied))
}

func assertAPIErrorCode(t *testing.T, body string, expected string) {
	t.Helper()

	var payload apiEnvelope
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		t.Fatalf("unmarshal API error: %v", err)
	}
	if payload.Error == nil {
		t.Fatal("expected error payload")
	}
	if payload.Error.Code != expected {
		t.Fatalf("expected error code %s, got %q", expected, payload.Error.Code)
	}
}
