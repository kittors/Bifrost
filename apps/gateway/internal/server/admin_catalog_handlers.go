package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/kittors/bifrost/apps/gateway/internal/auth"
)

// 后台目录 handler 聚合角色、服务、设备、审计及角色服务授权接口。

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

func (a *App) handleAdminServiceByID(writer http.ResponseWriter, request *http.Request) {
	serviceID, action, ok := parseAdminServicePath(request.URL.Path)
	if !ok {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	if action == "" && request.Method == http.MethodGet {
		a.handleAdminServiceDetail(writer, request, serviceID)
		return
	}
	if action == "" && request.Method == http.MethodPatch {
		a.handleAdminServiceUpdate(writer, request, serviceID)
		return
	}
	if action == "status" && request.Method == http.MethodPost {
		a.handleAdminServiceStatusSet(writer, request, serviceID)
		return
	}

	writer.WriteHeader(http.StatusMethodNotAllowed)
}

func (a *App) handleAdminDevices(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	a.handleAdminDeviceList(writer, request)
}

func (a *App) handleAdminDeviceByID(writer http.ResponseWriter, request *http.Request) {
	deviceID, action, ok := parseAdminDevicePath(request.URL.Path)
	if !ok {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	if action == "" && request.Method == http.MethodGet {
		a.handleAdminDeviceDetail(writer, request, deviceID)
		return
	}
	if action == "status" && request.Method == http.MethodPost {
		a.handleAdminDeviceStatusSet(writer, request, deviceID)
		return
	}

	writer.WriteHeader(http.StatusMethodNotAllowed)
}

func (a *App) handleAdminAuditEvents(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	a.handleAdminAuditEventList(writer, request)
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

func (a *App) handleAdminServiceList(writer http.ResponseWriter, request *http.Request) {
	requestID, timestamp := a.requestMeta(request)
	token, ok := bearerToken(request)
	if !ok {
		a.writeAPIError(writer, requestID, timestamp, missingBearerTokenError())
		return
	}

	result, err := a.authService.ListAdminServices(request.Context(), auth.ListAdminServicesInput{
		AccessToken: token,
		Page:        parseIntQuery(request, "page", 1),
		PageSize:    parseIntQuery(request, "pageSize", 20),
		Keyword:     strings.TrimSpace(request.URL.Query().Get("keyword")),
		Status:      strings.TrimSpace(request.URL.Query().Get("status")),
		Group:       strings.TrimSpace(request.URL.Query().Get("group")),
	})
	if err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}

	items := make([]map[string]any, 0, len(result.Items))
	for _, service := range result.Items {
		items = append(items, adminServicePayload(service))
	}

	a.writeAPISuccessWithPagination(writer, http.StatusOK, requestID, timestamp, map[string]any{
		"items": items,
	}, &result.Pagination)
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

	service, err := a.authService.CreateAdminService(request.Context(), auth.CreateAdminServiceInput{
		AccessToken: token,
		Key:         payload.Key,
		Name:        payload.Name,
		Description: payload.Description,
		Group:       payload.Group,
		Protocol:    payload.Protocol,
		UpstreamURL: payload.UpstreamURL,
		PublicPath:  payload.PublicPath,
		Enabled:     payload.Enabled,
	})
	if err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}

	a.writeAPISuccess(writer, http.StatusCreated, requestID, timestamp, adminServicePayload(service))
}

func (a *App) handleAdminServiceDetail(writer http.ResponseWriter, request *http.Request, serviceID string) {
	requestID, timestamp := a.requestMeta(request)
	token, ok := bearerToken(request)
	if !ok {
		a.writeAPIError(writer, requestID, timestamp, missingBearerTokenError())
		return
	}

	service, err := a.authService.GetAdminService(request.Context(), auth.GetAdminServiceInput{
		AccessToken: token,
		ServiceID:   serviceID,
	})
	if err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}

	a.writeAPISuccess(writer, http.StatusOK, requestID, timestamp, adminServicePayload(service))
}

func (a *App) handleAdminServiceUpdate(writer http.ResponseWriter, request *http.Request, serviceID string) {
	requestID, timestamp := a.requestMeta(request)
	token, ok := bearerToken(request)
	if !ok {
		a.writeAPIError(writer, requestID, timestamp, missingBearerTokenError())
		return
	}

	var payload struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Group       string `json:"group"`
		Protocol    string `json:"protocol"`
		UpstreamURL string `json:"upstreamUrl"`
		PublicPath  string `json:"publicPath"`
	}
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		a.writeAPIError(writer, requestID, timestamp, badJSONError())
		return
	}

	service, err := a.authService.UpdateAdminService(request.Context(), auth.UpdateAdminServiceInput{
		AccessToken: token,
		ServiceID:   serviceID,
		Name:        payload.Name,
		Description: payload.Description,
		Group:       payload.Group,
		Protocol:    payload.Protocol,
		UpstreamURL: payload.UpstreamURL,
		PublicPath:  payload.PublicPath,
	})
	if err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}

	a.writeAPISuccess(writer, http.StatusOK, requestID, timestamp, adminServicePayload(service))
}

