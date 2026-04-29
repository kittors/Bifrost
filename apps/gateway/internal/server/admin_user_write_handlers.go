package server

import (
	"encoding/json"
	"net/http"

	"github.com/kittors/bifrost/apps/gateway/internal/auth"
)

// 后台用户写入 handler 只处理创建、编辑与服务覆盖提交。
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
		a.writeAPIError(writer, requestID, timestamp, badJSONError())
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
		a.writeAPIError(writer, requestID, timestamp, badJSONError())
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

	overrides, err := a.authService.ReplaceUserServiceOverrides(request.Context(), auth.ReplaceUserServiceOverridesInput{
		AccessToken:     token,
		UserID:          userID,
		AllowServiceIDs: payload.AllowServiceIDs,
		DenyServiceIDs:  payload.DenyServiceIDs,
	})
	if err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}

	items := make([]map[string]any, 0, len(overrides))
	for _, override := range overrides {
		items = append(items, map[string]any{
			"serviceId": override.ServiceID,
			"effect":    override.Effect,
		})
	}

	a.writeAPISuccess(writer, http.StatusOK, requestID, timestamp, map[string]any{
		"items": items,
	})
}
