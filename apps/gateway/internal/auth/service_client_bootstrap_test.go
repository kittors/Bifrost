package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/kittors/bifrost/apps/gateway/internal/auth"
	"github.com/kittors/bifrost/apps/gateway/internal/database"
)

// 客户端首登 bootstrap 测试聚焦“首次登录即可安全绑定设备”的闭环。
func TestServiceBootstrapClientDeviceCreatesTrustedDeviceAndSession(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dsn := createTestDatabase(t, ctx)
	if err := database.MigrateUp(ctx, dsn); err != nil {
		t.Fatalf("migrate up: %v", err)
	}
	if err := database.SeedPhase1(ctx, dsn); err != nil {
		t.Fatalf("seed phase 1: %v", err)
	}

	db := openDB(t, dsn)
	now := time.Date(2026, time.April, 18, 3, 30, 0, 0, time.UTC)
	service := auth.Service{
		DB:               db,
		PasswordHasher:   auth.DefaultPasswordHasher(),
		TokenIssuer:      auth.TokenIssuer{Secret: []byte("bootstrap-secret-bootstrap-secret"), TTL: 15 * time.Minute, Now: func() time.Time { return now }},
		Now:              func() time.Time { return now },
		RefreshTokenTTL:  7 * 24 * time.Hour,
		SessionIDFactory: func() (string, error) { return "session_client_bootstrap_01", nil },
		DeviceIDFactory:  func() (string, error) { return "device_client_bootstrap_01", nil },
	}

	publicKey, _, fingerprint := generateEd25519Material(t)

	result, err := service.BootstrapClientDevice(ctx, auth.BootstrapClientDeviceInput{
		Username:             "alice",
		Password:             "ChangeMe123!",
		DeviceName:           "Alice MacBook Pro",
		DeviceOS:             "macOS",
		ClientVersion:        "0.1.0",
		PublicKey:            publicKey,
		PublicKeyFingerprint: fingerprint,
		RequestID:            "req_client_bootstrap_01",
	})
	if err != nil {
		t.Fatalf("bootstrap client device: %v", err)
	}

	if result.Device.ID != "device_client_bootstrap_01" {
		t.Fatalf("expected bootstrap device id, got %#v", result.Device)
	}
	if result.Device.Status != "trusted" {
		t.Fatalf("expected trusted bootstrap device, got %#v", result.Device)
	}
	if result.User.Username != "alice" {
		t.Fatalf("expected bootstrap login for alice, got %#v", result.User)
	}
	if result.AccessToken == "" || result.RefreshToken == "" {
		t.Fatalf("expected issued session tokens, got %#v", result)
	}
}
