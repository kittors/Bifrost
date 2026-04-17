package auth

import (
	"context"
	"crypto/ed25519"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/kittors/bifrost/apps/gateway/internal/contracts"
)

// 设备注册与挑战流程单独拆分，方便后续继续演进设备信任策略。
func (s Service) RegisterDevice(ctx context.Context, input RegisterDeviceInput) (DeviceResult, error) {
	claims, err := s.tokenIssuer().VerifyAccessToken(input.AccessToken)
	if err != nil {
		return DeviceResult{}, mapTokenError(err)
	}

	if input.Name == "" || input.OS == "" || input.ClientVersion == "" || input.PublicKey == "" || input.PublicKeyFingerprint == "" {
		return DeviceResult{}, &ServiceError{
			StatusCode:  http.StatusBadRequest,
			Code:        contracts.ErrorCodeCommonBadRequest,
			Message:     "device registration payload is incomplete",
			UserMessage: "请求参数不正确",
		}
	}

	if _, err := decodeEd25519PublicKey(input.PublicKey); err != nil {
		return DeviceResult{}, &ServiceError{
			StatusCode:  http.StatusUnauthorized,
			Code:        contracts.ErrorCodeDeviceKeyInvalid,
			Message:     "device public key is invalid",
			UserMessage: "设备身份校验失败",
		}
	}

	deviceID, err := s.newDeviceID()
	if err != nil {
		return DeviceResult{}, err
	}

	if _, err := s.db().ExecContext(
		ctx,
		`INSERT INTO devices (id, user_id, name, os, client_version, public_key, public_key_fingerprint, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'trusted')`,
		deviceID,
		claims.UserID,
		input.Name,
		input.OS,
		input.ClientVersion,
		input.PublicKey,
		input.PublicKeyFingerprint,
	); err != nil {
		if isUniqueViolation(err) {
			return DeviceResult{}, &ServiceError{
				StatusCode:  http.StatusConflict,
				Code:        contracts.ErrorCodeDeviceAlreadyBound,
				Message:     "device public key fingerprint is already bound",
				UserMessage: "设备已绑定",
			}
		}
		return DeviceResult{}, fmt.Errorf("insert device: %w", err)
	}

	return DeviceResult{ID: deviceID, Status: "trusted"}, nil
}

func (s Service) CreateDeviceChallenge(ctx context.Context, input CreateDeviceChallengeInput) (DeviceChallengeResult, error) {
	claims, err := s.tokenIssuer().VerifyAccessToken(input.AccessToken)
	if err != nil {
		return DeviceChallengeResult{}, mapTokenError(err)
	}

	if err := s.ensureTrustedDevice(ctx, claims.UserID, input.DeviceID, ""); err != nil {
		return DeviceChallengeResult{}, err
	}

	challengeID, err := s.newChallengeID()
	if err != nil {
		return DeviceChallengeResult{}, err
	}

	challenge, err := GenerateChallenge()
	if err != nil {
		return DeviceChallengeResult{}, err
	}

	ttl := s.challengeTTL()
	if _, err := s.db().ExecContext(
		ctx,
		`INSERT INTO device_challenges (id, device_id, challenge, expires_at)
		VALUES ($1, $2, $3, $4)`,
		challengeID,
		input.DeviceID,
		challenge,
		s.now().UTC().Add(ttl),
	); err != nil {
		return DeviceChallengeResult{}, fmt.Errorf("insert device challenge: %w", err)
	}

	return DeviceChallengeResult{
		ID:        challengeID,
		Challenge: challenge,
		ExpiresIn: int(ttl.Seconds()),
	}, nil
}

func (s Service) VerifyDeviceChallenge(ctx context.Context, input VerifyDeviceChallengeInput) (DeviceChallengeVerificationResult, error) {
	claims, err := s.tokenIssuer().VerifyAccessToken(input.AccessToken)
	if err != nil {
		return DeviceChallengeVerificationResult{}, mapTokenError(err)
	}

	challenge, err := s.loadDeviceChallenge(ctx, input.ChallengeID)
	if err != nil {
		return DeviceChallengeVerificationResult{}, err
	}

	if s.now().UTC().After(challenge.ExpiresAt.UTC()) {
		return DeviceChallengeVerificationResult{}, &ServiceError{
			StatusCode:  http.StatusUnauthorized,
			Code:        contracts.ErrorCodeDeviceChallengeExpired,
			Message:     "device challenge is expired",
			UserMessage: "设备验证已过期，请重试",
		}
	}

	device, err := s.loadDeviceKey(ctx, claims.UserID, challenge.DeviceID)
	if err != nil {
		return DeviceChallengeVerificationResult{}, err
	}

	publicKey, err := decodeEd25519PublicKey(device.PublicKey)
	if err != nil {
		return DeviceChallengeVerificationResult{}, &ServiceError{
			StatusCode:  http.StatusUnauthorized,
			Code:        contracts.ErrorCodeDeviceKeyInvalid,
			Message:     "device public key is invalid",
			UserMessage: "设备身份校验失败",
		}
	}

	rawChallenge, err := base64.RawURLEncoding.DecodeString(challenge.Challenge)
	if err != nil {
		return DeviceChallengeVerificationResult{}, fmt.Errorf("decode stored challenge: %w", err)
	}

	signature, err := base64.RawURLEncoding.DecodeString(input.Signature)
	if err != nil {
		return DeviceChallengeVerificationResult{}, &ServiceError{
			StatusCode:  http.StatusUnauthorized,
			Code:        contracts.ErrorCodeDeviceKeyInvalid,
			Message:     "device signature is invalid base64url",
			UserMessage: "设备身份校验失败",
		}
	}

	if !ed25519.Verify(publicKey, rawChallenge, signature) {
		return DeviceChallengeVerificationResult{}, &ServiceError{
			StatusCode:  http.StatusUnauthorized,
			Code:        contracts.ErrorCodeDeviceKeyInvalid,
			Message:     "device signature verification failed",
			UserMessage: "设备身份校验失败",
		}
	}

	if _, err := s.db().ExecContext(
		ctx,
		`UPDATE device_challenges
		SET verified_at = $2
		WHERE id = $1`,
		input.ChallengeID,
		s.now().UTC(),
	); err != nil {
		return DeviceChallengeVerificationResult{}, fmt.Errorf("mark device challenge verified: %w", err)
	}

	return DeviceChallengeVerificationResult{Verified: true}, nil
}

