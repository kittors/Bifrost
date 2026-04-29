package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
)

// 代理底层工具收敛到这里，避免和业务处理器互相穿插。
func (a *App) readProxyBody(writer http.ResponseWriter, request *http.Request) ([]byte, error) {
	if request.Body == nil {
		return nil, nil
	}

	limitedBody := http.MaxBytesReader(writer, request.Body, a.proxyBodyLimit())
	defer limitedBody.Close()
	return io.ReadAll(limitedBody)
}

func (a *App) dialUpstream(ctx context.Context, rawURL string) (net.Conn, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse upstream url: %w", err)
	}

	port := parsed.Port()
	if port == "" {
		switch parsed.Scheme {
		case "http":
			port = "80"
		case "https":
			port = "443"
		default:
			return nil, fmt.Errorf("unsupported websocket upstream scheme %q", parsed.Scheme)
		}
	}

	dialer := &net.Dialer{Timeout: a.proxyRequestTimeout()}
	return dialer.DialContext(ctx, "tcp", net.JoinHostPort(parsed.Hostname(), port))
}

func buildUpstreamURL(baseURL string, upstreamPath string, rawQuery string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(baseURL))
	if err != nil {
		return "", fmt.Errorf("parse upstream url: %w", err)
	}

	parsed.Path = strings.TrimRight(parsed.Path, "/") + upstreamPath
	parsed.RawQuery = rawQuery
	return parsed.String(), nil
}

func copyProxyHeaders(target http.Header, source http.Header) {
	for key, values := range source {
		lowerKey := strings.ToLower(key)
		switch lowerKey {
		case "authorization", "host", "connection", "proxy-connection", "keep-alive", "te", "trailer", "transfer-encoding", "upgrade":
			continue
		}

		for _, value := range values {
			target.Add(key, value)
		}
	}
}

func copyWebSocketHeaders(target http.Header, source http.Header) {
	for key, values := range source {
		lowerKey := strings.ToLower(key)
		switch lowerKey {
		case "authorization", "host":
			continue
		}

		for _, value := range values {
			target.Add(key, value)
		}
	}
}

func copyResponseHeaders(target http.Header, source http.Header) {
	for key, values := range source {
		if strings.EqualFold(key, "X-Request-Id") {
			continue
		}
		for _, value := range values {
			target.Add(key, value)
		}
	}
}

func isTimeoutError(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	var netErr net.Error
	return errors.As(err, &netErr) && netErr.Timeout()
}
