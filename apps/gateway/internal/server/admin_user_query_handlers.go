package server

import (
	"net/http"
	"strings"

	"github.com/kittors/bifrost/apps/gateway/internal/auth"
)

// 后台用户查询 handler 只负责列表、详情与覆盖策略读取。
func (a *App) handleAdminUserDetail(writer http.ResponseWriter, request *http.Request, userID string) {
	requestID, timestamp := a.requestMeta(request)
	token, ok := bearerToken(request)
	if !ok {
		a.writeAPIError(writer, requestID, timestamp, missingBearerTokenError())
		return
	}

	user, err := a.authService.GetAdminUser(request.Context(), auth.GetAdminUserInput{
		AccessToken: token,
		UserID:      userID,
	})
	if err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}

	a.writeAPISuccess(writer, http.StatusOK, requestID, timestamp, adminUserPayload(user))
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

func (a *App) handleAdminUserServiceOverridesList(writer http.ResponseWriter, request *http.Request, userID string) {
	requestID, timestamp := a.requestMeta(request)
	token, ok := bearerToken(request)
	if !ok {
		a.writeAPIError(writer, requestID, timestamp, missingBearerTokenError())
		return
	}

	overrides, err := a.authService.ListUserServiceOverrides(request.Context(), auth.ListUserServiceOverridesInput{
		AccessToken: token,
		UserID:      userID,
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
