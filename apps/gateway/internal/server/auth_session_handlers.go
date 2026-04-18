package server

import (
	"encoding/json"
	"net/http"

	"github.com/kittors/bifrost/apps/gateway/internal/auth"
	"github.com/kittors/bifrost/apps/gateway/internal/contracts"
)

// 认证会话 handler 只处理登录、刷新、退出和当前用户，保持会话入口聚合。

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
		RequestID:     requestID,
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
