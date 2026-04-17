package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
)

type Options struct {
	ReadyCheck func(ctx context.Context) error
	ReadyTime  string
	Upstreams  map[string]string
}

type App struct {
	handler    http.Handler
	readyCheck func(ctx context.Context) error
	readyTime  string
	upstreams  map[string]string
}

func New(options Options) *App {
	app := &App{
		readyCheck: options.ReadyCheck,
		readyTime:  options.ReadyTime,
		upstreams:  options.Upstreams,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", app.handleIndex)
	mux.HandleFunc("/healthz", app.handleHealthz)
	mux.HandleFunc("/readyz", app.handleReadyz)
	mux.HandleFunc("/debug/upstreams/", app.handleUpstreamProbe)
	app.handler = mux

	return app
}

func (a *App) Handler() http.Handler {
	return a.handler
}

func (a *App) handleIndex(writer http.ResponseWriter, request *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]any{
		"name":      "bifrost-gateway",
		"readyTime": a.readyTime,
		"service":   "gateway",
	})
}

func (a *App) handleHealthz(writer http.ResponseWriter, request *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func (a *App) handleReadyz(writer http.ResponseWriter, request *http.Request) {
	if a.readyCheck != nil {
		if err := a.readyCheck(request.Context()); err != nil {
			writeJSON(writer, http.StatusServiceUnavailable, map[string]any{
				"error":     err.Error(),
				"readyTime": a.readyTime,
				"status":    "not-ready",
				"upstreams": a.upstreams,
			})
			return
		}
	}

	writeJSON(writer, http.StatusOK, map[string]any{
		"readyTime": a.readyTime,
		"status":    "ready",
		"upstreams": a.upstreams,
	})
}

func (a *App) handleUpstreamProbe(writer http.ResponseWriter, request *http.Request) {
	serviceKey := strings.TrimPrefix(request.URL.Path, "/debug/upstreams/")
	target, ok := a.upstreams[serviceKey]
	if !ok {
		writeJSON(writer, http.StatusNotFound, map[string]string{
			"error":      "upstream not configured",
			"serviceKey": serviceKey,
		})
		return
	}

	targetURL := strings.TrimSuffix(target, "/") + "/whoami"
	ctx, cancel := context.WithTimeout(request.Context(), 3*time.Second)
	defer cancel()

	upstreamRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL, nil)
	if err != nil {
		writeJSON(writer, http.StatusBadGateway, map[string]string{
			"error": err.Error(),
		})
		return
	}

	upstreamResponse, err := http.DefaultClient.Do(upstreamRequest)
	if err != nil {
		writeJSON(writer, http.StatusBadGateway, map[string]string{
			"error":      err.Error(),
			"serviceKey": serviceKey,
			"target":     targetURL,
		})
		return
	}
	defer upstreamResponse.Body.Close()

	var upstreamBody map[string]any
	if err := json.NewDecoder(upstreamResponse.Body).Decode(&upstreamBody); err != nil {
		writeJSON(writer, http.StatusBadGateway, map[string]string{
			"error": err.Error(),
		})
		return
	}

	writeJSON(writer, http.StatusOK, map[string]any{
		"serviceKey": serviceKey,
		"target":     targetURL,
		"upstream":   upstreamBody,
	})
}

func writeJSON(writer http.ResponseWriter, statusCode int, payload any) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(statusCode)
	_ = json.NewEncoder(writer).Encode(payload)
}

func IsServerClosed(err error) bool {
	return errors.Is(err, http.ErrServerClosed)
}
