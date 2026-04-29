package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/kittors/bifrost/apps/gateway/internal/contracts"
)

// 中间件链单独存放，便于后续继续引入限流、CORS 或安全头时保持入口清晰。
func (a *App) requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requestID := strings.TrimSpace(request.Header.Get("X-Request-Id"))
		if requestID == "" {
			requestID = a.newRequestID()
		}

		writer.Header().Set("X-Request-Id", requestID)
		request = request.WithContext(context.WithValue(request.Context(), requestIDContextKey, requestID))
		next.ServeHTTP(writer, request)
	})
}

func (a *App) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer func() {
			recovered := recover()
			if recovered == nil {
				return
			}

			requestID, timestamp := a.requestMeta(request)
			if a.logger != nil {
				a.logger.Error("request panic recovered",
					"request_id", requestID,
					"path", request.URL.Path,
					"method", request.Method,
					"panic", recovered,
				)
			}

			a.writeAPIError(writer, requestID, timestamp, apiError{
				statusCode:  http.StatusInternalServerError,
				code:        contracts.ErrorCodeCommonInternalError,
				message:     fmt.Sprintf("panic: %v", recovered),
				userMessage: "服务暂时不可用，请稍后再试",
			})
		}()

		next.ServeHTTP(writer, request)
	})
}

func (a *App) accessLogMiddleware(next http.Handler) http.Handler {
	if a.logger == nil {
		return next
	}

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		recorder := &statusRecorder{ResponseWriter: writer, statusCode: http.StatusOK}
		startedAt := time.Now()
		next.ServeHTTP(recorder, request)

		a.logger.Info("http request completed",
			"request_id", requestIDFromContext(request.Context()),
			"method", request.Method,
			"path", request.URL.Path,
			"status", recorder.statusCode,
			"duration_ms", time.Since(startedAt).Milliseconds(),
		)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}
