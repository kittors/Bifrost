package server

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"sync"

	"github.com/kittors/bifrost/apps/gateway/internal/auth"
	"github.com/kittors/bifrost/apps/gateway/internal/contracts"
)

// 代理处理器集中承载受控访问入口，避免和后台管理接口混在同一个文件里。
func (a *App) handleServiceProxy(writer http.ResponseWriter, request *http.Request) {
	requestID, timestamp := a.requestMeta(request)
	serviceKey, upstreamPath, ok := parseProxyPath(request.URL.Path)
	if !ok {
		a.writeAPIError(writer, requestID, timestamp, apiError{
			statusCode:  http.StatusNotFound,
			code:        contracts.ErrorCodeGatewayRouteNotFound,
			message:     "proxy route not found",
			userMessage: "访问路径不存在",
		})
		return
	}

	if a.authService == nil {
		a.writeAPIError(writer, requestID, timestamp, apiError{
			statusCode:  http.StatusInternalServerError,
			code:        contracts.ErrorCodeCommonInternalError,
			message:     "auth service is not configured",
			userMessage: "服务暂时不可用，请稍后再试",
		})
		return
	}

	accessToken, accessTicket, ok := proxyCredential(request)
	if !ok {
		a.writeAPIError(writer, requestID, timestamp, missingBearerTokenError())
		return
	}

	target, err := a.authService.ResolveProxyRequest(request.Context(), auth.ResolveProxyRequestInput{
		AccessToken:  accessToken,
		AccessTicket: accessTicket,
		RequestID:    requestID,
		ServiceKey:   serviceKey,
	})
	if err != nil {
		a.writeMappedError(writer, requestID, timestamp, err)
		return
	}

	if isWebSocketUpgrade(request) {
		a.handleWebSocketProxy(writer, request, target, upstreamPath, requestID, timestamp)
		return
	}

	body, err := a.readProxyBody(writer, request)
	if err != nil {
		a.writeAPIError(writer, requestID, timestamp, apiError{
			statusCode:  http.StatusRequestEntityTooLarge,
			code:        contracts.ErrorCodeGatewayRequestTooLarge,
			message:     "request body exceeds proxy limit",
			userMessage: "请求体过大",
		})
		return
	}

	targetURL, err := buildUpstreamURL(target.UpstreamURL, upstreamPath, request.URL.RawQuery)
	if err != nil {
		a.writeAPIError(writer, requestID, timestamp, apiError{
			statusCode:  http.StatusBadGateway,
			code:        contracts.ErrorCodeServiceUpstreamInvalid,
			message:     err.Error(),
			userMessage: "服务暂时不可用，请稍后再试",
		})
		return
	}

	proxyCtx, cancel := context.WithTimeout(request.Context(), a.proxyRequestTimeout())
	defer cancel()

	upstreamRequest, err := http.NewRequestWithContext(proxyCtx, request.Method, targetURL, bytes.NewReader(body))
	if err != nil {
		a.writeAPIError(writer, requestID, timestamp, apiError{
			statusCode:  http.StatusBadGateway,
			code:        contracts.ErrorCodeGatewayBadUpstream,
			message:     err.Error(),
			userMessage: "服务暂时不可用，请稍后再试",
		})
		return
	}

	copyProxyHeaders(upstreamRequest.Header, request.Header)
	upstreamRequest.Header.Set("X-Bifrost-Request-Id", requestID)
	upstreamRequest.Header.Set("X-Bifrost-Service-Key", target.ServiceKey)
	upstreamRequest.Header.Set("X-Bifrost-Service-Id", target.ServiceID)
	upstreamRequest.Header.Set("X-Bifrost-User-Id", target.UserID)
	if target.DeviceID != "" {
		upstreamRequest.Header.Set("X-Bifrost-Device-Id", target.DeviceID)
	}
	if target.AccessSource != "" {
		upstreamRequest.Header.Set("X-Bifrost-Access-Source", target.AccessSource)
	}

	upstreamResponse, err := a.proxyHTTPClient().Do(upstreamRequest)
	if err != nil {
		if auditErr := a.authService.RecordProxyAccessEvent(request.Context(), auth.RecordProxyAccessEventInput{
			RequestID: requestID,
			Type:      contracts.AuditEventTypeServiceAccessUpstreamError,
			UserID:    target.UserID,
			DeviceID:  target.DeviceID,
			ServiceID: target.ServiceID,
			Result:    "failure",
			Summary:   "upstream request failed",
		}); auditErr != nil {
			a.writeMappedError(writer, requestID, timestamp, auditErr)
			return
		}

		if isTimeoutError(err) {
			a.writeAPIError(writer, requestID, timestamp, apiError{
				statusCode:  http.StatusGatewayTimeout,
				code:        contracts.ErrorCodeGatewayUpstreamTimeout,
				message:     err.Error(),
				userMessage: "上游服务响应超时",
			})
			return
		}

		a.writeAPIError(writer, requestID, timestamp, apiError{
			statusCode:  http.StatusBadGateway,
			code:        contracts.ErrorCodeGatewayBadUpstream,
			message:     err.Error(),
			userMessage: "上游服务暂时不可用",
		})
		return
	}
	defer upstreamResponse.Body.Close()

	if auditErr := a.authService.RecordProxyAccessEvent(request.Context(), auth.RecordProxyAccessEventInput{
		RequestID: requestID,
		Type:      contracts.AuditEventTypeServiceAccessGranted,
		UserID:    target.UserID,
		DeviceID:  target.DeviceID,
		ServiceID: target.ServiceID,
		Result:    "success",
		Summary:   "service access granted",
	}); auditErr != nil {
		a.writeMappedError(writer, requestID, timestamp, auditErr)
		return
	}

	copyResponseHeaders(writer.Header(), upstreamResponse.Header)
	writer.WriteHeader(upstreamResponse.StatusCode)
	_, _ = io.Copy(writer, upstreamResponse.Body)
}

