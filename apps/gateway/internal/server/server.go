package server

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/kittors/bifrost/apps/gateway/internal/auth"
	"github.com/kittors/bifrost/apps/gateway/internal/contracts"
)

type contextKey string

const requestIDContextKey contextKey = "request_id"

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

func (a *App) handleServiceProxy(writer http.ResponseWriter, request *http.Request) {
	requestID, timestamp := a.requestMeta(request)
	serviceKey, upstreamPath, ok := parseProxyPath(request.URL.Path)
	if !ok {
		a.writeAPIError(writer, requestID, timestamp, apiError{
			statusCode:  http.StatusNotFound,
			code:        contracts.ErrorCodeGatewayRouteNotFound,
			message:     "proxy route not found",
			userMessage: "访问路径不存在",
		})
		return
	}

	if a.authService == nil {
		a.writeAPIError(writer, requestID, timestamp, apiError{
			statusCode:  http.StatusInternalServerError,
			code:        contracts.ErrorCodeCommonInternalError,
			message:     "auth service is not configured",
			userMessage: "服务暂时不可用，请稍后再试",
		})
		return
	}

	token, ok := bearerToken(request)
	if !ok {
		a.writeAPIError(writer, requestID, timestamp, missingBearerTokenError())
		return
	}

	target, err := a.authService.ResolveProxyRequest(request.Context(), auth.ResolveProxyRequestInput{
		AccessToken: token,
		RequestID:   requestID,
		ServiceKey:  serviceKey,
	})
	if err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}

	body, err := a.readProxyBody(writer, request)
	if err != nil {
		a.writeAPIError(writer, requestID, timestamp, apiError{
			statusCode:  http.StatusRequestEntityTooLarge,
			code:        contracts.ErrorCodeGatewayRequestTooLarge,
			message:     "request body exceeds proxy limit",
			userMessage: "请求体过大",
		})
		return
	}

	targetURL, err := buildUpstreamURL(target.UpstreamURL, upstreamPath, request.URL.RawQuery)
	if err != nil {
		a.writeAPIError(writer, requestID, timestamp, apiError{
			statusCode:  http.StatusBadGateway,
			code:        contracts.ErrorCodeServiceUpstreamInvalid,
			message:     err.Error(),
			userMessage: "服务暂时不可用，请稍后再试",
		})
		return
	}

	proxyCtx, cancel := context.WithTimeout(request.Context(), a.proxyRequestTimeout())
	defer cancel()

	upstreamRequest, err := http.NewRequestWithContext(proxyCtx, request.Method, targetURL, bytes.NewReader(body))
	if err != nil {
		a.writeAPIError(writer, requestID, timestamp, apiError{
			statusCode:  http.StatusBadGateway,
			code:        contracts.ErrorCodeGatewayBadUpstream,
			message:     err.Error(),
			userMessage: "服务暂时不可用，请稍后再试",
		})
		return
	}

	copyProxyHeaders(upstreamRequest.Header, request.Header)
	upstreamRequest.Header.Set("X-Bifrost-Request-Id", requestID)
	upstreamRequest.Header.Set("X-Bifrost-Service-Key", target.ServiceKey)
	upstreamRequest.Header.Set("X-Bifrost-User-Id", target.UserID)
	if target.AccessSource != "" {
		upstreamRequest.Header.Set("X-Bifrost-Access-Source", target.AccessSource)
	}

	upstreamResponse, err := a.proxyHTTPClient().Do(upstreamRequest)
	if err != nil {
		if auditErr := a.authService.RecordProxyAccessEvent(request.Context(), auth.RecordProxyAccessEventInput{
			RequestID: requestID,
			Type:      contracts.AuditEventTypeServiceAccessUpstreamError,
			UserID:    target.UserID,
			DeviceID:  target.DeviceID,
			ServiceID: target.ServiceID,
			Result:    "failure",
			Summary:   "upstream request failed",
		}); auditErr != nil {
			a.writeMappedError(writer, requestID, timestamp, auditErr)
			return
		}

		if isTimeoutError(err) {
			a.writeAPIError(writer, requestID, timestamp, apiError{
				statusCode:  http.StatusGatewayTimeout,
				code:        contracts.ErrorCodeGatewayUpstreamTimeout,
				message:     err.Error(),
				userMessage: "上游服务响应超时",
			})
			return
		}

		a.writeAPIError(writer, requestID, timestamp, apiError{
			statusCode:  http.StatusBadGateway,
			code:        contracts.ErrorCodeGatewayBadUpstream,
			message:     err.Error(),
			userMessage: "上游服务暂时不可用",
		})
		return
	}
	defer upstreamResponse.Body.Close()

	if auditErr := a.authService.RecordProxyAccessEvent(request.Context(), auth.RecordProxyAccessEventInput{
		RequestID: requestID,
		Type:      contracts.AuditEventTypeServiceAccessGranted,
		UserID:    target.UserID,
		DeviceID:  target.DeviceID,
		ServiceID: target.ServiceID,
		Result:    "success",
		Summary:   "service access granted",
	}); auditErr != nil {
		a.writeMappedError(writer, requestID, timestamp, auditErr)
		return
	}

	copyResponseHeaders(writer.Header(), upstreamResponse.Header)
	writer.WriteHeader(upstreamResponse.StatusCode)
	_, _ = io.Copy(writer, upstreamResponse.Body)
}

