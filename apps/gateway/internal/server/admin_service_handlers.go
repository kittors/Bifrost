package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/kittors/bifrost/apps/gateway/internal/auth"
)

// 服务管理 handler 单独收敛，便于后续继续补充上游连通性与发布状态能力。
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

func (a *App) handleAdminServiceList(writer http.ResponseWriter, request *http.Request) {
	requestID, timestamp := a.requestMeta(request)
	token, ok := bearerToken(request)
	if !ok {
		a.writeAPIError(writer, requestID, timestamp, missingBearerTokenError())
		return
	}

	page, pageSize, queryErr := parsePaginationQuery(request)
	if queryErr != nil {
		a.writeAPIError(writer, requestID, timestamp, *queryErr)
		return
	}

	result, err := a.authService.ListAdminServices(request.Context(), auth.ListAdminServicesInput{
		AccessToken: token,
		Page:        page,
		PageSize:    pageSize,
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
