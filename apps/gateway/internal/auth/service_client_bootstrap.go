package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kittors/bifrost/apps/gateway/internal/contracts"
)

// 客户端首登设备 bootstrap 专门处理“账号密码正确但本机尚未绑定设备”的初次闭环。
func (s Service) BootstrapClientDevice(ctx context.Context, input BootstrapClientDeviceInput) (ClientBootstrapResult, error) {
	user, err := s.authenticateUser(ctx, input.Username, input.Password)
	if err != nil {
		return ClientBootstrapResult{}, err
	}

	if input.DeviceName == "" || input.DeviceOS == "" || input.ClientVersion == "" || input.PublicKey == "" || input.PublicKeyFingerprint == "" {
		return ClientBootstrapResult{}, &ServiceError{
			StatusCode:  http.StatusBadRequest,
			Code:        contracts.ErrorCodeCommonBadRequest,
			Message:     "device bootstrap payload is incomplete",
			UserMessage: "请求参数不正确",
		}
	}

	if _, err := decodeEd25519PublicKey(input.PublicKey); err != nil {
		return ClientBootstrapResult{}, &ServiceError{
			StatusCode:  http.StatusUnauthorized,
			Code:        contracts.ErrorCodeDeviceKeyInvalid,
			Message:     "device public key is invalid",
			UserMessage: "设备身份校验失败",
		}
	}

	deviceID, err := s.newDeviceID()
	if err != nil {
		return ClientBootstrapResult{}, err
	}

	if _, err := s.db().ExecContext(
		ctx,
		`INSERT INTO devices (id, user_id, name, os, client_version, public_key, public_key_fingerprint, status, last_seen_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'trusted', $8)`,
		deviceID,
		user.ID,
		input.DeviceName,
		input.DeviceOS,
		input.ClientVersion,
		input.PublicKey,
		input.PublicKeyFingerprint,
		s.now().UTC(),
	); err != nil {
		if isUniqueViolation(err) {
			return ClientBootstrapResult{}, &ServiceError{
				StatusCode:  http.StatusConflict,
				Code:        contracts.ErrorCodeDeviceAlreadyBound,
				Message:     "device public key fingerprint is already bound",
				UserMessage: "设备已绑定",
			}
		}
		return ClientBootstrapResult{}, fmt.Errorf("insert bootstrap device: %w", err)
	}

	session, err := s.createSession(ctx, user, &deviceID)
	if err != nil {
		return ClientBootstrapResult{}, err
	}

	if err := s.recordAuditEvent(ctx, auditEventInput{
		RequestID:   input.RequestID,
		Type:        contracts.AuditEventTypeDeviceRegistered,
		ActorUserID: user.ID,
		TargetType:  "device",
		TargetID:    deviceID,
		Result:      "success",
		Summary:     "client device bootstrapped",
	}); err != nil {
		return ClientBootstrapResult{}, err
	}

	if err := s.recordAuditEvent(ctx, auditEventInput{
		RequestID:   input.RequestID,
		Type:        contracts.AuditEventTypeAuthLoginSucceeded,
		ActorUserID: user.ID,
		TargetType:  "user",
		TargetID:    user.ID,
		Result:      "success",
		Summary:     "client bootstrap login succeeded",
	}); err != nil {
		return ClientBootstrapResult{}, err
	}

	return ClientBootstrapResult{
		AccessToken:  session.AccessToken,
		RefreshToken: session.RefreshToken,
		ExpiresIn:    session.ExpiresIn,
		User:         session.User,
		Device:       DeviceResult{ID: deviceID, Status: "trusted"},
	}, nil
}
