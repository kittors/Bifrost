package server_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kittors/bifrost/apps/gateway/internal/auth"
	"github.com/kittors/bifrost/apps/gateway/internal/contracts"
	"github.com/kittors/bifrost/apps/gateway/internal/server"
)

// 后台路由测试覆盖用户管理与后台配置接口的参数转发和状态码。

func TestAdminUserRoutes(t *testing.T) {
	t.Parallel()

	stub := &stubAuthService{
		adminUsers: auth.AdminUserListResult{
			Items: []auth.AdminUser{
				{
					ID:          "user_created_01",
					Username:    "charlie",
					DisplayName: "Charlie",
					Email:       "charlie@example.com",
					Status:      "enabled",
					Roles:       []string{"role_developer"},
				},
			},
			Pagination: contracts.Pagination{Page: 1, PageSize: 20, Total: 1, TotalPages: 1},
		},
		createdAdminUser: auth.AdminUser{
			ID:          "user_created_01",
			Username:    "charlie",
			DisplayName: "Charlie",
			Email:       "charlie@example.com",
			Status:      "enabled",
			Roles:       []string{"role_developer"},
		},
		updatedAdminUser: auth.AdminUser{
			ID:          "user_created_01",
			Username:    "charlie",
			DisplayName: "Charles",
			Email:       "charles@example.com",
			Status:      "enabled",
			Roles:       []string{"role_admin"},
		},
		adminUser: auth.AdminUser{
			ID:          "user_created_01",
			Username:    "charlie",
			DisplayName: "Charlie",
			Email:       "charlie@example.com",
			Status:      "enabled",
			Roles:       []string{"role_developer"},
		},
		statusAdminUser: auth.AdminUser{
			ID:          "user_created_01",
			Username:    "charlie",
			DisplayName: "Charlie",
			Email:       "charlie@example.com",
			Status:      "disabled",
			Roles:       []string{"role_developer"},
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
			return "req_admin_user_routes"
		},
		AuthService: stub,
	})

	listRecorder := httptest.NewRecorder()
	listRequest := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users?page=1&pageSize=20&keyword=charlie", nil)
	listRequest.Header.Set("Authorization", "Bearer admin-token")
	app.Handler().ServeHTTP(listRecorder, listRequest)

	if listRecorder.Code != http.StatusOK {
		t.Fatalf("expected list status 200, got %d", listRecorder.Code)
	}

	if stub.listAdminUsersInput.Keyword != "charlie" {
		t.Fatalf("expected list keyword charlie, got %q", stub.listAdminUsersInput.Keyword)
	}

	createRecorder := httptest.NewRecorder()
	createRequest := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/admin/users",
		strings.NewReader(`{"username":"charlie","displayName":"Charlie","email":"charlie@example.com","password":"ChangeMe123!","roleIds":["role_developer"]}`),
	)
	createRequest.Header.Set("Authorization", "Bearer admin-token")
	createRequest.Header.Set("Content-Type", "application/json")
	app.Handler().ServeHTTP(createRecorder, createRequest)

	if createRecorder.Code != http.StatusCreated {
		t.Fatalf("expected create status 201, got %d", createRecorder.Code)
	}

	if stub.createAdminUserInput.Username != "charlie" {
		t.Fatalf("expected create username charlie, got %q", stub.createAdminUserInput.Username)
	}

	updateRecorder := httptest.NewRecorder()
	updateRequest := httptest.NewRequest(
		http.MethodPatch,
		"/api/v1/admin/users/user_created_01",
		strings.NewReader(`{"displayName":"Charles","email":"charles@example.com","roleIds":["role_admin"]}`),
	)
	updateRequest.Header.Set("Authorization", "Bearer admin-token")
	updateRequest.Header.Set("Content-Type", "application/json")
	app.Handler().ServeHTTP(updateRecorder, updateRequest)

	if updateRecorder.Code != http.StatusOK {
		t.Fatalf("expected update status 200, got %d", updateRecorder.Code)
	}

	if stub.updateAdminUserInput.UserID != "user_created_01" {
		t.Fatalf("expected update user id forwarded, got %q", stub.updateAdminUserInput.UserID)
	}

	detailRecorder := httptest.NewRecorder()
	detailRequest := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users/user_created_01", nil)
	detailRequest.Header.Set("Authorization", "Bearer admin-token")
	app.Handler().ServeHTTP(detailRecorder, detailRequest)

	if detailRecorder.Code != http.StatusOK {
		t.Fatalf("expected detail status 200, got %d", detailRecorder.Code)
	}
	if stub.getAdminUserInput.UserID != "user_created_01" {
		t.Fatalf("expected detail user id forwarded, got %q", stub.getAdminUserInput.UserID)
	}

	resetRecorder := httptest.NewRecorder()
	resetRequest := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/admin/users/user_created_01/reset-password",
		strings.NewReader(`{"password":"NewPassword123!"}`),
	)
	resetRequest.Header.Set("Authorization", "Bearer admin-token")
	resetRequest.Header.Set("Content-Type", "application/json")
	app.Handler().ServeHTTP(resetRecorder, resetRequest)

	if resetRecorder.Code != http.StatusOK {
		t.Fatalf("expected reset password status 200, got %d", resetRecorder.Code)
	}
	if stub.resetAdminUserPasswordInput.Password != "NewPassword123!" {
		t.Fatalf("expected reset password forwarded, got %q", stub.resetAdminUserPasswordInput.Password)
	}

	statusRecorder := httptest.NewRecorder()
	statusRequest := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/admin/users/user_created_01/status",
		strings.NewReader(`{"status":"disabled"}`),
	)
	statusRequest.Header.Set("Authorization", "Bearer admin-token")
	statusRequest.Header.Set("Content-Type", "application/json")
	app.Handler().ServeHTTP(statusRecorder, statusRequest)

	if statusRecorder.Code != http.StatusOK {
		t.Fatalf("expected status update 200, got %d", statusRecorder.Code)
	}
	if stub.setAdminUserStatusInput.Status != "disabled" {
		t.Fatalf("expected disabled status forwarded, got %q", stub.setAdminUserStatusInput.Status)
	}
}

