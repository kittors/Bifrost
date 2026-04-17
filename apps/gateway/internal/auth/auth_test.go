package auth_test

import (
	"strings"
	"testing"
	"time"

	"github.com/kittors/bifrost/apps/gateway/internal/auth"
)

func TestPasswordHasherHashAndVerify(t *testing.T) {
	t.Parallel()

	hasher := auth.DefaultPasswordHasher()

	hash, err := hasher.Hash("correct horse battery staple")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	if !strings.HasPrefix(hash, "$argon2id$") {
		t.Fatalf("expected argon2id encoded hash, got %q", hash)
	}

	ok, err := hasher.Verify(hash, "correct horse battery staple")
	if err != nil {
		t.Fatalf("verify password: %v", err)
	}

	if !ok {
		t.Fatal("expected password verification to succeed")
	}
}

func TestPasswordHasherRejectsWrongPassword(t *testing.T) {
	t.Parallel()

	hasher := auth.DefaultPasswordHasher()

	hash, err := hasher.Hash("correct horse battery staple")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	ok, err := hasher.Verify(hash, "wrong-password")
	if err != nil {
		t.Fatalf("verify wrong password: %v", err)
	}

	if ok {
		t.Fatal("expected password verification to fail")
	}
}

func TestPasswordHasherRejectsMalformedHash(t *testing.T) {
	t.Parallel()

	hasher := auth.DefaultPasswordHasher()

	if _, err := hasher.Verify("not-a-valid-hash", "irrelevant"); err == nil {
		t.Fatal("expected malformed hash verification to fail")
	}
}

func TestTokenIssuerIssueAndVerifyAccessToken(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.April, 17, 13, 0, 0, 0, time.UTC)
	issuer := auth.TokenIssuer{
		Secret: []byte("0123456789abcdef0123456789abcdef"),
		TTL:    15 * time.Minute,
		Now: func() time.Time {
			return now
		},
	}

	token, expiresAt, err := issuer.IssueAccessToken(auth.AccessTokenClaims{
		UserID:            "user_admin",
		DeviceID:          "device_admin",
		SessionID:         "session_admin",
		PermissionVersion: 3,
	})
	if err != nil {
		t.Fatalf("issue access token: %v", err)
	}

	if want := now.Add(15 * time.Minute); !expiresAt.Equal(want) {
		t.Fatalf("expected expiresAt %s, got %s", want, expiresAt)
	}

	claims, err := issuer.VerifyAccessToken(token)
	if err != nil {
		t.Fatalf("verify access token: %v", err)
	}

	if claims.UserID != "user_admin" {
		t.Fatalf("expected user id user_admin, got %q", claims.UserID)
	}

	if claims.DeviceID != "device_admin" {
		t.Fatalf("expected device id device_admin, got %q", claims.DeviceID)
	}

	if claims.SessionID != "session_admin" {
		t.Fatalf("expected session id session_admin, got %q", claims.SessionID)
	}

	if claims.PermissionVersion != 3 {
		t.Fatalf("expected permission version 3, got %d", claims.PermissionVersion)
	}
}

func TestTokenIssuerAllowsAdminAccessTokenWithoutDeviceID(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.April, 17, 13, 0, 0, 0, time.UTC)
	issuer := auth.TokenIssuer{
		Secret: []byte("0123456789abcdef0123456789abcdef"),
		TTL:    15 * time.Minute,
		Now: func() time.Time {
			return now
		},
	}

	token, _, err := issuer.IssueAccessToken(auth.AccessTokenClaims{
		UserID:            "user_admin",
		SessionID:         "session_admin",
		PermissionVersion: 1,
	})
	if err != nil {
		t.Fatalf("issue access token: %v", err)
	}

	claims, err := issuer.VerifyAccessToken(token)
	if err != nil {
		t.Fatalf("verify access token without device id: %v", err)
	}

	if claims.DeviceID != "" {
		t.Fatalf("expected empty device id, got %q", claims.DeviceID)
	}
}

func TestTokenIssuerRejectsExpiredAccessToken(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.April, 17, 13, 0, 0, 0, time.UTC)
	issuer := auth.TokenIssuer{
		Secret: []byte("0123456789abcdef0123456789abcdef"),
		TTL:    15 * time.Minute,
		Now: func() time.Time {
			return now
		},
	}

	token, _, err := issuer.IssueAccessToken(auth.AccessTokenClaims{
		UserID:            "user_admin",
		DeviceID:          "device_admin",
		SessionID:         "session_admin",
		PermissionVersion: 1,
	})
	if err != nil {
		t.Fatalf("issue access token: %v", err)
	}

	issuer.Now = func() time.Time {
		return now.Add(16 * time.Minute)
	}

	if _, err := issuer.VerifyAccessToken(token); err == nil {
		t.Fatal("expected expired access token verification to fail")
	}
}

func TestTokenIssuerRejectsTamperedAccessToken(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.April, 17, 13, 0, 0, 0, time.UTC)
	issuer := auth.TokenIssuer{
		Secret: []byte("0123456789abcdef0123456789abcdef"),
		TTL:    15 * time.Minute,
		Now: func() time.Time {
			return now
		},
	}

	token, _, err := issuer.IssueAccessToken(auth.AccessTokenClaims{
		UserID:            "user_admin",
		DeviceID:          "device_admin",
		SessionID:         "session_admin",
		PermissionVersion: 1,
	})
	if err != nil {
		t.Fatalf("issue access token: %v", err)
	}

	tampered := token[:len(token)-1] + "x"
	if _, err := issuer.VerifyAccessToken(tampered); err == nil {
		t.Fatal("expected tampered access token verification to fail")
	}
}

func TestRefreshTokenUtilitiesGenerateAndHash(t *testing.T) {
	t.Parallel()

	tokenOne, err := auth.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("generate refresh token 1: %v", err)
	}

	tokenTwo, err := auth.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("generate refresh token 2: %v", err)
	}

	if tokenOne == tokenTwo {
		t.Fatal("expected refresh tokens to be unique")
	}

	hashOne := auth.HashRefreshToken(tokenOne)
	hashOneRepeat := auth.HashRefreshToken(tokenOne)
	hashTwo := auth.HashRefreshToken(tokenTwo)

	if hashOne == "" {
		t.Fatal("expected non-empty refresh token hash")
	}

	if hashOne != hashOneRepeat {
		t.Fatal("expected refresh token hash to be deterministic")
	}

	if hashOne == hashTwo {
		t.Fatal("expected different refresh tokens to have different hashes")
	}
}