func writeJSON(writer http.ResponseWriter, statusCode int, payload any) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(statusCode)
	_ = json.NewEncoder(writer).Encode(payload)
}

func (a *App) handleAdminLogin(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	requestID, timestamp := a.requestMeta(request)
	if a.authService == nil {
		a.writeAPIError(writer, requestID, timestamp, apiError{
			statusCode:  http.StatusInternalServerError,
			code:        contracts.ErrorCodeCommonInternalError,
			message:     "auth service is not configured",
			userMessage: "服务暂时不可用，请稍后再试",
		})
		return
	}

	var payload struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		a.writeAPIError(writer, requestID, timestamp, apiError{
			statusCode:  http.StatusBadRequest,
			code:        contracts.ErrorCodeCommonBadRequest,
			message:     "request body must be valid JSON",
			userMessage: "请求参数不正确",
		})
		return
	}

	result, err := a.authService.AdminLogin(request.Context(), auth.AdminLoginInput{
		Username:  payload.Username,
		Password:  payload.Password,
		RequestID: requestID,
	})
	if err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}

	a.writeAPISuccess(writer, http.StatusOK, requestID, timestamp, map[string]any{
		"accessToken":  result.AccessToken,
		"refreshToken": result.RefreshToken,
		"expiresIn":    result.ExpiresIn,
		"user": map[string]any{
			"id":          result.User.ID,
			"username":    result.User.Username,
			"displayName": result.User.DisplayName,
			"roles":       result.User.Roles,
		},
	})
}

func (a *App) handleClientLogin(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	requestID, timestamp := a.requestMeta(request)
	if a.authService == nil {
		a.writeAPIError(writer, requestID, timestamp, apiError{
			statusCode:  http.StatusInternalServerError,
			code:        contracts.ErrorCodeCommonInternalError,
			message:     "auth service is not configured",
			userMessage: "服务暂时不可用，请稍后再试",
		})
		return
	}

	var payload struct {
		Username      string `json:"username"`
		Password      string `json:"password"`
		DeviceID      string `json:"deviceId"`
		ClientVersion string `json:"clientVersion"`
	}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		a.writeAPIError(writer, requestID, timestamp, apiError{
			statusCode:  http.StatusBadRequest,
			code:        contracts.ErrorCodeCommonBadRequest,
			message:     "request body must be valid JSON",
			userMessage: "请求参数不正确",
		})
		return
	}

	result, err := a.authService.ClientLogin(request.Context(), auth.ClientLoginInput{
		Username:      payload.Username,
		Password:      payload.Password,
		DeviceID:      payload.DeviceID,
		ClientVersion: payload.ClientVersion,
	})
	if err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}

	a.writeAPISuccess(writer, http.StatusOK, requestID, timestamp, map[string]any{
		"accessToken":  result.AccessToken,
		"refreshToken": result.RefreshToken,
		"expiresIn":    result.ExpiresIn,
		"user":         loginUserPayload(result.User),
	})
}

func (a *App) handleAdminRefresh(writer http.ResponseWriter, request *http.Request) {
	a.handleRefresh(writer, request, false)
}

func (a *App) handleClientRefresh(writer http.ResponseWriter, request *http.Request) {
	a.handleRefresh(writer, request, true)
}

func (a *App) handleRefresh(writer http.ResponseWriter, request *http.Request, requireDevice bool) {
	if request.Method != http.MethodPost {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	requestID, timestamp := a.requestMeta(request)
	if a.authService == nil {
		a.writeAPIError(writer, requestID, timestamp, apiError{
			statusCode:  http.StatusInternalServerError,
			code:        contracts.ErrorCodeCommonInternalError,
			message:     "auth service is not configured",
			userMessage: "服务暂时不可用，请稍后再试",
		})
		return
	}

	var payload struct {
		RefreshToken string `json:"refreshToken"`
		DeviceID     string `json:"deviceId"`
	}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		a.writeAPIError(writer, requestID, timestamp, apiError{
			statusCode:  http.StatusBadRequest,
			code:        contracts.ErrorCodeCommonBadRequest,
			message:     "request body must be valid JSON",
			userMessage: "请求参数不正确",
		})
		return
	}

	input := auth.RefreshInput{RefreshToken: payload.RefreshToken}
	if requireDevice {
		input.DeviceID = payload.DeviceID
	}

	result, err := a.authService.RefreshSession(request.Context(), input)
	if err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}

	a.writeAPISuccess(writer, http.StatusOK, requestID, timestamp, map[string]any{
		"accessToken":  result.AccessToken,
		"refreshToken": result.RefreshToken,
		"expiresIn":    result.ExpiresIn,
		"user":         loginUserPayload(result.User),
	})
}

func (a *App) handleLogout(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	requestID, timestamp := a.requestMeta(request)
	token, ok := bearerToken(request)
	if !ok {
		a.writeAPIError(writer, requestID, timestamp, apiError{
			statusCode:  http.StatusUnauthorized,
			code:        contracts.ErrorCodeAuthInvalidToken,
			message:     "bearer token is required",
			userMessage: "登录状态已失效，请重新登录",
		})
		return
	}

	if err := a.authService.Logout(request.Context(), auth.LogoutInput{AccessToken: token}); err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}

	a.writeAPISuccess(writer, http.StatusOK, requestID, timestamp, map[string]any{
		"revoked": true,
	})
}