func (s Service) ensureTrustedDevice(ctx context.Context, userID string, deviceID string, clientVersion string) error {
	row := s.db().QueryRowContext(
		ctx,
		`SELECT status
		FROM devices
		WHERE id = $1 AND user_id = $2`,
		deviceID,
		userID,
	)

	var status string
	if err := row.Scan(&status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &ServiceError{
				StatusCode:  http.StatusForbidden,
				Code:        contracts.ErrorCodeDeviceNotTrusted,
				Message:     "device not found for user",
				UserMessage: "当前设备未被信任",
			}
		}
		return fmt.Errorf("query device: %w", err)
	}

	if status != "trusted" {
		return &ServiceError{
			StatusCode:  http.StatusForbidden,
			Code:        contracts.ErrorCodeDeviceDisabled,
			Message:     "device is disabled",
			UserMessage: "当前设备已被禁用",
		}
	}

	if _, err := s.db().ExecContext(
		ctx,
		`UPDATE devices
		SET last_seen_at = $3, client_version = CASE WHEN $4 <> '' THEN $4 ELSE client_version END, updated_at = $3
		WHERE id = $1 AND user_id = $2`,
		deviceID,
		userID,
		s.now().UTC(),
		clientVersion,
	); err != nil {
		return fmt.Errorf("update device last seen: %w", err)
	}

	return nil
}

type deviceChallengeRecord struct {
	ID        string
	DeviceID  string
	Challenge string
	ExpiresAt time.Time
}

type deviceKeyRecord struct {
	ID        string
	PublicKey string
}

func (s Service) loadDeviceChallenge(ctx context.Context, challengeID string) (deviceChallengeRecord, error) {
	row := s.db().QueryRowContext(
		ctx,
		`SELECT id, device_id, challenge, expires_at
		FROM device_challenges
		WHERE id = $1 AND verified_at IS NULL`,
		challengeID,
	)

	var record deviceChallengeRecord
	if err := row.Scan(&record.ID, &record.DeviceID, &record.Challenge, &record.ExpiresAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return deviceChallengeRecord{}, &ServiceError{
				StatusCode:  http.StatusNotFound,
				Code:        contracts.ErrorCodeDeviceNotFound,
				Message:     "device challenge not found",
				UserMessage: "设备不存在",
			}
		}
		return deviceChallengeRecord{}, fmt.Errorf("query device challenge: %w", err)
	}

	return record, nil
}

func (s Service) loadDeviceKey(ctx context.Context, userID string, deviceID string) (deviceKeyRecord, error) {
	row := s.db().QueryRowContext(
		ctx,
		`SELECT id, public_key
		FROM devices
		WHERE id = $1 AND user_id = $2 AND status = 'trusted'`,
		deviceID,
		userID,
	)

	var record deviceKeyRecord
	if err := row.Scan(&record.ID, &record.PublicKey); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return deviceKeyRecord{}, &ServiceError{
				StatusCode:  http.StatusForbidden,
				Code:        contracts.ErrorCodeDeviceNotTrusted,
				Message:     "device not found for user",
				UserMessage: "当前设备未被信任",
			}
		}
		return deviceKeyRecord{}, fmt.Errorf("query device public key: %w", err)
	}

	return record, nil
}

func (s Service) challengeTTL() time.Duration {
	if s.ChallengeTTL > 0 {
		return s.ChallengeTTL
	}
	return 2 * time.Minute
}

func decodeEd25519PublicKey(encoded string) (ed25519.PublicKey, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	if len(decoded) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("expected ed25519 public key length %d, got %d", ed25519.PublicKeySize, len(decoded))
	}
	return ed25519.PublicKey(decoded), nil
}
