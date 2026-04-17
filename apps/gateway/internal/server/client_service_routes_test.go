package server_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kittors/bifrost/apps/gateway/internal/auth"
	"github.com/kittors/bifrost/apps/gateway/internal/server"
)

// 客户端服务路由测试关注列表、详情、访问 URL 与 Cookie 行为。

func TestClientServiceRoutes(t *testing.T) {
	t.Parallel()

	stub := &stubAuthService{
		clientServices: []auth.ClientService{
			{
				ID:           "service_gitlab",
				Key:          "gitlab",
				Name:         "GitLab",
				Description:  "Source code",
				Group:        "engineering",
				Status:       "enabled",
				AccessSource: "role",
			},
		},
		clientService: auth.ClientService{
			ID:           "service_gitlab",
			Key:          "gitlab",
			Name:         "GitLab",
			Description:  "Source code",
			Group:        "engineering",
			Status:       "enabled",
			AccessSource: "role",
		},
		serviceAccessURL: auth.ServiceAccessURLResult{
			PublicPath:   "/s/gitlab",
			ExpiresIn:    300,
			AccessTicket: "ticket_gitlab_01",
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
			return "req_service_routes"
		},
		AuthService: stub,
	})

	listRecorder := httptest.NewRecorder()
	listRequest := httptest.NewRequest(http.MethodGet, "/api/v1/client/services?keyword=git&group=engineering", nil)
	listRequest.Header.Set("Authorization", "Bearer access-token")
	app.Handler().ServeHTTP(listRecorder, listRequest)

	if listRecorder.Code != http.StatusOK {
		t.Fatalf("expected list status 200, got %d", listRecorder.Code)
	}

	if stub.listClientServicesInput.Keyword != "git" {
		t.Fatalf("expected list keyword git, got %q", stub.listClientServicesInput.Keyword)
	}

	detailRecorder := httptest.NewRecorder()
	detailRequest := httptest.NewRequest(http.MethodGet, "/api/v1/client/services/service_gitlab", nil)
	detailRequest.Header.Set("Authorization", "Bearer access-token")
	app.Handler().ServeHTTP(detailRecorder, detailRequest)

	if detailRecorder.Code != http.StatusOK {
		t.Fatalf("expected detail status 200, got %d", detailRecorder.Code)
	}

	if stub.getClientServiceInput.ServiceID != "service_gitlab" {
		t.Fatalf("expected detail service id forwarded, got %q", stub.getClientServiceInput.ServiceID)
	}

	accessRecorder := httptest.NewRecorder()
	accessRequest := httptest.NewRequest(http.MethodPost, "/api/v1/client/services/service_gitlab/access-url", nil)
	accessRequest.Header.Set("Authorization", "Bearer access-token")
	accessRequest.Host = "127.0.0.1:18080"
	app.Handler().ServeHTTP(accessRecorder, accessRequest)

	if accessRecorder.Code != http.StatusOK {
		t.Fatalf("expected access-url status 200, got %d", accessRecorder.Code)
	}

	if stub.createServiceAccessURLInput.ServiceID != "service_gitlab" {
		t.Fatalf("expected access-url service id forwarded, got %q", stub.createServiceAccessURLInput.ServiceID)
	}

	cookie := accessRecorder.Result().Cookies()
	if len(cookie) != 1 {
		t.Fatalf("expected one access cookie, got %d", len(cookie))
	}
	if cookie[0].Name != "bifrost_access_ticket" || cookie[0].Value != "ticket_gitlab_01" {
		t.Fatalf("unexpected access cookie: %#v", cookie[0])
	}
	if cookie[0].Path != "/s/gitlab" || !cookie[0].HttpOnly {
		t.Fatalf("expected HttpOnly cookie scoped to service path, got %#v", cookie[0])
	}
	if cookie[0].SameSite != http.SameSiteLaxMode {
		t.Fatalf("expected SameSite=Lax, got %v", cookie[0].SameSite)
	}
	if cookie[0].Secure {
		t.Fatal("expected non-secure cookie on plain http request")
	}

	secureRecorder := httptest.NewRecorder()
	secureRequest := httptest.NewRequest(http.MethodPost, "/api/v1/client/services/service_gitlab/access-url", nil)
	secureRequest.Header.Set("Authorization", "Bearer access-token")
	secureRequest.Header.Set("X-Forwarded-Proto", "https")
	secureRequest.Host = "bifrost.example.com"
	app.Handler().ServeHTTP(secureRecorder, secureRequest)

	secureCookies := secureRecorder.Result().Cookies()
	if len(secureCookies) != 1 || !secureCookies[0].Secure {
		t.Fatalf("expected secure cookie on forwarded https request, got %#v", secureCookies)
	}
}