func (a *App) handleCurrentUser(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	requestID, timestamp := a.requestMeta(request)
	token, ok := bearerToken(request)
	if !ok {
		a.writeAPIError(writer, requestID, timestamp, apiError{
			statusCode:  http.StatusUnauthorized,
			code:        contracts.ErrorCodeAuthInvalidToken,
			message:     "bearer token is required",
			userMessage: "登录状态已失效，请重新登录",
		})
		return
	}

	user, err := a.authService.CurrentUser(request.Context(), auth.CurrentUserInput{AccessToken: token})
	if err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}

	a.writeAPISuccess(writer, http.StatusOK, requestID, timestamp, map[string]any{
		"user": loginUserPayload(user),
	})
}

func (a *App) handleDeviceRegister(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	requestID, timestamp := a.requestMeta(request)
	token, ok := bearerToken(request)
	if !ok {
		a.writeAPIError(writer, requestID, timestamp, apiError{
			statusCode:  http.StatusUnauthorized,
			code:        contracts.ErrorCodeAuthInvalidToken,
			message:     "bearer token is required",
			userMessage: "登录状态已失效，请重新登录",
		})
		return
	}

	var payload struct {
		Name                 string `json:"name"`
		OS                   string `json:"os"`
		ClientVersion        string `json:"clientVersion"`
		PublicKey            string `json:"publicKey"`
		PublicKeyFingerprint string `json:"publicKeyFingerprint"`
	}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		a.writeAPIError(writer, requestID, timestamp, apiError{
			statusCode:  http.StatusBadRequest,
			code:        contracts.ErrorCodeCommonBadRequest,
			message:     "request body must be valid JSON",
			userMessage: "请求参数不正确",
		})
		return
	}

	device, err := a.authService.RegisterDevice(request.Context(), auth.RegisterDeviceInput{
		AccessToken:          token,
		Name:                 payload.Name,
		OS:                   payload.OS,
		ClientVersion:        payload.ClientVersion,
		PublicKey:            payload.PublicKey,
		PublicKeyFingerprint: payload.PublicKeyFingerprint,
	})
	if err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}

	a.writeAPISuccess(writer, http.StatusCreated, requestID, timestamp, map[string]any{
		"deviceId": device.ID,
		"status":   device.Status,
	})
}

func (a *App) handleDeviceChallenge(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	requestID, timestamp := a.requestMeta(request)
	token, ok := bearerToken(request)
	if !ok {
		a.writeAPIError(writer, requestID, timestamp, apiError{
			statusCode:  http.StatusUnauthorized,
			code:        contracts.ErrorCodeAuthInvalidToken,
			message:     "bearer token is required",
			userMessage: "登录状态已失效，请重新登录",
		})
		return
	}

	var payload struct {
		DeviceID string `json:"deviceId"`
	}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		a.writeAPIError(writer, requestID, timestamp, apiError{
			statusCode:  http.StatusBadRequest,
			code:        contracts.ErrorCodeCommonBadRequest,
			message:     "request body must be valid JSON",
			userMessage: "请求参数不正确",
		})
		return
	}

	challenge, err := a.authService.CreateDeviceChallenge(request.Context(), auth.CreateDeviceChallengeInput{
		AccessToken: token,
		DeviceID:    payload.DeviceID,
	})
	if err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}

	a.writeAPISuccess(writer, http.StatusOK, requestID, timestamp, map[string]any{
		"challengeId": challenge.ID,
		"challenge":   challenge.Challenge,
		"expiresIn":   challenge.ExpiresIn,
	})
}

func (a *App) handleDeviceChallengeVerify(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	requestID, timestamp := a.requestMeta(request)
	token, ok := bearerToken(request)
	if !ok {
		a.writeAPIError(writer, requestID, timestamp, apiError{
			statusCode:  http.StatusUnauthorized,
			code:        contracts.ErrorCodeAuthInvalidToken,
			message:     "bearer token is required",
			userMessage: "登录状态已失效，请重新登录",
		})
		return
	}

	var payload struct {
		ChallengeID string `json:"challengeId"`
		Signature   string `json:"signature"`
	}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		a.writeAPIError(writer, requestID, timestamp, apiError{
			statusCode:  http.StatusBadRequest,
			code:        contracts.ErrorCodeCommonBadRequest,
			message:     "request body must be valid JSON",
			userMessage: "请求参数不正确",
		})
		return
	}

	result, err := a.authService.VerifyDeviceChallenge(request.Context(), auth.VerifyDeviceChallengeInput{
		AccessToken: token,
		ChallengeID: payload.ChallengeID,
		Signature:   payload.Signature,
	})
	if err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}

	a.writeAPISuccess(writer, http.StatusOK, requestID, timestamp, map[string]any{
		"verified": result.Verified,
	})
}

func (a *App) handleClientServices(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	requestID, timestamp := a.requestMeta(request)
	token, ok := bearerToken(request)
	if !ok {
		a.writeAPIError(writer, requestID, timestamp, missingBearerTokenError())
		return
	}

	services, err := a.authService.ListClientServices(request.Context(), auth.ListClientServicesInput{
		AccessToken: token,
		Keyword:     strings.TrimSpace(request.URL.Query().Get("keyword")),
		Group:       strings.TrimSpace(request.URL.Query().Get("group")),
	})
	if err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}

	items := make([]map[string]any, 0, len(services))
	for _, service := range services {
		items = append(items, clientServicePayload(service))
	}

	a.writeAPISuccess(writer, http.StatusOK, requestID, timestamp, map[string]any{
		"items": items,
	})
}

