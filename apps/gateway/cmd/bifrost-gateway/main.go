package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type gatewayServer struct {
	client    *http.Client
	readyTime string
	upstreams map[string]string
}

func main() {
	server := &gatewayServer{
		client: &http.Client{
			Timeout: 3 * time.Second,
		},
		readyTime: time.Now().UTC().Format(time.RFC3339),
		upstreams: map[string]string{
			"gitlab":         envOrDefault("BIFROST_UPSTREAM_GITLAB", "http://mock-gitlab:8080"),
			"jenkins":        envOrDefault("BIFROST_UPSTREAM_JENKINS", "http://mock-jenkins:8080"),
			"docs":           envOrDefault("BIFROST_UPSTREAM_DOCS", "http://mock-docs:8080"),
			"internal-admin": envOrDefault("BIFROST_UPSTREAM_INTERNAL_ADMIN", "http://mock-internal-admin:8080"),
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", server.handleIndex)
	mux.HandleFunc("/healthz", server.handleHealthz)
	mux.HandleFunc("/readyz", server.handleReadyz)
	mux.HandleFunc("/debug/upstreams/", server.handleUpstreamProbe)

	listenAddress := ":" + envOrDefault("PORT", "8080")
	log.Printf("bifrost gateway listening on %s", listenAddress)

	if err := http.ListenAndServe(listenAddress, mux); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func envOrDefault(key string, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}

	return fallback
}

func (s *gatewayServer) handleIndex(writer http.ResponseWriter, request *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]any{
		"name":      "bifrost-gateway",
		"readyTime": s.readyTime,
		"service":   "gateway",
	})
}

func (s *gatewayServer) handleHealthz(writer http.ResponseWriter, request *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func (s *gatewayServer) handleReadyz(writer http.ResponseWriter, request *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]any{
		"status":    "ready",
		"readyTime": s.readyTime,
		"upstreams": s.upstreams,
	})
}

func (s *gatewayServer) handleUpstreamProbe(writer http.ResponseWriter, request *http.Request) {
	serviceKey := strings.TrimPrefix(request.URL.Path, "/debug/upstreams/")
	target, ok := s.upstreams[serviceKey]
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

	upstreamResponse, err := s.client.Do(upstreamRequest)
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
	if err := json.NewEncoder(writer).Encode(payload); err != nil {
		log.Printf("encode response: %v", err)
	}
}
