package server

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/kittors/bifrost/apps/gateway/internal/auth"
)

type contextKey string

const (
	requestIDContextKey     contextKey = "request_id"
	serviceAccessCookieName            = "bifrost_access_ticket"
)

type Options struct {
	ReadyCheck        func(ctx context.Context) error
	ReadyTime         string
	Upstreams         map[string]string
	Now               func() time.Time
	RequestID         func() string
	Logger            *slog.Logger
	RegisterRoutes    func(mux *http.ServeMux, app *App)
	HTTPClient        *http.Client
	ProxyTimeout      time.Duration
	MaxProxyBodyBytes int64
	AuthService       AuthService
}

type App struct {
	handler           http.Handler
	readyCheck        func(ctx context.Context) error
	readyTime         string
	upstreams         map[string]string
	now               func() time.Time
	requestID         func() string
	logger            *slog.Logger
	httpClient        *http.Client
	proxyTimeout      time.Duration
	maxProxyBodyBytes int64
	authService       AuthService
}

type AuthService interface {
	AdminLogin(ctx context.Context, input auth.AdminLoginInput) (auth.LoginResult, error)
	ClientLogin(ctx context.Context, input auth.ClientLoginInput) (auth.LoginResult, error)
	RefreshSession(ctx context.Context, input auth.RefreshInput) (auth.LoginResult, error)
	Logout(ctx context.Context, input auth.LogoutInput) error
	CurrentUser(ctx context.Context, input auth.CurrentUserInput) (auth.LoginUser, error)
	RegisterDevice(ctx context.Context, input auth.RegisterDeviceInput) (auth.DeviceResult, error)
	CreateDeviceChallenge(ctx context.Context, input auth.CreateDeviceChallengeInput) (auth.DeviceChallengeResult, error)
	VerifyDeviceChallenge(ctx context.Context, input auth.VerifyDeviceChallengeInput) (auth.DeviceChallengeVerificationResult, error)
	ListClientServices(ctx context.Context, input auth.ListClientServicesInput) ([]auth.ClientService, error)
	GetClientService(ctx context.Context, input auth.GetClientServiceInput) (auth.ClientService, error)
	CreateServiceAccessURL(ctx context.Context, input auth.CreateServiceAccessURLInput) (auth.ServiceAccessURLResult, error)
	ListAdminUsers(ctx context.Context, input auth.ListAdminUsersInput) (auth.AdminUserListResult, error)
	CreateAdminUser(ctx context.Context, input auth.CreateAdminUserInput) (auth.AdminUser, error)
	UpdateAdminUser(ctx context.Context, input auth.UpdateAdminUserInput) (auth.AdminUser, error)
	GetAdminUser(ctx context.Context, input auth.GetAdminUserInput) (auth.AdminUser, error)
	ResetAdminUserPassword(ctx context.Context, input auth.ResetAdminUserPasswordInput) error
	SetAdminUserStatus(ctx context.Context, input auth.SetAdminUserStatusInput) (auth.AdminUser, error)
	ListAdminRoles(ctx context.Context, input auth.ListAdminRolesInput) (auth.AdminRoleListResult, error)
	CreateAdminRole(ctx context.Context, input auth.CreateAdminRoleInput) (auth.AdminRole, error)
	ListAdminServices(ctx context.Context, input auth.ListAdminServicesInput) (auth.AdminServiceListResult, error)
	CreateAdminService(ctx context.Context, input auth.CreateAdminServiceInput) (auth.AdminService, error)
	ListAdminDevices(ctx context.Context, input auth.ListAdminDevicesInput) (auth.AdminDeviceListResult, error)
	ListAdminAuditEvents(ctx context.Context, input auth.ListAdminAuditEventsInput) (auth.AdminAuditEventListResult, error)
	ReplaceRoleServices(ctx context.Context, input auth.ReplaceRoleServicesInput) error
	ReplaceUserServiceOverrides(ctx context.Context, input auth.ReplaceUserServiceOverridesInput) ([]auth.UserServiceOverride, error)
	ResolveProxyRequest(ctx context.Context, input auth.ResolveProxyRequestInput) (auth.ResolveProxyRequestResult, error)
	RecordProxyAccessEvent(ctx context.Context, input auth.RecordProxyAccessEventInput) error
}

