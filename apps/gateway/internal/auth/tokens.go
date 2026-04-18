package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	refreshTokenLength = 32
	challengeLength    = 32
)

type AccessTokenClaims struct {
	UserID            string
	DeviceID          string
	SessionID         string
	PermissionVersion int
	IssuedAt          time.Time
	ExpiresAt         time.Time
}

type ServiceAccessTicketClaims struct {
	UserID    string
	DeviceID  string
	SessionID string
	ServiceID string
	IssuedAt  time.Time
	ExpiresAt time.Time
}

// TokenIssuer 统一负责访问令牌和短期服务访问票据，避免散落在业务文件中。
type TokenIssuer struct {
	Secret []byte
	TTL    time.Duration
	Now    func() time.Time
}

func (i TokenIssuer) IssueAccessToken(claims AccessTokenClaims) (string, time.Time, error) {
	if len(i.Secret) == 0 {
		return "", time.Time{}, errors.New("token secret is required")
	}

	now := i.now().UTC()
	expiresAt := now.Add(i.ttl())

	payload := accessTokenPayload{
		Type:              "access",
		Subject:           claims.UserID,
		UserID:            claims.UserID,
		DeviceID:          claims.DeviceID,
		SessionID:         claims.SessionID,
		PermissionVersion: claims.PermissionVersion,
		IssuedAt:          now.Unix(),
		ExpiresAt:         expiresAt.Unix(),
	}

	token, err := signToken(i.Secret, payload)
	if err != nil {
		return "", time.Time{}, err
	}

	return token, expiresAt, nil
}

func (i TokenIssuer) VerifyAccessToken(token string) (AccessTokenClaims, error) {
	if len(i.Secret) == 0 {
		return AccessTokenClaims{}, errors.New("token secret is required")
	}

	payload, err := verifyToken(i.Secret, token)
	if err != nil {
		return AccessTokenClaims{}, err
	}
	if payload.Type != "access" {
		return AccessTokenClaims{}, ErrInvalidToken
	}

	now := i.now().UTC()
	expiresAt := time.Unix(payload.ExpiresAt, 0).UTC()
	if now.After(expiresAt) {
		return AccessTokenClaims{}, ErrExpiredToken
	}

	return AccessTokenClaims{
		UserID:            payload.UserID,
		DeviceID:          payload.DeviceID,
		SessionID:         payload.SessionID,
		PermissionVersion: payload.PermissionVersion,
		IssuedAt:          time.Unix(payload.IssuedAt, 0).UTC(),
		ExpiresAt:         expiresAt,
	}, nil
}

func (i TokenIssuer) IssueServiceAccessTicket(claims ServiceAccessTicketClaims) (string, time.Time, error) {
	if len(i.Secret) == 0 {
		return "", time.Time{}, errors.New("token secret is required")
	}

	now := i.now().UTC()
	expiresAt := now.Add(i.ttl())

	payload := accessTokenPayload{
		Type:      "service_access",
		Subject:   claims.UserID,
		UserID:    claims.UserID,
		DeviceID:  claims.DeviceID,
		SessionID: claims.SessionID,
		ServiceID: claims.ServiceID,
		IssuedAt:  now.Unix(),
		ExpiresAt: expiresAt.Unix(),
	}

	token, err := signToken(i.Secret, payload)
	if err != nil {
		return "", time.Time{}, err
	}

	return token, expiresAt, nil
}

