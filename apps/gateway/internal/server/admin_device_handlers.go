package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/kittors/bifrost/apps/gateway/internal/auth"
)

// 设备管理 handler 只关注设备列表、详情与状态变更。
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

func (a *App) handleAdminDeviceList(writer http.ResponseWriter, request *http.Request) {
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

	result, err := a.authService.ListAdminDevices(request.Context(), auth.ListAdminDevicesInput{
		AccessToken: token,
		Page:        page,
		PageSize:    pageSize,
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
