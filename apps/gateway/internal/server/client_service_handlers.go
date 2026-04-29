package server

import (
	"net/http"
	"strings"

	"github.com/kittors/bifrost/apps/gateway/internal/auth"
)

// 客户端服务 handler 负责服务目录、详情和短期访问 URL。

func (a *App) handleClientServices(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	requestID, timestamp := a.requestMeta(request)
	token, ok := bearerToken(request)
	if !ok {
		a.writeAPIError(writer, requestID, timestamp, missingBearerTokenError())
		return
	}

	services, err := a.authService.ListClientServices(request.Context(), auth.ListClientServicesInput{
		AccessToken: token,
		Keyword:     strings.TrimSpace(request.URL.Query().Get("keyword")),
		Group:       strings.TrimSpace(request.URL.Query().Get("group")),
	})
	if err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}

	items := make([]map[string]any, 0, len(services))
	for _, service := range services {
		items = append(items, clientServicePayload(service))
	}

	a.writeAPISuccess(writer, http.StatusOK, requestID, timestamp, map[string]any{
		"items": items,
	})
}

func (a *App) handleClientServiceByID(writer http.ResponseWriter, request *http.Request) {
	serviceID, action, ok := parseClientServicePath(request.URL.Path)
	if !ok {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	if action == "access-url" {
		a.handleClientServiceAccessURL(writer, request, serviceID)
		return
	}

	if request.Method != http.MethodGet {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	requestID, timestamp := a.requestMeta(request)
	token, ok := bearerToken(request)
	if !ok {
		a.writeAPIError(writer, requestID, timestamp, missingBearerTokenError())
		return
	}

	service, err := a.authService.GetClientService(request.Context(), auth.GetClientServiceInput{
		AccessToken: token,
		ServiceID:   serviceID,
	})
	if err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}

	a.writeAPISuccess(writer, http.StatusOK, requestID, timestamp, clientServicePayload(service))
}

func (a *App) handleClientServiceAccessURL(writer http.ResponseWriter, request *http.Request, serviceID string) {
	if request.Method != http.MethodPost {
		writer.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	requestID, timestamp := a.requestMeta(request)
	token, ok := bearerToken(request)
	if !ok {
		a.writeAPIError(writer, requestID, timestamp, missingBearerTokenError())
		return
	}

	result, err := a.authService.CreateServiceAccessURL(request.Context(), auth.CreateServiceAccessURLInput{
		AccessToken: token,
		ServiceID:   serviceID,
	})
	if err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}

	http.SetCookie(writer, &http.Cookie{
		Name:     serviceAccessCookieName,
		Value:    result.AccessTicket,
		Path:     result.PublicPath,
		MaxAge:   result.ExpiresIn,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   requestIsSecure(request),
	})

	a.writeAPISuccess(writer, http.StatusOK, requestID, timestamp, map[string]any{
		"url":       absoluteURL(request, result.PublicPath),
		"expiresIn": result.ExpiresIn,
	})
}