func (a *App) handleAdminUsers(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		a.handleAdminUserList(writer, request)
	case http.MethodPost:
		a.handleAdminUserCreate(writer, request)
	default:
		writer.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (a *App) handleAdminUserByID(writer http.ResponseWriter, request *http.Request) {
	userID, action, ok := parseAdminUserPath(request.URL.Path)
	if !ok {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	if action == "service-overrides" && request.Method == http.MethodPut {
		a.handleAdminUserServiceOverridesReplace(writer, request, userID)
		return
	}

	if request.Method != http.MethodPatch {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	a.handleAdminUserUpdate(writer, request, userID)
}

func (a *App) handleAdminRoles(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		a.handleAdminRoleList(writer, request)
	case http.MethodPost:
		a.handleAdminRoleCreate(writer, request)
	default:
		writer.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (a *App) handleAdminRoleByID(writer http.ResponseWriter, request *http.Request) {
	roleID, action, ok := parseAdminRolePath(request.URL.Path)
	if !ok {
		writer.WriteHeader(http.StatusNotFound)
		return
	}
	if action == "services" && request.Method == http.MethodPut {
		a.handleAdminRoleServicesReplace(writer, request, roleID)
		return
	}
	writer.WriteHeader(http.StatusMethodNotAllowed)
}

func (a *App) handleAdminServices(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case http.MethodGet:
		a.handleAdminServiceList(writer, request)
	case http.MethodPost:
		a.handleAdminServiceCreate(writer, request)
	default:
		writer.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (a *App) handleAdminDevices(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	a.handleAdminDeviceList(writer, request)
}

func (a *App) handleAdminAuditEvents(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	a.handleAdminAuditEventList(writer, request)
}

func (a *App) handleClientServiceByID(writer http.ResponseWriter, request *http.Request) {
	serviceID, action, ok := parseClientServicePath(request.URL.Path)
	if !ok {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	if action == "access-url" {
		a.handleClientServiceAccessURL(writer, request, serviceID)
		return
	}

	if request.Method != http.MethodGet {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	requestID, timestamp := a.requestMeta(request)
	token, ok := bearerToken(request)
	if !ok {
		a.writeAPIError(writer, requestID, timestamp, missingBearerTokenError())
		return
	}

	service, err := a.authService.GetClientService(request.Context(), auth.GetClientServiceInput{
		AccessToken: token,
		ServiceID:   serviceID,
	})
	if err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}

	a.writeAPISuccess(writer, http.StatusOK, requestID, timestamp, clientServicePayload(service))
}

func (a *App) handleClientServiceAccessURL(writer http.ResponseWriter, request *http.Request, serviceID string) {
	if request.Method != http.MethodPost {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	requestID, timestamp := a.requestMeta(request)
	token, ok := bearerToken(request)
	if !ok {
		a.writeAPIError(writer, requestID, timestamp, missingBearerTokenError())
		return
	}

	result, err := a.authService.CreateServiceAccessURL(request.Context(), auth.CreateServiceAccessURLInput{
		AccessToken: token,
		ServiceID:   serviceID,
	})
	if err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}

	a.writeAPISuccess(writer, http.StatusOK, requestID, timestamp, map[string]any{
		"url":       absoluteURL(request, result.PublicPath),
		"expiresIn": result.ExpiresIn,
	})
}

func (a *App) handleAdminUserList(writer http.ResponseWriter, request *http.Request) {
	requestID, timestamp := a.requestMeta(request)
	token, ok := bearerToken(request)
	if !ok {
		a.writeAPIError(writer, requestID, timestamp, missingBearerTokenError())
		return
	}

	result, err := a.authService.ListAdminUsers(request.Context(), auth.ListAdminUsersInput{
		AccessToken: token,
		Page:        parseIntQuery(request, "page", 1),
		PageSize:    parseIntQuery(request, "pageSize", 20),
		Keyword:     strings.TrimSpace(request.URL.Query().Get("keyword")),
		Status:      strings.TrimSpace(request.URL.Query().Get("status")),
		RoleID:      strings.TrimSpace(request.URL.Query().Get("roleId")),
	})
	if err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}

	items := make([]map[string]any, 0, len(result.Items))
	for _, user := range result.Items {
		items = append(items, adminUserPayload(user))
	}

	a.writeAPISuccessWithPagination(writer, http.StatusOK, requestID, timestamp, map[string]any{
		"items": items,
	}, &result.Pagination)
}

func (a *App) handleAdminUserCreate(writer http.ResponseWriter, request *http.Request) {
	requestID, timestamp := a.requestMeta(request)
	token, ok := bearerToken(request)
	if !ok {
		a.writeAPIError(writer, requestID, timestamp, missingBearerTokenError())
		return
	}

	var payload struct {
		Username    string   `json:"username"`
		DisplayName string   `json:"displayName"`
		Email       string   `json:"email"`
		Password    string   `json:"password"`
		RoleIDs     []string `json:"roleIds"`
	}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		a.writeAPIError(writer, requestID, timestamp, apiError{
			statusCode:  http.StatusBadRequest,
			code:        contracts.ErrorCodeCommonBadRequest,
			message:     "request body must be valid JSON",
			userMessage: "请求参数不正确",
		})
		return
	}

	user, err := a.authService.CreateAdminUser(request.Context(), auth.CreateAdminUserInput{
		AccessToken: token,
		RequestID:   requestID,
		Username:    payload.Username,
		DisplayName: payload.DisplayName,
		Email:       payload.Email,
		Password:    payload.Password,
		RoleIDs:     payload.RoleIDs,
	})
	if err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}

	a.writeAPISuccess(writer, http.StatusCreated, requestID, timestamp, adminUserPayload(user))
}

func (a *App) handleAdminUserUpdate(writer http.ResponseWriter, request *http.Request, userID string) {
	requestID, timestamp := a.requestMeta(request)
	token, ok := bearerToken(request)
	if !ok {
		a.writeAPIError(writer, requestID, timestamp, missingBearerTokenError())
		return
	}

	var payload struct {
		DisplayName string   `json:"displayName"`
		Email       string   `json:"email"`
		RoleIDs     []string `json:"roleIds"`
	}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		a.writeAPIError(writer, requestID, timestamp, apiError{
			statusCode:  http.StatusBadRequest,
			code:        contracts.ErrorCodeCommonBadRequest,
			message:     "request body must be valid JSON",
			userMessage: "请求参数不正确",
		})
		return
	}

	user, err := a.authService.UpdateAdminUser(request.Context(), auth.UpdateAdminUserInput{
		AccessToken: token,
		UserID:      userID,
		DisplayName: payload.DisplayName,
		Email:       payload.Email,
		RoleIDs:     payload.RoleIDs,
	})
	if err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}

	a.writeAPISuccess(writer, http.StatusOK, requestID, timestamp, adminUserPayload(user))
}

func (a *App) handleAdminRoleList(writer http.ResponseWriter, request *http.Request) {
	requestID, timestamp := a.requestMeta(request)
	token, ok := bearerToken(request)
	if !ok {
		a.writeAPIError(writer, requestID, timestamp, missingBearerTokenError())
		return
	}
	result, err := a.authService.ListAdminRoles(request.Context(), auth.ListAdminRolesInput{AccessToken: token, Page: parseIntQuery(request, "page", 1), PageSize: parseIntQuery(request, "pageSize", 20), Keyword: strings.TrimSpace(request.URL.Query().Get("keyword"))})
	if err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}
	items := make([]map[string]any, 0, len(result.Items))
	for _, role := range result.Items {
		items = append(items, adminRolePayload(role))
	}
	a.writeAPISuccessWithPagination(writer, http.StatusOK, requestID, timestamp, map[string]any{"items": items}, &result.Pagination)
}