func (i TokenIssuer) VerifyServiceAccessTicket(token string) (ServiceAccessTicketClaims, error) {
	if len(i.Secret) == 0 {
		return ServiceAccessTicketClaims{}, errors.New("token secret is required")
	}

	payload, err := verifyToken(i.Secret, token)
	if err != nil {
		return ServiceAccessTicketClaims{}, err
	}
	if payload.Type != "service_access" || payload.ServiceID == "" {
		return ServiceAccessTicketClaims{}, ErrInvalidToken
	}

	now := i.now().UTC()
	expiresAt := time.Unix(payload.ExpiresAt, 0).UTC()
	if now.After(expiresAt) {
		return ServiceAccessTicketClaims{}, ErrExpiredToken
	}

	return ServiceAccessTicketClaims{
		UserID:    payload.UserID,
		DeviceID:  payload.DeviceID,
		SessionID: payload.SessionID,
		ServiceID: payload.ServiceID,
		IssuedAt:  time.Unix(payload.IssuedAt, 0).UTC(),
		ExpiresAt: expiresAt,
	}, nil
}

func GenerateRefreshToken() (string, error) {
	random := make([]byte, refreshTokenLength)
	if _, err := rand.Read(random); err != nil {
		return "", fmt.Errorf("read refresh token bytes: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(random), nil
}

func GenerateChallenge() (string, error) {
	random := make([]byte, challengeLength)
	if _, err := rand.Read(random); err != nil {
		return "", fmt.Errorf("read challenge bytes: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(random), nil
}

func HashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

type accessTokenHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type accessTokenPayload struct {
	Type              string `json:"typ"`
	Subject           string `json:"sub"`
	UserID            string `json:"uid"`
	DeviceID          string `json:"did"`
	SessionID         string `json:"sid"`
	ServiceID         string `json:"svc,omitempty"`
	PermissionVersion int    `json:"pv"`
	IssuedAt          int64  `json:"iat"`
	ExpiresAt         int64  `json:"exp"`
}

func signToken(secret []byte, payload accessTokenPayload) (string, error) {
	headerJSON, err := json.Marshal(accessTokenHeader{
		Alg: "HS256",
		Typ: "JWT",
	})
	if err != nil {
		return "", fmt.Errorf("marshal token header: %w", err)
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal token payload: %w", err)
	}

	headerPart := base64.RawURLEncoding.EncodeToString(headerJSON)
	payloadPart := base64.RawURLEncoding.EncodeToString(payloadJSON)
	message := headerPart + "." + payloadPart

	mac := hmac.New(sha256.New, secret)
	if _, err := mac.Write([]byte(message)); err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}

	signaturePart := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return message + "." + signaturePart, nil
}

func verifyToken(secret []byte, token string) (accessTokenPayload, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return accessTokenPayload{}, ErrInvalidToken
	}

	message := parts[0] + "." + parts[1]

	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return accessTokenPayload{}, ErrInvalidToken
	}

	mac := hmac.New(sha256.New, secret)
	if _, err := mac.Write([]byte(message)); err != nil {
		return accessTokenPayload{}, fmt.Errorf("sign token: %w", err)
	}
	if !hmac.Equal(signature, mac.Sum(nil)) {
		return accessTokenPayload{}, ErrInvalidToken
	}

	var header accessTokenHeader
	if err := decodeTokenPart(parts[0], &header); err != nil {
		return accessTokenPayload{}, ErrInvalidToken
	}
	if header.Alg != "HS256" || header.Typ != "JWT" {
		return accessTokenPayload{}, ErrInvalidToken
	}

	var payload accessTokenPayload
	if err := decodeTokenPart(parts[1], &payload); err != nil {
		return accessTokenPayload{}, ErrInvalidToken
	}
	if payload.Subject == "" || payload.UserID == "" || payload.SessionID == "" {
		return accessTokenPayload{}, ErrInvalidToken
	}

	return payload, nil
}

func decodeTokenPart(part string, target any) error {
	decoded, err := base64.RawURLEncoding.DecodeString(part)
	if err != nil {
		return err
	}

	return json.Unmarshal(decoded, target)
}

func (i TokenIssuer) now() time.Time {
	if i.Now != nil {
		return i.Now()
	}
	return time.Now()
}

func (i TokenIssuer) ttl() time.Duration {
	if i.TTL > 0 {
		return i.TTL
	}
	return 15 * time.Minute
}
