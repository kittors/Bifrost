package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/kittors/bifrost/apps/gateway/internal/contracts"
)

// 路径、鉴权头和查询参数解析统一放在这里，避免 handler 内部掺杂过多样板代码。
func bearerToken(request *http.Request) (string, bool) {
	header := strings.TrimSpace(request.Header.Get("Authorization"))
	value, ok := strings.CutPrefix(header, "Bearer ")
	if !ok || strings.TrimSpace(value) == "" {
		return "", false
	}
	return strings.TrimSpace(value), true
}

func missingBearerTokenError() apiError {
	return apiError{
		statusCode:  http.StatusUnauthorized,
		code:        contracts.ErrorCodeAuthInvalidToken,
		message:     "bearer token is required",
		userMessage: "登录状态已失效，请重新登录",
	}
}

func badJSONError() apiError {
	return apiError{
		statusCode:  http.StatusBadRequest,
		code:        contracts.ErrorCodeCommonBadRequest,
		message:     "request body must be valid JSON",
		userMessage: "请求参数不正确",
	}
}

func parseClientServicePath(path string) (string, string, bool) {
	remaining := strings.TrimPrefix(path, "/api/v1/client/services/")
	if remaining == path || remaining == "" {
		return "", "", false
	}

	parts := strings.Split(strings.Trim(remaining, "/"), "/")
	if len(parts) == 1 {
		return parts[0], "", parts[0] != ""
	}
	if len(parts) == 2 && parts[1] == "access-url" {
		return parts[0], parts[1], parts[0] != ""
	}
	return "", "", false
}

func parseAdminUserPath(path string) (string, string, bool) {
	remaining := strings.TrimPrefix(path, "/api/v1/admin/users/")
	if remaining == path || remaining == "" {
		return "", "", false
	}

	parts := strings.Split(strings.Trim(remaining, "/"), "/")
	if len(parts) == 1 {
		return parts[0], "", parts[0] != ""
	}
	if len(parts) == 2 && parts[1] == "service-overrides" {
		return parts[0], parts[1], parts[0] != ""
	}
	if len(parts) == 2 && (parts[1] == "reset-password" || parts[1] == "status") {
		return parts[0], parts[1], parts[0] != ""
	}
	return "", "", false
}

func parseAdminRolePath(path string) (string, string, bool) {
	remaining := strings.TrimPrefix(path, "/api/v1/admin/roles/")
	if remaining == path || remaining == "" {
		return "", "", false
	}

	parts := strings.Split(strings.Trim(remaining, "/"), "/")
	if len(parts) == 2 && parts[1] == "services" {
		return parts[0], parts[1], parts[0] != ""
	}
	return "", "", false
}

func parseProxyPath(path string) (string, string, bool) {
	remaining := strings.TrimPrefix(path, "/s/")
	if remaining == path || remaining == "" {
		return "", "", false
	}

	parts := strings.SplitN(strings.TrimPrefix(remaining, "/"), "/", 2)
	serviceKey := strings.TrimSpace(parts[0])
	if serviceKey == "" {
		return "", "", false
	}

	if len(parts) == 1 {
		return serviceKey, "/", true
	}

	return serviceKey, "/" + parts[1], true
}

func proxyCredential(request *http.Request) (string, string, bool) {
	if token, ok := bearerToken(request); ok {
		return token, "", true
	}

	cookie, err := request.Cookie(serviceAccessCookieName)
	if err == nil && strings.TrimSpace(cookie.Value) != "" {
		return "", strings.TrimSpace(cookie.Value), true
	}

	return "", "", false
}

func requestIsSecure(request *http.Request) bool {
	if request.TLS != nil {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(request.Header.Get("X-Forwarded-Proto")), "https")
}

func isWebSocketUpgrade(request *http.Request) bool {
	return strings.EqualFold(strings.TrimSpace(request.Header.Get("Upgrade")), "websocket") &&
		strings.Contains(strings.ToLower(request.Header.Get("Connection")), "upgrade")
}

func absoluteURL(request *http.Request, publicPath string) string {
	scheme := strings.TrimSpace(request.Header.Get("X-Forwarded-Proto"))
	if scheme == "" {
		scheme = "http"
	}
	return scheme + "://" + request.Host + strings.TrimRight(publicPath, "/") + "/"
}

func parseIntQuery(request *http.Request, key string, fallback int) int {
	value := strings.TrimSpace(request.URL.Query().Get(key))
	if value == "" {
		return fallback
	}
	var parsed int
	if _, err := fmt.Sscanf(value, "%d", &parsed); err != nil || parsed <= 0 {
		return fallback
	}
	return parsed
}
