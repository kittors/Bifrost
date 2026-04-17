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

func TestAdminLoginReturnsEnvelope(t *testing.T) {
	t.Parallel()

	stub := &stubAuthService{
		adminLoginResult: auth.LoginResult{
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
			ExpiresIn:    900,
			User: auth.LoginUser{
				ID:          "user_admin",
				Username:    "admin",
				DisplayName: "Administrator",
				Roles:       []string{"role_admin"},
			},
		},
	}

	app := server.New(server.Options{
		ReadyCheck: func(ctx context.Context) error {
			return nil
		},
		ReadyTime: "2026-04-17T12:00:00Z",
		Upstreams: map[string]string{},
		Now: func() time.Time {
			return time.Date(2026, time.April, 17, 13, 45, 0, 0, time.UTC)
		},
		RequestID: func() string {
			return "req_test_01"
		},
		AuthService: stub,
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/admin/auth/login",
		strings.NewReader(`{"username":"admin","password":"correct horse battery staple"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	app.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}

	var payload apiEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}

	if !payload.Success {
		t.Fatal("expected success response")
	}

	if payload.Meta.RequestID != "req_test_01" {
		t.Fatalf("expected request id req_test_01, got %q", payload.Meta.RequestID)
	}

	if payload.Meta.Timestamp != "2026-04-17T13:45:00Z" {
		t.Fatalf("expected timestamp 2026-04-17T13:45:00Z, got %q", payload.Meta.Timestamp)
	}

	if payload.Error != nil {
		t.Fatalf("expected nil error, got %#v", payload.Error)
	}

	var data loginResponse
	if err := json.Unmarshal(payload.Data, &data); err != nil {
		t.Fatalf("unmarshal login data: %v", err)
	}

	if data.User.ID != "user_admin" {
		t.Fatalf("expected user_admin payload, got %#v", data)
	}

	if stub.adminLoginInput.Username != "admin" {
		t.Fatalf("expected admin username, got %q", stub.adminLoginInput.Username)
	}
}

func TestClientLoginReturnsErrorEnvelope(t *testing.T) {
	t.Parallel()

	stub := &stubAuthService{
		clientLoginError: &auth.ServiceError{
			StatusCode:  http.StatusForbidden,
			Code:        contracts.ErrorCodeDeviceNotTrusted,
			Message:     "device not found for user",
			UserMessage: "当前设备未被信任",
		},
	}

	app := server.New(server.Options{
		ReadyCheck: func(ctx context.Context) error {
			return nil
		},
		ReadyTime: "2026-04-17T12:00:00Z",
		Upstreams: map[string]string{},
		Now: func() time.Time {
			return time.Date(2026, time.April, 17, 13, 45, 0, 0, time.UTC)
		},
		RequestID: func() string {
			return "req_test_02"
		},
		AuthService: stub,
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/client/auth/login",
		strings.NewReader(`{"username":"alice","password":"correct horse battery staple","deviceId":"device_alice_01","clientVersion":"1.0.0"}`),
	)
	request.Header.Set("Content-Type", "application/json")
	app.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", recorder.Code)
	}

	var payload apiEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}

	if payload.Success {
		t.Fatal("expected failed response")
	}

	if string(payload.Data) != "null" {
		t.Fatalf("expected nil data, got %#v", payload.Data)
	}

	if payload.Error == nil {
		t.Fatal("expected error payload")
	}

	if payload.Error.Code != string(contracts.ErrorCodeDeviceNotTrusted) {
		t.Fatalf("expected device not trusted code, got %q", payload.Error.Code)
	}

	if payload.Meta.RequestID != "req_test_02" {
		t.Fatalf("expected request id req_test_02, got %q", payload.Meta.RequestID)
	}
}

func TestAdminRefreshLogoutAndMeRoutes(t *testing.T) {
	t.Parallel()

	stub := &stubAuthService{
		refreshResult: auth.LoginResult{
			AccessToken:  "new-access-token",
			RefreshToken: "new-refresh-token",
			ExpiresIn:    900,
			User: auth.LoginUser{
				ID:          "user_admin",
				Username:    "admin",
				DisplayName: "Administrator",
				Roles:       []string{"role_admin"},
			},
		},
		currentUser: auth.LoginUser{
			ID:          "user_admin",
			Username:    "admin",
			DisplayName: "Administrator",
			Roles:       []string{"role_admin"},
		},
	}

	app := server.New(server.Options{
		ReadyCheck: func(ctx context.Context) error {
			return nil
		},
		ReadyTime: "2026-04-17T12:00:00Z",
		Upstreams: map[string]string{},
		Now: func() time.Time {
			return time.Date(2026, time.April, 17, 13, 45, 0, 0, time.UTC)
		},
		RequestID: func() string {
			return "req_auth_routes"
		},
		AuthService: stub,
	})

	refreshRecorder := httptest.NewRecorder()
	refreshRequest := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/admin/auth/refresh",
		strings.NewReader(`{"refreshToken":"old-refresh-token"}`),
	)
	refreshRequest.Header.Set("Content-Type", "application/json")
	app.Handler().ServeHTTP(refreshRecorder, refreshRequest)

	if refreshRecorder.Code != http.StatusOK {
		t.Fatalf("expected refresh status 200, got %d", refreshRecorder.Code)
	}

	if stub.refreshInput.RefreshToken != "old-refresh-token" {
		t.Fatalf("expected refresh token to be forwarded, got %q", stub.refreshInput.RefreshToken)
	}

	logoutRecorder := httptest.NewRecorder()
	logoutRequest := httptest.NewRequest(http.MethodPost, "/api/v1/admin/auth/logout", nil)
	logoutRequest.Header.Set("Authorization", "Bearer access-token")
	app.Handler().ServeHTTP(logoutRecorder, logoutRequest)

	if logoutRecorder.Code != http.StatusOK {
		t.Fatalf("expected logout status 200, got %d", logoutRecorder.Code)
	}

	if stub.logoutInput.AccessToken != "access-token" {
		t.Fatalf("expected logout access token to be forwarded, got %q", stub.logoutInput.AccessToken)
	}

	meRecorder := httptest.NewRecorder()
	meRequest := httptest.NewRequest(http.MethodGet, "/api/v1/admin/auth/me", nil)
	meRequest.Header.Set("Authorization", "Bearer access-token")
	app.Handler().ServeHTTP(meRecorder, meRequest)

	if meRecorder.Code != http.StatusOK {
		t.Fatalf("expected me status 200, got %d", meRecorder.Code)
	}

	if stub.currentUserInput.AccessToken != "access-token" {
		t.Fatalf("expected me access token to be forwarded, got %q", stub.currentUserInput.AccessToken)
	}
}

type stubAuthService struct {
	adminLoginInput  auth.AdminLoginInput
	adminLoginResult auth.LoginResult
	adminLoginError  error

	clientLoginInput  auth.ClientLoginInput
	clientLoginResult auth.LoginResult
	clientLoginError  error

	refreshInput  auth.RefreshInput
	refreshResult auth.LoginResult
	refreshError  error

	logoutInput auth.LogoutInput
	logoutError error

	currentUserInput auth.CurrentUserInput
	currentUser      auth.LoginUser
	currentUserError error
}

func (s *stubAuthService) AdminLogin(_ context.Context, input auth.AdminLoginInput) (auth.LoginResult, error) {
	s.adminLoginInput = input
	return s.adminLoginResult, s.adminLoginError
}

func (s *stubAuthService) ClientLogin(_ context.Context, input auth.ClientLoginInput) (auth.LoginResult, error) {
	s.clientLoginInput = input
	return s.clientLoginResult, s.clientLoginError
}

func (s *stubAuthService) RefreshSession(_ context.Context, input auth.RefreshInput) (auth.LoginResult, error) {
	s.refreshInput = input
	return s.refreshResult, s.refreshError
}

func (s *stubAuthService) Logout(_ context.Context, input auth.LogoutInput) error {
	s.logoutInput = input
	return s.logoutError
}

func (s *stubAuthService) CurrentUser(_ context.Context, input auth.CurrentUserInput) (auth.LoginUser, error) {
	s.currentUserInput = input
	return s.currentUser, s.currentUserError
}

type apiEnvelope struct {
	Success bool              `json:"success"`
	Data    json.RawMessage   `json:"data"`
	Meta    envelopeMeta      `json:"meta"`
	Error   *envelopeAPIError `json:"error"`
}

type loginResponse struct {
	AccessToken  string         `json:"accessToken"`
	RefreshToken string         `json:"refreshToken"`
	ExpiresIn    int            `json:"expiresIn"`
	User         loginUserShape `json:"user"`
}

type loginUserShape struct {
	ID          string   `json:"id"`
	Username    string   `json:"username"`
	DisplayName string   `json:"displayName"`
	Roles       []string `json:"roles"`
}

type envelopeMeta struct {
	RequestID string `json:"requestId"`
	Timestamp string `json:"timestamp"`
}

type envelopeAPIError struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	UserMessage string `json:"userMessage"`
}