// App 入口文件只负责装配依赖与注册路由，避免继续膨胀成业务大杂烩。
func New(options Options) *App {
	httpClient := options.HTTPClient
	if httpClient == nil {
		httpClient = defaultProxyHTTPClient()
	}

	app := &App{
		readyCheck:        options.ReadyCheck,
		readyTime:         options.ReadyTime,
		upstreams:         options.Upstreams,
		now:               options.Now,
		requestID:         options.RequestID,
		logger:            options.Logger,
		httpClient:        httpClient,
		proxyTimeout:      options.ProxyTimeout,
		maxProxyBodyBytes: options.MaxProxyBodyBytes,
		authService:       options.AuthService,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", app.handleIndex)
	mux.HandleFunc("/healthz", app.handleHealthz)
	mux.HandleFunc("/readyz", app.handleReadyz)
	mux.HandleFunc("/debug/upstreams/", app.handleUpstreamProbe)
	mux.HandleFunc("/s/", app.handleServiceProxy)
	mux.HandleFunc("/api/v1/admin/auth/login", app.handleAdminLogin)
	mux.HandleFunc("/api/v1/admin/auth/refresh", app.handleAdminRefresh)
	mux.HandleFunc("/api/v1/admin/auth/logout", app.handleLogout)
	mux.HandleFunc("/api/v1/admin/auth/me", app.handleCurrentUser)
	mux.HandleFunc("/api/v1/admin/users", app.handleAdminUsers)
	mux.HandleFunc("/api/v1/admin/users/", app.handleAdminUserByID)
	mux.HandleFunc("/api/v1/admin/roles", app.handleAdminRoles)
	mux.HandleFunc("/api/v1/admin/roles/", app.handleAdminRoleByID)
	mux.HandleFunc("/api/v1/admin/services", app.handleAdminServices)
	mux.HandleFunc("/api/v1/admin/devices", app.handleAdminDevices)
	mux.HandleFunc("/api/v1/admin/audit-events", app.handleAdminAuditEvents)
	mux.HandleFunc("/api/v1/client/auth/login", app.handleClientLogin)
	mux.HandleFunc("/api/v1/client/auth/refresh", app.handleClientRefresh)
	mux.HandleFunc("/api/v1/client/auth/logout", app.handleLogout)
	mux.HandleFunc("/api/v1/client/me", app.handleCurrentUser)
	mux.HandleFunc("/api/v1/client/devices/register", app.handleDeviceRegister)
	mux.HandleFunc("/api/v1/client/devices/challenge", app.handleDeviceChallenge)
	mux.HandleFunc("/api/v1/client/devices/challenge/verify", app.handleDeviceChallengeVerify)
	mux.HandleFunc("/api/v1/client/services", app.handleClientServices)
	mux.HandleFunc("/api/v1/client/services/", app.handleClientServiceByID)
	if options.RegisterRoutes != nil {
		options.RegisterRoutes(mux, app)
	}

	var handler http.Handler = mux
	handler = app.recoveryMiddleware(handler)
	handler = app.accessLogMiddleware(handler)
	handler = app.requestIDMiddleware(handler)
	app.handler = handler

	return app
}

func (a *App) Handler() http.Handler {
	return a.handler
}

func (a *App) handleIndex(writer http.ResponseWriter, request *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]any{
		"name":      "bifrost-gateway",
		"readyTime": a.readyTime,
		"service":   "gateway",
	})
}

func (a *App) handleHealthz(writer http.ResponseWriter, request *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func (a *App) handleReadyz(writer http.ResponseWriter, request *http.Request) {
	if a.readyCheck != nil {
		if err := a.readyCheck(request.Context()); err != nil {
			writeJSON(writer, http.StatusServiceUnavailable, map[string]any{
				"error":     err.Error(),
				"readyTime": a.readyTime,
				"status":    "not-ready",
				"upstreams": a.upstreams,
			})
			return
		}
	}

	writeJSON(writer, http.StatusOK, map[string]any{
		"readyTime": a.readyTime,
		"status":    "ready",
		"upstreams": a.upstreams,
	})
}

func (a *App) handleUpstreamProbe(writer http.ResponseWriter, request *http.Request) {
	serviceKey := strings.TrimPrefix(request.URL.Path, "/debug/upstreams/")
	target, ok := a.upstreams[serviceKey]
	if !ok {
		writeJSON(writer, http.StatusNotFound, map[string]string{
			"error":      "upstream not configured",
			"serviceKey": serviceKey,
		})
		return
	}

	targetURL := strings.TrimSuffix(target, "/") + "/whoami"
	ctx, cancel := context.WithTimeout(request.Context(), 3*time.Second)
	defer cancel()

	upstreamRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		writeJSON(writer, http.StatusBadGateway, map[string]string{
			"error": err.Error(),
		})
		return
	}

	upstreamResponse, err := a.proxyHTTPClient().Do(upstreamRequest)
	if err != nil {
		writeJSON(writer, http.StatusBadGateway, map[string]string{
			"error":      err.Error(),
			"serviceKey": serviceKey,
			"target":     targetURL,
		})
		return
	}
	defer upstreamResponse.Body.Close()

	var upstreamBody map[string]any
	if err := json.NewDecoder(upstreamResponse.Body).Decode(&upstreamBody); err != nil {
		writeJSON(writer, http.StatusBadGateway, map[string]string{
			"error": err.Error(),
		})
		return
	}

	writeJSON(writer, http.StatusOK, map[string]any{
		"serviceKey": serviceKey,
		"target":     targetURL,
		"upstream":   upstreamBody,
	})
}

func (a *App) nowUTC() time.Time {
	if a.now != nil {
		return a.now().UTC()
	}
	return time.Now().UTC()
}

func (a *App) proxyHTTPClient() *http.Client {
	return a.httpClient
}

func defaultProxyHTTPClient() *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   5 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}

func (a *App) proxyRequestTimeout() time.Duration {
	if a.proxyTimeout > 0 {
		return a.proxyTimeout
	}
	return 5 * time.Second
}

func (a *App) proxyBodyLimit() int64 {
	if a.maxProxyBodyBytes > 0 {
		return a.maxProxyBodyBytes
	}
	return 1 << 20
}

func (a *App) newRequestID() string {
	if a.requestID != nil {
		return a.requestID()
	}

	random := make([]byte, 8)
	if _, err := rand.Read(random); err != nil {
		return fmt.Sprintf("req_%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("req_%x", random)
}

func IsServerClosed(err error) bool {
	return errors.Is(err, http.ErrServerClosed)
}