func (a *App) handleAdminRoleCreate(writer http.ResponseWriter, request *http.Request) {
	requestID, timestamp := a.requestMeta(request)
	token, ok := bearerToken(request)
	if !ok {
		a.writeAPIError(writer, requestID, timestamp, missingBearerTokenError())
		return
	}
	var payload struct {
		Name        string `json:"name"`
		DisplayName string `json:"displayName"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		a.writeAPIError(writer, requestID, timestamp, badJSONError())
		return
	}
	role, err := a.authService.CreateAdminRole(request.Context(), auth.CreateAdminRoleInput{AccessToken: token, Name: payload.Name, DisplayName: payload.DisplayName, Description: payload.Description})
	if err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}
	a.writeAPISuccess(writer, http.StatusCreated, requestID, timestamp, adminRolePayload(role))
}

func (a *App) handleAdminServiceList(writer http.ResponseWriter, request *http.Request) {
	requestID, timestamp := a.requestMeta(request)
	token, ok := bearerToken(request)
	if !ok {
		a.writeAPIError(writer, requestID, timestamp, missingBearerTokenError())
		return
	}
	result, err := a.authService.ListAdminServices(request.Context(), auth.ListAdminServicesInput{AccessToken: token, Page: parseIntQuery(request, "page", 1), PageSize: parseIntQuery(request, "pageSize", 20), Keyword: strings.TrimSpace(request.URL.Query().Get("keyword")), Status: strings.TrimSpace(request.URL.Query().Get("status")), Group: strings.TrimSpace(request.URL.Query().Get("group"))})
	if err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}
	items := make([]map[string]any, 0, len(result.Items))
	for _, service := range result.Items {
		items = append(items, adminServicePayload(service))
	}
	a.writeAPISuccessWithPagination(writer, http.StatusOK, requestID, timestamp, map[string]any{"items": items}, &result.Pagination)
}

func (a *App) handleAdminServiceCreate(writer http.ResponseWriter, request *http.Request) {
	requestID, timestamp := a.requestMeta(request)
	token, ok := bearerToken(request)
	if !ok {
		a.writeAPIError(writer, requestID, timestamp, missingBearerTokenError())
		return
	}
	var payload struct {
		Key         string `json:"key"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Group       string `json:"group"`
		Protocol    string `json:"protocol"`
		UpstreamURL string `json:"upstreamUrl"`
		PublicPath  string `json:"publicPath"`
		Enabled     bool   `json:"enabled"`
	}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		a.writeAPIError(writer, requestID, timestamp, badJSONError())
		return
	}
	service, err := a.authService.CreateAdminService(request.Context(), auth.CreateAdminServiceInput{AccessToken: token, Key: payload.Key, Name: payload.Name, Description: payload.Description, Group: payload.Group, Protocol: payload.Protocol, UpstreamURL: payload.UpstreamURL, PublicPath: payload.PublicPath, Enabled: payload.Enabled})
	if err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}
	a.writeAPISuccess(writer, http.StatusCreated, requestID, timestamp, adminServicePayload(service))
}

