package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/kittors/bifrost/apps/gateway/internal/auth"
	"github.com/kittors/bifrost/apps/gateway/internal/contracts"
)

// 响应封装统一收敛到这里，避免每个 handler 自己拼接结构导致风格漂移。
func writeJSON(writer http.ResponseWriter, statusCode int, payload any) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(statusCode)
	_ = json.NewEncoder(writer).Encode(payload)
}

func (a *App) requestMeta(request *http.Request) (string, string) {
	requestID := requestIDFromContext(request.Context())
	if requestID == "" {
		requestID = strings.TrimSpace(request.Header.Get("X-Request-Id"))
	}
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

func (a *App) writeAPISuccessWithPagination(
	writer http.ResponseWriter,
	statusCode int,
	requestID string,
	timestamp string,
	data any,
	pagination *contracts.Pagination,
) {
	writer.Header().Set("X-Request-Id", requestID)
	meta := map[string]any{
		"requestId": requestID,
		"timestamp": timestamp,
	}
	if pagination != nil {
		meta["pagination"] = pagination
	}
	writeJSON(writer, statusCode, map[string]any{
		"success": true,
		"data":    data,
		"meta":    meta,
		"error":   nil,
	})
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

func requestIDFromContext(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDContextKey).(string)
	return strings.TrimSpace(requestID)
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

func adminUserPayload(user auth.AdminUser) map[string]any {
	return map[string]any{
		"id":          user.ID,
		"username":    user.Username,
		"displayName": user.DisplayName,
		"email":       user.Email,
		"status":      user.Status,
		"roles":       user.Roles,
	}
}

func adminRolePayload(role auth.AdminRole) map[string]any {
	return map[string]any{
		"id":          role.ID,
		"name":        role.Name,
		"displayName": role.DisplayName,
		"description": role.Description,
	}
}

func adminServicePayload(service auth.AdminService) map[string]any {
	return map[string]any{
		"id":          service.ID,
		"key":         service.Key,
		"name":        service.Name,
		"description": service.Description,
		"group":       service.Group,
		"protocol":    service.Protocol,
		"upstreamUrl": service.UpstreamURL,
		"publicPath":  service.PublicPath,
		"status":      service.Status,
	}
}

func adminDevicePayload(device auth.AdminDevice) map[string]any {
	return map[string]any{
		"id":                   device.ID,
		"userId":               device.UserID,
		"userUsername":         device.UserUsername,
		"name":                 device.Name,
		"os":                   device.OS,
		"clientVersion":        device.ClientVersion,
		"publicKeyFingerprint": device.PublicKeyFingerprint,
		"status":               device.Status,
	}
}

func adminAuditEventPayload(event auth.AdminAuditEvent) map[string]any {
	return map[string]any{
		"id":          event.ID,
		"requestId":   event.RequestID,
		"type":        event.Type,
		"actorUserId": event.ActorUserID,
		"targetType":  event.TargetType,
		"targetId":    event.TargetID,
		"serviceId":   event.ServiceID,
		"result":      event.Result,
		"summary":     event.Summary,
	}
}
