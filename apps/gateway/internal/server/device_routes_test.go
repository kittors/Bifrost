package server_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kittors/bifrost/apps/gateway/internal/auth"
	"github.com/kittors/bifrost/apps/gateway/internal/server"
)

// 设备路由测试覆盖注册、挑战和挑战验证这条客户端设备链路。

func TestClientDeviceRegisterChallengeAndVerifyRoutes(t *testing.T) {
	t.Parallel()

	stub := &stubAuthService{
		registerDeviceResult: auth.DeviceResult{
			ID:     "device_registered_01",
			Status: "trusted",
		},
		deviceChallengeResult: auth.DeviceChallengeResult{
			ID:        "challenge_01",
			Challenge: "base64url-challenge",
			ExpiresIn: 120,
		},
		verifyDeviceChallengeResult: auth.DeviceChallengeVerificationResult{
			Verified: true,
		},
	}

	app := server.New(server.Options{
		ReadyCheck: func(ctx context.Context) error {
			return nil
		},
		ReadyTime: "2026-04-17T12:00:00Z",
		Upstreams: map[string]string{},
		Now: func() time.Time {
			return time.Date(2026, time.April, 17, 13, 45, 0, 0, time.UTC)
		},
		RequestID: func() string {
			return "req_device_routes"
		},
		AuthService: stub,
	})

	registerRecorder := httptest.NewRecorder()
	registerRequest := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/client/devices/register",
		strings.NewReader(`{"name":"Alice MacBook Pro","os":"macOS","clientVersion":"1.0.0","publicKey":"public-key","publicKeyFingerprint":"fingerprint"}`),
	)
	registerRequest.Header.Set("Content-Type", "application/json")
	registerRequest.Header.Set("Authorization", "Bearer access-token")
	app.Handler().ServeHTTP(registerRecorder, registerRequest)

	if registerRecorder.Code != http.StatusCreated {
		t.Fatalf("expected register status 201, got %d", registerRecorder.Code)
	}

	if stub.registerDeviceInput.AccessToken != "access-token" {
		t.Fatalf("expected register access token forwarded, got %q", stub.registerDeviceInput.AccessToken)
	}

	challengeRecorder := httptest.NewRecorder()
	challengeRequest := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/client/devices/challenge",
		strings.NewReader(`{"deviceId":"device_registered_01"}`),
	)
	challengeRequest.Header.Set("Content-Type", "application/json")
	challengeRequest.Header.Set("Authorization", "Bearer access-token")
	app.Handler().ServeHTTP(challengeRecorder, challengeRequest)

	if challengeRecorder.Code != http.StatusOK {
		t.Fatalf("expected challenge status 200, got %d", challengeRecorder.Code)
	}

	if stub.deviceChallengeInput.DeviceID != "device_registered_01" {
		t.Fatalf("expected challenge device forwarded, got %q", stub.deviceChallengeInput.DeviceID)
	}

	verifyRecorder := httptest.NewRecorder()
	verifyRequest := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/client/devices/challenge/verify",
		strings.NewReader(`{"challengeId":"challenge_01","signature":"signature"}`),
	)
	verifyRequest.Header.Set("Content-Type", "application/json")
	verifyRequest.Header.Set("Authorization", "Bearer access-token")
	app.Handler().ServeHTTP(verifyRecorder, verifyRequest)

	if verifyRecorder.Code != http.StatusOK {
		t.Fatalf("expected verify status 200, got %d", verifyRecorder.Code)
	}

	if stub.verifyDeviceChallengeInput.ChallengeID != "challenge_01" {
		t.Fatalf("expected verify challenge id forwarded, got %q", stub.verifyDeviceChallengeInput.ChallengeID)
	}
}