func (a *App) handleAdminDeviceList(writer http.ResponseWriter, request *http.Request) {
	requestID, timestamp := a.requestMeta(request)
	token, ok := bearerToken(request)
	if !ok {
		a.writeAPIError(writer, requestID, timestamp, missingBearerTokenError())
		return
	}
	result, err := a.authService.ListAdminDevices(request.Context(), auth.ListAdminDevicesInput{AccessToken: token, Page: parseIntQuery(request, "page", 1), PageSize: parseIntQuery(request, "pageSize", 20), Keyword: strings.TrimSpace(request.URL.Query().Get("keyword")), Status: strings.TrimSpace(request.URL.Query().Get("status")), UserID: strings.TrimSpace(request.URL.Query().Get("userId"))})
	if err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}
	items := make([]map[string]any, 0, len(result.Items))
	for _, device := range result.Items {
		items = append(items, adminDevicePayload(device))
	}
	a.writeAPISuccessWithPagination(writer, http.StatusOK, requestID, timestamp, map[string]any{"items": items}, &result.Pagination)
}

func (a *App) handleAdminAuditEventList(writer http.ResponseWriter, request *http.Request) {
	requestID, timestamp := a.requestMeta(request)
	token, ok := bearerToken(request)
	if !ok {
		a.writeAPIError(writer, requestID, timestamp, missingBearerTokenError())
		return
	}
	result, err := a.authService.ListAdminAuditEvents(request.Context(), auth.ListAdminAuditEventsInput{AccessToken: token, Page: parseIntQuery(request, "page", 1), PageSize: parseIntQuery(request, "pageSize", 20), Type: strings.TrimSpace(request.URL.Query().Get("type")), ActorUserID: strings.TrimSpace(request.URL.Query().Get("actorUserId")), TargetType: strings.TrimSpace(request.URL.Query().Get("targetType")), TargetID: strings.TrimSpace(request.URL.Query().Get("targetId")), ServiceID: strings.TrimSpace(request.URL.Query().Get("serviceId")), Result: strings.TrimSpace(request.URL.Query().Get("result"))})
	if err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}
	items := make([]map[string]any, 0, len(result.Items))
	for _, event := range result.Items {
		items = append(items, adminAuditEventPayload(event))
	}
	a.writeAPISuccessWithPagination(writer, http.StatusOK, requestID, timestamp, map[string]any{"items": items}, &result.Pagination)
}

func (a *App) handleAdminRoleServicesReplace(writer http.ResponseWriter, request *http.Request, roleID string) {
	requestID, timestamp := a.requestMeta(request)
	token, ok := bearerToken(request)
	if !ok {
		a.writeAPIError(writer, requestID, timestamp, missingBearerTokenError())
		return
	}
	var payload struct {
		ServiceIDs []string `json:"serviceIds"`
	}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		a.writeAPIError(writer, requestID, timestamp, badJSONError())
		return
	}
	if err := a.authService.ReplaceRoleServices(request.Context(), auth.ReplaceRoleServicesInput{AccessToken: token, RoleID: roleID, ServiceIDs: payload.ServiceIDs}); err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}
	a.writeAPISuccess(writer, http.StatusOK, requestID, timestamp, map[string]any{"roleId": roleID, "serviceIds": payload.ServiceIDs})
}

func (a *App) handleAdminUserServiceOverridesReplace(writer http.ResponseWriter, request *http.Request, userID string) {
	requestID, timestamp := a.requestMeta(request)
	token, ok := bearerToken(request)
	if !ok {
		a.writeAPIError(writer, requestID, timestamp, missingBearerTokenError())
		return
	}
	var payload struct {
		AllowServiceIDs []string `json:"allowServiceIds"`
		DenyServiceIDs  []string `json:"denyServiceIds"`
	}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		a.writeAPIError(writer, requestID, timestamp, badJSONError())
		return
	}
	overrides, err := a.authService.ReplaceUserServiceOverrides(request.Context(), auth.ReplaceUserServiceOverridesInput{AccessToken: token, UserID: userID, AllowServiceIDs: payload.AllowServiceIDs, DenyServiceIDs: payload.DenyServiceIDs})
	if err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}
	items := make([]map[string]any, 0, len(overrides))
	for _, override := range overrides {
		items = append(items, map[string]any{"serviceId": override.ServiceID, "effect": override.Effect})
	}
	a.writeAPISuccess(writer, http.StatusOK, requestID, timestamp, map[string]any{"items": items})
}

func (a *App) requestMeta(request *http.Request) (string, string) {
	requestID := requestIDFromContext(request.Context())
	if requestID == "" {
		requestID = strings.TrimSpace(request.Header.Get("X-Request-Id"))
	}
	if requestID == "" {
		requestID = a.newRequestID()
	}

	return requestID, a.nowUTC().Format(time.RFC3339)
}

func (a *App) writeAPISuccess(writer http.ResponseWriter, statusCode int, requestID string, timestamp string, data any) {
	writer.Header().Set("X-Request-Id", requestID)
	writeJSON(writer, statusCode, map[string]any{
		"success": true,
		"data":    data,
		"meta": map[string]any{
			"requestId": requestID,
			"timestamp": timestamp,
		},
		"error": nil,
	})
}

