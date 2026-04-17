package server

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/kittors/bifrost/apps/gateway/internal/auth"
	"github.com/kittors/bifrost/apps/gateway/internal/contracts"
)

type Options struct {
	ReadyCheck  func(ctx context.Context) error
	ReadyTime   string
	Upstreams   map[string]string
	Now         func() time.Time
	RequestID   func() string
	AuthService AuthService
}

type App struct {
	handler     http.Handler
	readyCheck  func(ctx context.Context) error
	readyTime   string
	upstreams   map[string]string
	now         func() time.Time
	requestID   func() string
	authService AuthService
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
}

func New(options Options) *App {
	app := &App{
		readyCheck:  options.ReadyCheck,
		readyTime:   options.ReadyTime,
		upstreams:   options.Upstreams,
		now:         options.Now,
		requestID:   options.RequestID,
		authService: options.AuthService,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", app.handleIndex)
	mux.HandleFunc("/healthz", app.handleHealthz)
	mux.HandleFunc("/readyz", app.handleReadyz)
	mux.HandleFunc("/debug/upstreams/", app.handleUpstreamProbe)
	mux.HandleFunc("/api/v1/admin/auth/login", app.handleAdminLogin)
	mux.HandleFunc("/api/v1/admin/auth/refresh", app.handleAdminRefresh)
	mux.HandleFunc("/api/v1/admin/auth/logout", app.handleLogout)
	mux.HandleFunc("/api/v1/admin/auth/me", app.handleCurrentUser)
	mux.HandleFunc("/api/v1/client/auth/login", app.handleClientLogin)
	mux.HandleFunc("/api/v1/client/auth/refresh", app.handleClientRefresh)
	mux.HandleFunc("/api/v1/client/auth/logout", app.handleLogout)
	mux.HandleFunc("/api/v1/client/me", app.handleCurrentUser)
	mux.HandleFunc("/api/v1/client/devices/register", app.handleDeviceRegister)
	mux.HandleFunc("/api/v1/client/devices/challenge", app.handleDeviceChallenge)
	mux.HandleFunc("/api/v1/client/devices/challenge/verify", app.handleDeviceChallengeVerify)
	mux.HandleFunc("/api/v1/client/services", app.handleClientServices)
	mux.HandleFunc("/api/v1/client/services/", app.handleClientServiceByID)
	app.handler = mux

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

	upstreamResponse, err := http.DefaultClient.Do(upstreamRequest)
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
		Username: payload.Username,
		Password: payload.Password,
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

func (a *App) requestMeta(request *http.Request) (string, string) {
	requestID := strings.TrimSpace(request.Header.Get("X-Request-Id"))
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

func absoluteURL(request *http.Request, publicPath string) string {
	scheme := strings.TrimSpace(request.Header.Get("X-Forwarded-Proto"))
	if scheme == "" {
		scheme = "http"
	}
	return scheme + "://" + request.Host + strings.TrimRight(publicPath, "/") + "/"
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

func (a *App) nowUTC() time.Time {
	if a.now != nil {
		return a.now().UTC()
	}
	return time.Now().UTC()
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