func (a *App) handleAdminServiceStatusSet(writer http.ResponseWriter, request *http.Request, serviceID string) {
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

	service, err := a.authService.SetAdminServiceStatus(request.Context(), auth.SetAdminServiceStatusInput{
		AccessToken: token,
		RequestID:   requestID,
		ServiceID:   serviceID,
		Status:      payload.Status,
	})
	if err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}

	a.writeAPISuccess(writer, http.StatusOK, requestID, timestamp, adminServicePayload(service))
}

func (a *App) handleAdminDeviceList(writer http.ResponseWriter, request *http.Request) {
	requestID, timestamp := a.requestMeta(request)
	token, ok := bearerToken(request)
	if !ok {
		a.writeAPIError(writer, requestID, timestamp, missingBearerTokenError())
		return
	}

	result, err := a.authService.ListAdminDevices(request.Context(), auth.ListAdminDevicesInput{
		AccessToken: token,
		Page:        parseIntQuery(request, "page", 1),
		PageSize:    parseIntQuery(request, "pageSize", 20),
		Keyword:     strings.TrimSpace(request.URL.Query().Get("keyword")),
		Status:      strings.TrimSpace(request.URL.Query().Get("status")),
		UserID:      strings.TrimSpace(request.URL.Query().Get("userId")),
	})
	if err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}

	items := make([]map[string]any, 0, len(result.Items))
	for _, device := range result.Items {
		items = append(items, adminDevicePayload(device))
	}

	a.writeAPISuccessWithPagination(writer, http.StatusOK, requestID, timestamp, map[string]any{
		"items": items,
	}, &result.Pagination)
}

func (a *App) handleAdminDeviceDetail(writer http.ResponseWriter, request *http.Request, deviceID string) {
	requestID, timestamp := a.requestMeta(request)
	token, ok := bearerToken(request)
	if !ok {
		a.writeAPIError(writer, requestID, timestamp, missingBearerTokenError())
		return
	}

	device, err := a.authService.GetAdminDevice(request.Context(), auth.GetAdminDeviceInput{
		AccessToken: token,
		DeviceID:    deviceID,
	})
	if err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}

	a.writeAPISuccess(writer, http.StatusOK, requestID, timestamp, adminDevicePayload(device))
}

func (a *App) handleAdminDeviceStatusSet(writer http.ResponseWriter, request *http.Request, deviceID string) {
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

	device, err := a.authService.SetAdminDeviceStatus(request.Context(), auth.SetAdminDeviceStatusInput{
		AccessToken: token,
		RequestID:   requestID,
		DeviceID:    deviceID,
		Status:      payload.Status,
	})
	if err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}

	a.writeAPISuccess(writer, http.StatusOK, requestID, timestamp, adminDevicePayload(device))
}

func (a *App) handleAdminAuditEventList(writer http.ResponseWriter, request *http.Request) {
	requestID, timestamp := a.requestMeta(request)
	token, ok := bearerToken(request)
	if !ok {
		a.writeAPIError(writer, requestID, timestamp, missingBearerTokenError())
		return
	}

	result, err := a.authService.ListAdminAuditEvents(request.Context(), auth.ListAdminAuditEventsInput{
		AccessToken: token,
		Page:        parseIntQuery(request, "page", 1),
		PageSize:    parseIntQuery(request, "pageSize", 20),
		Type:        strings.TrimSpace(request.URL.Query().Get("type")),
		ActorUserID: strings.TrimSpace(request.URL.Query().Get("actorUserId")),
		TargetType:  strings.TrimSpace(request.URL.Query().Get("targetType")),
		TargetID:    strings.TrimSpace(request.URL.Query().Get("targetId")),
		ServiceID:   strings.TrimSpace(request.URL.Query().Get("serviceId")),
		Result:      strings.TrimSpace(request.URL.Query().Get("result")),
	})
	if err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}

	items := make([]map[string]any, 0, len(result.Items))
	for _, event := range result.Items {
		items = append(items, adminAuditEventPayload(event))
	}

	a.writeAPISuccessWithPagination(writer, http.StatusOK, requestID, timestamp, map[string]any{
		"items": items,
	}, &result.Pagination)
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
