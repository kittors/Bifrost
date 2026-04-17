package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/kittors/bifrost/apps/gateway/internal/auth"
	"github.com/kittors/bifrost/apps/gateway/internal/contracts"
)

// 客户端与登录相关接口统一归档，方便后续继续拆出更细的鉴权流程。
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
		a.writeAPIError(writer, requestID, timestamp, badJSONError())
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
		"user":         loginUserPayload(result.User),
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
		a.writeAPIError(writer, requestID, timestamp, badJSONError())
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
		a.writeAPIError(writer, requestID, timestamp, badJSONError())
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
		a.writeAPIError(writer, requestID, timestamp, missingBearerTokenError())
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
		a.writeAPIError(writer, requestID, timestamp, missingBearerTokenError())
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
		a.writeAPIError(writer, requestID, timestamp, missingBearerTokenError())
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
		a.writeAPIError(writer, requestID, timestamp, badJSONError())
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
		a.writeAPIError(writer, requestID, timestamp, missingBearerTokenError())
		return
	}

	var payload struct {
		DeviceID string `json:"deviceId"`
	}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		a.writeAPIError(writer, requestID, timestamp, badJSONError())
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
		a.writeAPIError(writer, requestID, timestamp, missingBearerTokenError())
		return
	}

	var payload struct {
		ChallengeID string `json:"challengeId"`
		Signature   string `json:"signature"`
	}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		a.writeAPIError(writer, requestID, timestamp, badJSONError())
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

	http.SetCookie(writer, &http.Cookie{
		Name:     serviceAccessCookieName,
		Value:    result.AccessTicket,
		Path:     result.PublicPath,
		MaxAge:   result.ExpiresIn,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   requestIsSecure(request),
	})

	a.writeAPISuccess(writer, http.StatusOK, requestID, timestamp, map[string]any{
		"url":       absoluteURL(request, result.PublicPath),
		"expiresIn": result.ExpiresIn,
	})
}
