package server

import (
	"net/http"
	"strings"

	"github.com/kittors/bifrost/apps/gateway/internal/auth"
)

// 审计查询 handler 只负责读取审计事件列表，方便后续扩展更多审计视图。
func (a *App) handleAdminAuditEvents(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	a.handleAdminAuditEventList(writer, request)
}

func (a *App) handleAdminAuditEventList(writer http.ResponseWriter, request *http.Request) {
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

	result, err := a.authService.ListAdminAuditEvents(request.Context(), auth.ListAdminAuditEventsInput{
		AccessToken: token,
		Page:        page,
		PageSize:    pageSize,
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