// WebSocket 代理需要接管连接生命周期，因此单独放到专门文件，减少普通 HTTP 代理复杂度外溢。
func (a *App) handleWebSocketProxy(
	writer http.ResponseWriter,
	request *http.Request,
	target auth.ResolveProxyRequestResult,
	upstreamPath string,
	requestID string,
	timestamp string,
) {
	targetURL, err := buildUpstreamURL(target.UpstreamURL, upstreamPath, request.URL.RawQuery)
	if err != nil {
		a.writeAPIError(writer, requestID, timestamp, apiError{
			statusCode:  http.StatusBadGateway,
			code:        contracts.ErrorCodeServiceUpstreamInvalid,
			message:     err.Error(),
			userMessage: "服务暂时不可用，请稍后再试",
		})
		return
	}

	hijacker, ok := writer.(http.Hijacker)
	if !ok {
		a.writeAPIError(writer, requestID, timestamp, apiError{
			statusCode:  http.StatusInternalServerError,
			code:        contracts.ErrorCodeCommonInternalError,
			message:     "response writer does not support hijacking",
			userMessage: "服务暂时不可用，请稍后再试",
		})
		return
	}

	upstreamConn, err := a.dialUpstream(request.Context(), targetURL)
	if err != nil {
		_ = a.authService.RecordProxyAccessEvent(request.Context(), auth.RecordProxyAccessEventInput{
			RequestID: requestID,
			Type:      contracts.AuditEventTypeServiceAccessUpstreamError,
			UserID:    target.UserID,
			DeviceID:  target.DeviceID,
			ServiceID: target.ServiceID,
			Result:    "failure",
			Summary:   "websocket upstream dial failed",
		})
		if isTimeoutError(err) {
			a.writeAPIError(writer, requestID, timestamp, apiError{
				statusCode:  http.StatusGatewayTimeout,
				code:        contracts.ErrorCodeGatewayUpstreamTimeout,
				message:     err.Error(),
				userMessage: "上游服务响应超时",
			})
			return
		}
		a.writeAPIError(writer, requestID, timestamp, apiError{
			statusCode:  http.StatusBadGateway,
			code:        contracts.ErrorCodeGatewayBadUpstream,
			message:     err.Error(),
			userMessage: "上游服务暂时不可用",
		})
		return
	}

	upstreamRequest, err := http.NewRequestWithContext(request.Context(), request.Method, targetURL, nil)
	if err != nil {
		_ = upstreamConn.Close()
		a.writeAPIError(writer, requestID, timestamp, apiError{
			statusCode:  http.StatusBadGateway,
			code:        contracts.ErrorCodeGatewayBadUpstream,
			message:     err.Error(),
			userMessage: "服务暂时不可用，请稍后再试",
		})
		return
	}
	copyWebSocketHeaders(upstreamRequest.Header, request.Header)
	upstreamRequest.Host = upstreamRequest.URL.Host
	upstreamRequest.Header.Set("X-Bifrost-Request-Id", requestID)
	upstreamRequest.Header.Set("X-Bifrost-Service-Key", target.ServiceKey)
	upstreamRequest.Header.Set("X-Bifrost-Service-Id", target.ServiceID)
	upstreamRequest.Header.Set("X-Bifrost-User-Id", target.UserID)
	if target.DeviceID != "" {
		upstreamRequest.Header.Set("X-Bifrost-Device-Id", target.DeviceID)
	}

	if err := upstreamRequest.Write(upstreamConn); err != nil {
		_ = upstreamConn.Close()
		a.writeAPIError(writer, requestID, timestamp, apiError{
			statusCode:  http.StatusBadGateway,
			code:        contracts.ErrorCodeGatewayBadUpstream,
			message:     err.Error(),
			userMessage: "上游服务暂时不可用",
		})
		return
	}

	clientConn, clientBuf, err := hijacker.Hijack()
	if err != nil {
		_ = upstreamConn.Close()
		a.writeAPIError(writer, requestID, timestamp, apiError{
			statusCode:  http.StatusInternalServerError,
			code:        contracts.ErrorCodeCommonInternalError,
			message:     err.Error(),
			userMessage: "服务暂时不可用，请稍后再试",
		})
		return
	}

	_ = a.authService.RecordProxyAccessEvent(request.Context(), auth.RecordProxyAccessEventInput{
		RequestID: requestID,
		Type:      contracts.AuditEventTypeServiceAccessGranted,
		UserID:    target.UserID,
		DeviceID:  target.DeviceID,
		ServiceID: target.ServiceID,
		Result:    "success",
		Summary:   "websocket access granted",
	})

	errCh := make(chan error, 2)
	var once sync.Once
	closeConns := func() {
		_ = clientConn.Close()
		_ = upstreamConn.Close()
	}

	go func() {
		_, copyErr := io.Copy(upstreamConn, clientBuf)
		once.Do(closeConns)
		errCh <- copyErr
	}()

	go func() {
		_, copyErr := io.Copy(clientConn, upstreamConn)
		once.Do(closeConns)
		errCh <- copyErr
	}()

	<-errCh
}
