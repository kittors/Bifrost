package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/kittors/bifrost/apps/gateway/internal/auth"
)

// 角色管理相关 handler 聚合在一起，避免角色路由与其他后台模块互相缠绕。
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
	if action == "" && request.Method == http.MethodPatch {
		a.handleAdminRoleUpdate(writer, request, roleID)
		return
	}
	if action == "services" && request.Method == http.MethodPut {
		a.handleAdminRoleServicesReplace(writer, request, roleID)
		return
	}
	writer.WriteHeader(http.StatusMethodNotAllowed)
}

func (a *App) handleAdminRoleUpdate(writer http.ResponseWriter, request *http.Request, roleID string) {
	requestID, timestamp := a.requestMeta(request)
	token, ok := bearerToken(request)
	if !ok {
		a.writeAPIError(writer, requestID, timestamp, missingBearerTokenError())
		return
	}

	var payload struct {
		DisplayName string `json:"displayName"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		a.writeAPIError(writer, requestID, timestamp, badJSONError())
		return
	}

	role, err := a.authService.UpdateAdminRole(request.Context(), auth.UpdateAdminRoleInput{
		AccessToken: token,
		RoleID:      roleID,
		DisplayName: payload.DisplayName,
		Description: payload.Description,
	})
	if err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}

	a.writeAPISuccess(writer, http.StatusOK, requestID, timestamp, adminRolePayload(role))
}

func (a *App) handleAdminRoleList(writer http.ResponseWriter, request *http.Request) {
	requestID, timestamp := a.requestMeta(request)
	token, ok := bearerToken(request)
	if !ok {
		a.writeAPIError(writer, requestID, timestamp, missingBearerTokenError())
		return
	}

	result, err := a.authService.ListAdminRoles(request.Context(), auth.ListAdminRolesInput{
		AccessToken: token,
		Page:        parseIntQuery(request, "page", 1),
		PageSize:    parseIntQuery(request, "pageSize", 20),
		Keyword:     strings.TrimSpace(request.URL.Query().Get("keyword")),
	})
	if err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}

	items := make([]map[string]any, 0, len(result.Items))
	for _, role := range result.Items {
		items = append(items, adminRolePayload(role))
	}

	a.writeAPISuccessWithPagination(writer, http.StatusOK, requestID, timestamp, map[string]any{
		"items": items,
	}, &result.Pagination)
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

	role, err := a.authService.CreateAdminRole(request.Context(), auth.CreateAdminRoleInput{
		AccessToken: token,
		Name:        payload.Name,
		DisplayName: payload.DisplayName,
		Description: payload.Description,
	})
	if err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}

	a.writeAPISuccess(writer, http.StatusCreated, requestID, timestamp, adminRolePayload(role))
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

	if err := a.authService.ReplaceRoleServices(request.Context(), auth.ReplaceRoleServicesInput{
		AccessToken: token,
		RoleID:      roleID,
		ServiceIDs:  payload.ServiceIDs,
	}); err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}

	a.writeAPISuccess(writer, http.StatusOK, requestID, timestamp, map[string]any{
		"roleId":     roleID,
		"serviceIds": payload.ServiceIDs,
	})
}