func TestAdminConfigRoutes(t *testing.T) {
	t.Parallel()

	stub := &stubAuthService{
		adminRoles: auth.AdminRoleListResult{
			Items:      []auth.AdminRole{{ID: "role_ops", Name: "ops", DisplayName: "Operations"}},
			Pagination: contracts.Pagination{Page: 1, PageSize: 20, Total: 1, TotalPages: 1},
		},
		createdAdminRole: auth.AdminRole{ID: "role_created_01", Name: "qa", DisplayName: "QA"},
		adminServices: auth.AdminServiceListResult{
			Items:      []auth.AdminService{{ID: "service_docs", Key: "docs", Name: "Docs", Status: "enabled"}},
			Pagination: contracts.Pagination{Page: 1, PageSize: 20, Total: 1, TotalPages: 1},
		},
		createdAdminService: auth.AdminService{ID: "service_created_01", Key: "gitlab", Name: "GitLab", Status: "enabled"},
		adminService:        auth.AdminService{ID: "service_docs", Key: "docs", Name: "Docs", Status: "enabled"},
		updatedAdminService: auth.AdminService{ID: "service_docs", Key: "docs", Name: "Docs Portal", Status: "enabled"},
		statusAdminService:  auth.AdminService{ID: "service_docs", Key: "docs", Name: "Docs Portal", Status: "disabled"},
		adminDevices: auth.AdminDeviceListResult{
			Items:      []auth.AdminDevice{{ID: "device_01", UserID: "user_alice", Name: "Alice Mac", Status: "trusted"}},
			Pagination: contracts.Pagination{Page: 1, PageSize: 20, Total: 1, TotalPages: 1},
		},
		adminDevice:       auth.AdminDevice{ID: "device_01", UserID: "user_alice", UserUsername: "alice", Name: "Alice Mac", Status: "trusted"},
		statusAdminDevice: auth.AdminDevice{ID: "device_01", UserID: "user_alice", UserUsername: "alice", Name: "Alice Mac", Status: "disabled"},
		adminAuditEvents: auth.AdminAuditEventListResult{
			Items:      []auth.AdminAuditEvent{{ID: "audit_01", Type: "auth.login.succeeded", Result: "success"}},
			Pagination: contracts.Pagination{Page: 1, PageSize: 20, Total: 1, TotalPages: 1},
		},
		userServiceOverrides: []auth.UserServiceOverride{{ServiceID: "service_docs", Effect: "allow"}},
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
			return "req_admin_config_routes"
		},
		AuthService: stub,
	})

	requests := []struct {
		method string
		path   string
		body   string
		want   int
	}{
		{http.MethodGet, "/api/v1/admin/roles?keyword=ops", "", http.StatusOK},
		{http.MethodPost, "/api/v1/admin/roles", `{"name":"qa","displayName":"QA","description":"Quality"}`, http.StatusCreated},
		{http.MethodGet, "/api/v1/admin/services?group=shared", "", http.StatusOK},
		{http.MethodPost, "/api/v1/admin/services", `{"key":"gitlab","name":"GitLab","description":"Code","group":"engineering","protocol":"http","upstreamUrl":"http://gitlab:8080","publicPath":"/s/gitlab","enabled":true}`, http.StatusCreated},
		{http.MethodGet, "/api/v1/admin/services/service_docs", "", http.StatusOK},
		{http.MethodPatch, "/api/v1/admin/services/service_docs", `{"name":"Docs Portal","description":"Docs","group":"shared","protocol":"http","upstreamUrl":"http://docs:8080","publicPath":"/s/docs"}`, http.StatusOK},
		{http.MethodPost, "/api/v1/admin/services/service_docs/status", `{"status":"disabled"}`, http.StatusOK},
		{http.MethodGet, "/api/v1/admin/devices?userId=user_alice", "", http.StatusOK},
		{http.MethodGet, "/api/v1/admin/devices/device_01", "", http.StatusOK},
		{http.MethodPost, "/api/v1/admin/devices/device_01/status", `{"status":"disabled"}`, http.StatusOK},
		{http.MethodGet, "/api/v1/admin/audit-events?type=auth.login.succeeded", "", http.StatusOK},
		{http.MethodPut, "/api/v1/admin/roles/role_ops/services", `{"serviceIds":["service_docs"]}`, http.StatusOK},
		{http.MethodPut, "/api/v1/admin/users/user_alice/service-overrides", `{"allowServiceIds":["service_docs"],"denyServiceIds":["service_gitlab"]}`, http.StatusOK},
	}

	for _, item := range requests {
		t.Run(item.method+" "+item.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			var body *strings.Reader
			if item.body == "" {
				body = strings.NewReader("")
			} else {
				body = strings.NewReader(item.body)
			}
			request := httptest.NewRequest(item.method, item.path, body)
			request.Header.Set("Authorization", "Bearer admin-token")
			request.Header.Set("Content-Type", "application/json")
			app.Handler().ServeHTTP(recorder, request)
			if recorder.Code != item.want {
				t.Fatalf("expected status %d, got %d with body %s", item.want, recorder.Code, recorder.Body.String())
			}
		})
	}

	if stub.listAdminRolesInput.Keyword != "ops" {
		t.Fatalf("expected role keyword ops, got %q", stub.listAdminRolesInput.Keyword)
	}
	if stub.replaceRoleServicesInput.RoleID != "role_ops" {
		t.Fatalf("expected role service role id role_ops, got %q", stub.replaceRoleServicesInput.RoleID)
	}
	if stub.replaceUserServiceOverridesInput.UserID != "user_alice" {
		t.Fatalf("expected override user id user_alice, got %q", stub.replaceUserServiceOverridesInput.UserID)
	}
	if stub.updateAdminServiceInput.Name != "Docs Portal" {
		t.Fatalf("expected service update name Docs Portal, got %q", stub.updateAdminServiceInput.Name)
	}
	if stub.setAdminServiceStatusInput.Status != "disabled" {
		t.Fatalf("expected service disabled status, got %q", stub.setAdminServiceStatusInput.Status)
	}
	if stub.setAdminDeviceStatusInput.Status != "disabled" {
		t.Fatalf("expected device disabled status, got %q", stub.setAdminDeviceStatusInput.Status)
	}
}
