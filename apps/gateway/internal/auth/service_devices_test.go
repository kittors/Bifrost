package auth_test

import (
	"context"
	"crypto/ed25519"
	"database/sql"
	"encoding/base64"
	"testing"
	"time"

	"github.com/kittors/bifrost/apps/gateway/internal/auth"
	"github.com/kittors/bifrost/apps/gateway/internal/database"
)

// 设备测试聚焦注册、挑战和签名验证，后续设备信任策略可在这里扩展。

func TestServiceRegisterDeviceCreatesTrustedDeviceForCurrentUser(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dsn := createTestDatabase(t, ctx)
	if err := database.MigrateUp(ctx, dsn); err != nil {
		t.Fatalf("migrate up: %v", err)
	}

	db := openDB(t, dsn)
	now := time.Date(2026, time.April, 17, 13, 30, 0, 0, time.UTC)
	service := auth.Service{
		DB:               db,
		PasswordHasher:   auth.DefaultPasswordHasher(),
		TokenIssuer:      auth.TokenIssuer{Secret: []byte("0123456789abcdef0123456789abcdef"), TTL: 15 * time.Minute, Now: func() time.Time { return now }},
		Now:              func() time.Time { return now },
		RefreshTokenTTL:  7 * 24 * time.Hour,
		SessionIDFactory: func() (string, error) { return "session_register_01", nil },
		DeviceIDFactory:  func() (string, error) { return "device_registered_01", nil },
	}

	insertUserWithRoles(t, ctx, db, "user_alice", "alice", "Alice", "correct horse battery staple", []roleSeed{{id: "role_developer", name: "developer", displayName: "Developer"}})

	accessToken := issueAccessTokenForTest(t, service.TokenIssuer, "user_alice", "", "session_register_01")

	publicKey, _, fingerprint := generateEd25519Material(t)

	device, err := service.RegisterDevice(ctx, auth.RegisterDeviceInput{
		AccessToken:          accessToken,
		Name:                 "Alice MacBook Pro",
		OS:                   "macOS",
		ClientVersion:        "1.0.0",
		PublicKey:            publicKey,
		PublicKeyFingerprint: fingerprint,
	})
	if err != nil {
		t.Fatalf("register device: %v", err)
	}

	if device.ID != "device_registered_01" {
		t.Fatalf("expected device id device_registered_01, got %q", device.ID)
	}

	if device.Status != "trusted" {
		t.Fatalf("expected trusted status, got %q", device.Status)
	}

	var storedUserID string
	var storedFingerprint string
	if err := db.QueryRowContext(
		ctx,
		`SELECT user_id, public_key_fingerprint
		FROM devices
		WHERE id = $1`,
		"device_registered_01",
	).Scan(&storedUserID, &storedFingerprint); err != nil {
		t.Fatalf("query registered device: %v", err)
	}

	if storedUserID != "user_alice" {
		t.Fatalf("expected device bound to user_alice, got %q", storedUserID)
	}

	if storedFingerprint != fingerprint {
		t.Fatalf("expected fingerprint %q, got %q", fingerprint, storedFingerprint)
	}
}

func TestServiceCreateAndVerifyDeviceChallenge(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dsn := createTestDatabase(t, ctx)
	if err := database.MigrateUp(ctx, dsn); err != nil {
		t.Fatalf("migrate up: %v", err)
	}

	db := openDB(t, dsn)
	now := time.Date(2026, time.April, 17, 13, 30, 0, 0, time.UTC)
	service := auth.Service{
		DB:                 db,
		PasswordHasher:     auth.DefaultPasswordHasher(),
		TokenIssuer:        auth.TokenIssuer{Secret: []byte("0123456789abcdef0123456789abcdef"), TTL: 15 * time.Minute, Now: func() time.Time { return now }},
		Now:                func() time.Time { return now },
		RefreshTokenTTL:    7 * 24 * time.Hour,
		SessionIDFactory:   func() (string, error) { return "session_challenge_01", nil },
		ChallengeIDFactory: func() (string, error) { return "challenge_01", nil },
		ChallengeTTL:       2 * time.Minute,
	}

	insertUserWithRoles(t, ctx, db, "user_alice", "alice", "Alice", "correct horse battery staple", []roleSeed{{id: "role_developer", name: "developer", displayName: "Developer"}})
	publicKey, privateKey, fingerprint := generateEd25519Material(t)
	insertDeviceWithKey(t, ctx, db, "device_alice_sign", "user_alice", "trusted", publicKey, fingerprint)

	accessToken := issueAccessTokenForTest(t, service.TokenIssuer, "user_alice", "", "session_challenge_01")

	challenge, err := service.CreateDeviceChallenge(ctx, auth.CreateDeviceChallengeInput{
		AccessToken: accessToken,
		DeviceID:    "device_alice_sign",
	})
	if err != nil {
		t.Fatalf("create device challenge: %v", err)
	}

	if challenge.ID != "challenge_01" {
		t.Fatalf("expected challenge id challenge_01, got %q", challenge.ID)
	}

	if challenge.ExpiresIn != 120 {
		t.Fatalf("expected expiresIn 120, got %d", challenge.ExpiresIn)
	}

	rawChallenge, err := base64.RawURLEncoding.DecodeString(challenge.Challenge)
	if err != nil {
		t.Fatalf("decode challenge: %v", err)
	}

	signature := ed25519.Sign(privateKey, rawChallenge)
	verified, err := service.VerifyDeviceChallenge(ctx, auth.VerifyDeviceChallengeInput{
		AccessToken: accessToken,
		ChallengeID: "challenge_01",
		Signature:   base64.RawURLEncoding.EncodeToString(signature),
	})
	if err != nil {
		t.Fatalf("verify device challenge: %v", err)
	}

	if !verified.Verified {
		t.Fatal("expected verified challenge")
	}

	var verifiedAt sql.NullTime
	if err := db.QueryRowContext(
		ctx,
		`SELECT verified_at
		FROM device_challenges
		WHERE id = $1`,
		"challenge_01",
	).Scan(&verifiedAt); err != nil {
		t.Fatalf("query verified challenge: %v", err)
	}

	if !verifiedAt.Valid {
		t.Fatal("expected challenge verified_at to be set")
	}
}
