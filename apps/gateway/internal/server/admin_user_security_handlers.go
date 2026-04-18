package server

import (
	"encoding/json"
	"net/http"

	"github.com/kittors/bifrost/apps/gateway/internal/auth"
)

// 后台用户安全 handler 只负责密码重置与账号启停。
func (a *App) handleAdminUserPasswordReset(writer http.ResponseWriter, request *http.Request, userID string) {
	requestID, timestamp := a.requestMeta(request)
	token, ok := bearerToken(request)
	if !ok {
		a.writeAPIError(writer, requestID, timestamp, missingBearerTokenError())
		return
	}

	var payload struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		a.writeAPIError(writer, requestID, timestamp, badJSONError())
		return
	}

	if err := a.authService.ResetAdminUserPassword(request.Context(), auth.ResetAdminUserPasswordInput{
		AccessToken: token,
		RequestID:   requestID,
		UserID:      userID,
		Password:    payload.Password,
	}); err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}

	a.writeAPISuccess(writer, http.StatusOK, requestID, timestamp, map[string]any{
		"reset": true,
	})
}

func (a *App) handleAdminUserStatusSet(writer http.ResponseWriter, request *http.Request, userID string) {
	requestID, timestamp := a.requestMeta(request)
	token, ok := bearerToken(request)
	if !ok {
		a.writeAPIError(writer, requestID, timestamp, missingBearerTokenError())
		return
	}

	var payload struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		a.writeAPIError(writer, requestID, timestamp, badJSONError())
		return
	}

	user, err := a.authService.SetAdminUserStatus(request.Context(), auth.SetAdminUserStatusInput{
		AccessToken: token,
		RequestID:   requestID,
		UserID:      userID,
		Status:      payload.Status,
	})
	if err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}

	a.writeAPISuccess(writer, http.StatusOK, requestID, timestamp, adminUserPayload(user))
}