func (a *App) writeAPISuccessWithPagination(writer http.ResponseWriter, statusCode int, requestID string, timestamp string, data any, pagination *contracts.Pagination) {
	writer.Header().Set("X-Request-Id", requestID)
	meta := map[string]any{
		"requestId": requestID,
		"timestamp": timestamp,
	}
	if pagination != nil {
		meta["pagination"] = pagination
	}
	writeJSON(writer, statusCode, map[string]any{
		"success": true,
		"data":    data,
		"meta":    meta,
		"error":   nil,
	})
}

func bearerToken(request *http.Request) (string, bool) {
	header := strings.TrimSpace(request.Header.Get("Authorization"))
	value, ok := strings.CutPrefix(header, "Bearer ")
	if !ok || strings.TrimSpace(value) == "" {
		return "", false
	}
	return strings.TrimSpace(value), true
}

func missingBearerTokenError() apiError {
	return apiError{
		statusCode:  http.StatusUnauthorized,
		code:        contracts.ErrorCodeAuthInvalidToken,
		message:     "bearer token is required",
		userMessage: "登录状态已失效，请重新登录",
	}
}

func badJSONError() apiError {
	return apiError{
		statusCode:  http.StatusBadRequest,
		code:        contracts.ErrorCodeCommonBadRequest,
		message:     "request body must be valid JSON",
		userMessage: "请求参数不正确",
	}
}

func loginUserPayload(user auth.LoginUser) map[string]any {
	return map[string]any{
		"id":          user.ID,
		"username":    user.Username,
		"displayName": user.DisplayName,
		"roles":       user.Roles,
	}
}

func clientServicePayload(service auth.ClientService) map[string]any {
	return map[string]any{
		"id":           service.ID,
		"key":          service.Key,
		"name":         service.Name,
		"description":  service.Description,
		"group":        service.Group,
		"status":       service.Status,
		"accessSource": service.AccessSource,
	}
}

func adminUserPayload(user auth.AdminUser) map[string]any {
	return map[string]any{
		"id":          user.ID,
		"username":    user.Username,
		"displayName": user.DisplayName,
		"email":       user.Email,
		"status":      user.Status,
		"roles":       user.Roles,
	}
}

func adminRolePayload(role auth.AdminRole) map[string]any {
	return map[string]any{
		"id":          role.ID,
		"name":        role.Name,
		"displayName": role.DisplayName,
		"description": role.Description,
	}
}

func adminServicePayload(service auth.AdminService) map[string]any {
	return map[string]any{
		"id":          service.ID,
		"key":         service.Key,
		"name":        service.Name,
		"description": service.Description,
		"group":       service.Group,
		"protocol":    service.Protocol,
		"upstreamUrl": service.UpstreamURL,
		"publicPath":  service.PublicPath,
		"status":      service.Status,
	}
}

func adminDevicePayload(device auth.AdminDevice) map[string]any {
	return map[string]any{
		"id":                   device.ID,
		"userId":               device.UserID,
		"userUsername":         device.UserUsername,
		"name":                 device.Name,
		"os":                   device.OS,
		"clientVersion":        device.ClientVersion,
		"publicKeyFingerprint": device.PublicKeyFingerprint,
		"status":               device.Status,
	}
}

func adminAuditEventPayload(event auth.AdminAuditEvent) map[string]any {
	return map[string]any{
		"id":          event.ID,
		"requestId":   event.RequestID,
		"type":        event.Type,
		"actorUserId": event.ActorUserID,
		"targetType":  event.TargetType,
		"targetId":    event.TargetID,
		"serviceId":   event.ServiceID,
		"result":      event.Result,
		"summary":     event.Summary,
	}
}

func parseClientServicePath(path string) (string, string, bool) {
	remaining := strings.TrimPrefix(path, "/api/v1/client/services/")
	if remaining == path || remaining == "" {
		return "", "", false
	}

	parts := strings.Split(strings.Trim(remaining, "/"), "/")
	if len(parts) == 1 {
		return parts[0], "", parts[0] != ""
	}
	if len(parts) == 2 && parts[1] == "access-url" {
		return parts[0], parts[1], parts[0] != ""
	}
	return "", "", false
}

func parseAdminUserPath(path string) (string, string, bool) {
	remaining := strings.TrimPrefix(path, "/api/v1/admin/users/")
	if remaining == path || remaining == "" {
		return "", "", false
	}
	parts := strings.Split(strings.Trim(remaining, "/"), "/")
	if len(parts) == 1 {
		return parts[0], "", parts[0] != ""
	}
	if len(parts) == 2 && parts[1] == "service-overrides" {
		return parts[0], parts[1], parts[0] != ""
	}
	return "", "", false
}

func parseAdminRolePath(path string) (string, string, bool) {
	remaining := strings.TrimPrefix(path, "/api/v1/admin/roles/")
	if remaining == path || remaining == "" {
		return "", "", false
	}
	parts := strings.Split(strings.Trim(remaining, "/"), "/")
	if len(parts) == 2 && parts[1] == "services" {
		return parts[0], parts[1], parts[0] != ""
	}
	return "", "", false
}

