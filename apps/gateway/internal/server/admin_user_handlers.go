package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/kittors/bifrost/apps/gateway/internal/auth"
)

// 后台用户 handler 负责用户列表、创建、更新和用户级服务覆盖。

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
