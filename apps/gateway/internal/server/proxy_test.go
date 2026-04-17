package server_test

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kittors/bifrost/apps/gateway/internal/auth"
	"github.com/kittors/bifrost/apps/gateway/internal/contracts"
	"github.com/kittors/bifrost/apps/gateway/internal/server"
)

func TestServiceProxyForwardsRewrittenPathQueryAndDebugHeaders(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_ = json.NewEncoder(writer).Encode(map[string]any{
			"path":           request.URL.Path,
			"query":          request.URL.RawQuery,
			"bifrostRequest": request.Header.Get("X-Bifrost-Request-Id"),
			"bifrostService": request.Header.Get("X-Bifrost-Service-Key"),
			"bifrostUser":    request.Header.Get("X-Bifrost-User-Id"),
		})
	}))
	defer upstream.Close()

	stub := &stubAuthService{
		resolveProxyResult: auth.ResolveProxyRequestResult{
			ServiceID:    "service_gitlab",
			ServiceKey:   "gitlab",
			UpstreamURL:  upstream.URL,
			UserID:       "user_alice",
			AccessSource: "role",
		},
	}
	app := server.New(server.Options{
		ReadyCheck:  func(context.Context) error { return nil },
		ReadyTime:   "2026-04-17T12:00:00Z",
		Upstreams:   map[string]string{},
		RequestID:   func() string { return "req_proxy_01" },
		AuthService: stub,
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/s/gitlab/api/v4/projects?visibility=private", nil)
	request.Header.Set("Authorization", "Bearer access-token")
	app.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected proxy status 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if stub.resolveProxyInput.ServiceKey != "gitlab" {
		t.Fatalf("expected service key gitlab, got %q", stub.resolveProxyInput.ServiceKey)
	}
	if stub.resolveProxyInput.AccessToken != "access-token" {
		t.Fatalf("expected access token to be used for proxy authorization, got %q", stub.resolveProxyInput.AccessToken)
	}
	if stub.resolveProxyInput.RequestID != "req_proxy_01" {
		t.Fatalf("expected request id req_proxy_01, got %q", stub.resolveProxyInput.RequestID)
	}

	var payload map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal upstream payload: %v", err)
	}
	if payload["path"] != "/api/v4/projects" {
		t.Fatalf("expected rewritten path /api/v4/projects, got %q", payload["path"])
	}
	if payload["query"] != "visibility=private" {
		t.Fatalf("expected query visibility=private, got %q", payload["query"])
	}
	if payload["bifrostRequest"] != "req_proxy_01" || payload["bifrostService"] != "gitlab" || payload["bifrostUser"] != "user_alice" {
		t.Fatalf("expected debug headers to be injected, got %#v", payload)
	}
	if stub.recordProxyAccessEventInput.Type != "service.access.granted" {
		t.Fatalf("expected service access granted audit event, got %#v", stub.recordProxyAccessEventInput)
	}
	if stub.recordProxyAccessEventInput.RequestID != "req_proxy_01" {
		t.Fatalf("expected audit request id req_proxy_01, got %q", stub.recordProxyAccessEventInput.RequestID)
	}
}

func TestServiceProxyAcceptsShortLivedAccessCookie(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_ = json.NewEncoder(writer).Encode(map[string]string{
			"serviceKey": request.Header.Get("X-Bifrost-Service-Key"),
		})
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
		ReadyCheck:  func(context.Context) error { return nil },
		ReadyTime:   "2026-04-17T12:00:00Z",
		Upstreams:   map[string]string{},
		RequestID:   func() string { return "req_proxy_cookie" },
		AuthService: stub,
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/s/gitlab/whoami", nil)
	request.AddCookie(&http.Cookie{Name: "bifrost_access_ticket", Value: "ticket_gitlab_01"})
	app.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected proxy status 200, got %d body=%s", recorder.Code, recorder.Body.String())
	}
	if stub.resolveProxyInput.AccessToken != "" {
		t.Fatalf("expected access token to be empty when cookie is used, got %q", stub.resolveProxyInput.AccessToken)
	}
	if stub.resolveProxyInput.AccessTicket != "ticket_gitlab_01" {
		t.Fatalf("expected access ticket from cookie, got %q", stub.resolveProxyInput.AccessTicket)
	}
}

func TestServiceProxySupportsWebSocketUpgrade(t *testing.T) {
	t.Parallel()

	upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if !strings.EqualFold(request.Header.Get("Upgrade"), "websocket") {
			t.Fatalf("expected websocket upgrade header, got %q", request.Header.Get("Upgrade"))
		}
		if request.URL.Path != "/-/cable" {
			t.Fatalf("expected rewritten websocket path /-/cable, got %q", request.URL.Path)
		}

		hijacker, ok := writer.(http.Hijacker)
		if !ok {
			t.Fatal("upstream response writer does not support hijacking")
		}
		conn, _, err := hijacker.Hijack()
		if err != nil {
			t.Fatalf("hijack upstream connection: %v", err)
		}
		defer conn.Close()

		_, _ = conn.Write([]byte("HTTP/1.1 101 Switching Protocols\r\nConnection: Upgrade\r\nUpgrade: websocket\r\n\r\n"))
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
		ReadyCheck:  func(context.Context) error { return nil },
		ReadyTime:   "2026-04-17T12:00:00Z",
		Upstreams:   map[string]string{},
		RequestID:   func() string { return "req_proxy_ws" },
		AuthService: stub,
	})
	proxyServer := httptest.NewServer(app.Handler())
	defer proxyServer.Close()

	proxyURL := strings.TrimPrefix(proxyServer.URL, "http://")
	conn, err := net.Dial("tcp", proxyURL)
	if err != nil {
		t.Fatalf("dial proxy server: %v", err)
	}
	defer conn.Close()

	_, err = fmt.Fprintf(
		conn,
		"GET /s/gitlab/-/cable HTTP/1.1\r\nHost: %s\r\nConnection: Upgrade\r\nUpgrade: websocket\r\nSec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\nSec-WebSocket-Version: 13\r\nCookie: bifrost_access_ticket=ticket_ws_01\r\n\r\n",
		proxyURL,
	)
	if err != nil {
		t.Fatalf("write websocket request: %v", err)
	}

	response, err := http.ReadResponse(bufio.NewReader(conn), nil)
	if err != nil {
		t.Fatalf("read websocket response: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusSwitchingProtocols {
		t.Fatalf("expected websocket status 101, got %d", response.StatusCode)
	}
	if stub.resolveProxyInput.AccessTicket != "ticket_ws_01" {
		t.Fatalf("expected websocket access ticket from cookie, got %q", stub.resolveProxyInput.AccessTicket)
	}
	if stub.recordProxyAccessEventInput.Type != "service.access.granted" {
		t.Fatalf("expected access granted audit event for websocket, got %#v", stub.recordProxyAccessEventInput)
	}
}

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