func parseProxyPath(path string) (string, string, bool) {
	remaining := strings.TrimPrefix(path, "/s/")
	if remaining == path || remaining == "" {
		return "", "", false
	}

	parts := strings.SplitN(strings.TrimPrefix(remaining, "/"), "/", 2)
	serviceKey := strings.TrimSpace(parts[0])
	if serviceKey == "" {
		return "", "", false
	}

	if len(parts) == 1 {
		return serviceKey, "/", true
	}

	return serviceKey, "/" + parts[1], true
}

func absoluteURL(request *http.Request, publicPath string) string {
	scheme := strings.TrimSpace(request.Header.Get("X-Forwarded-Proto"))
	if scheme == "" {
		scheme = "http"
	}
	return scheme + "://" + request.Host + strings.TrimRight(publicPath, "/") + "/"
}

func parseIntQuery(request *http.Request, key string, fallback int) int {
	value := strings.TrimSpace(request.URL.Query().Get(key))
	if value == "" {
		return fallback
	}
	var parsed int
	if _, err := fmt.Sscanf(value, "%d", &parsed); err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}

func (a *App) writeMappedError(writer http.ResponseWriter, requestID string, timestamp string, err error) {
	var serviceErr *auth.ServiceError
	if errors.As(err, &serviceErr) {
		a.writeAPIError(writer, requestID, timestamp, apiError{
			statusCode:  serviceErr.StatusCode,
			code:        serviceErr.Code,
			message:     serviceErr.Message,
			userMessage: serviceErr.UserMessage,
		})
		return
	}

	a.writeAPIError(writer, requestID, timestamp, apiError{
		statusCode:  http.StatusInternalServerError,
		code:        contracts.ErrorCodeCommonInternalError,
		message:     err.Error(),
		userMessage: "服务暂时不可用，请稍后再试",
	})
}

func (a *App) writeAPIError(writer http.ResponseWriter, requestID string, timestamp string, err apiError) {
	writer.Header().Set("X-Request-Id", requestID)
	writeJSON(writer, err.statusCode, map[string]any{
		"success": false,
		"data":    nil,
		"meta": map[string]any{
			"requestId": requestID,
			"timestamp": timestamp,
		},
		"error": map[string]any{
			"code":        err.code,
			"message":     err.message,
			"userMessage": err.userMessage,
			"details":     map[string]any{},
		},
	})
}

type apiError struct {
	statusCode  int
	code        contracts.ErrorCode
	message     string
	userMessage string
}

func requestIDFromContext(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDContextKey).(string)
	return strings.TrimSpace(requestID)
}

func (a *App) requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requestID := strings.TrimSpace(request.Header.Get("X-Request-Id"))
		if requestID == "" {
			requestID = a.newRequestID()
		}

		writer.Header().Set("X-Request-Id", requestID)
		request = request.WithContext(context.WithValue(request.Context(), requestIDContextKey, requestID))
		next.ServeHTTP(writer, request)
	})
}

func (a *App) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer func() {
			recovered := recover()
			if recovered == nil {
				return
			}

			requestID, timestamp := a.requestMeta(request)
			if a.logger != nil {
				a.logger.Error("request panic recovered",
					"request_id", requestID,
					"path", request.URL.Path,
					"method", request.Method,
					"panic", recovered,
				)
			}

			a.writeAPIError(writer, requestID, timestamp, apiError{
				statusCode:  http.StatusInternalServerError,
				code:        contracts.ErrorCodeCommonInternalError,
				message:     fmt.Sprintf("panic: %v", recovered),
				userMessage: "服务暂时不可用，请稍后再试",
			})
		}()

		next.ServeHTTP(writer, request)
	})
}

func (a *App) accessLogMiddleware(next http.Handler) http.Handler {
	if a.logger == nil {
		return next
	}

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		recorder := &statusRecorder{ResponseWriter: writer, statusCode: http.StatusOK}
		startedAt := time.Now()
		next.ServeHTTP(recorder, request)

		a.logger.Info("http request completed",
			"request_id", requestIDFromContext(request.Context()),
			"method", request.Method,
			"path", request.URL.Path,
			"status", recorder.statusCode,
			"duration_ms", time.Since(startedAt).Milliseconds(),
		)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
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

func (a *App) readProxyBody(writer http.ResponseWriter, request *http.Request) ([]byte, error) {
	if request.Body == nil {
		return nil, nil
	}

	limitedBody := http.MaxBytesReader(writer, request.Body, a.proxyBodyLimit())
	defer limitedBody.Close()
	return io.ReadAll(limitedBody)
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

func buildUpstreamURL(baseURL string, upstreamPath string, rawQuery string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return "", fmt.Errorf("parse upstream url: %w", err)
	}

	parsed.Path = strings.TrimRight(parsed.Path, "/") + upstreamPath
	parsed.RawQuery = rawQuery
	return parsed.String(), nil
}

func copyProxyHeaders(target http.Header, source http.Header) {
	for key, values := range source {
		lowerKey := strings.ToLower(key)
		switch lowerKey {
		case "authorization", "host", "connection", "proxy-connection", "keep-alive", "te", "trailer", "transfer-encoding", "upgrade":
			continue
		}

		for _, value := range values {
			target.Add(key, value)
		}
	}
}

func copyResponseHeaders(target http.Header, source http.Header) {
	for key, values := range source {
		if strings.EqualFold(key, "X-Request-Id") {
			continue
		}
		for _, value := range values {
			target.Add(key, value)
		}
	}
}

func isTimeoutError(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}
