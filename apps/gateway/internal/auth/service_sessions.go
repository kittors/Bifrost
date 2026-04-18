package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/kittors/bifrost/apps/gateway/internal/contracts"
)

// 会话、用户加载和 principal 解析独立收敛，避免认证入口文件继续膨胀。
func (s Service) createSession(ctx context.Context, user userRecord, deviceID *string) (LoginResult, error) {
	sessionID, err := s.newSessionID()
	if err != nil {
		return LoginResult{}, err
	}

	refreshToken, err := GenerateRefreshToken()
	if err != nil {
		return LoginResult{}, err
	}

	now := s.now().UTC()
	refreshExpiresAt := now.Add(s.refreshTokenTTL())
	accessToken, accessExpiresAt, err := s.tokenIssuer().IssueAccessToken(AccessTokenClaims{
		UserID:            user.ID,
		DeviceID:          valueOrEmpty(deviceID),
		SessionID:         sessionID,
		PermissionVersion: 1,
	})
	if err != nil {
		return LoginResult{}, fmt.Errorf("issue access token: %w", err)
	}

	var deviceValue any
	if deviceID == nil {
		deviceValue = nil
	} else {
		deviceValue = *deviceID
	}

	if _, err := s.db().ExecContext(
		ctx,
		`INSERT INTO sessions (id, user_id, device_id, refresh_token_hash, status, expires_at)
		VALUES ($1, $2, $3, $4, 'active', $5)`,
		sessionID,
		user.ID,
		deviceValue,
		HashRefreshToken(refreshToken),
		refreshExpiresAt,
	); err != nil {
		return LoginResult{}, fmt.Errorf("insert session: %w", err)
	}

	return LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(accessExpiresAt.Sub(now).Seconds()),
		User: LoginUser{
			ID:          user.ID,
			Username:    user.Username,
			DisplayName: user.DisplayName,
			Roles:       append([]string(nil), user.RoleIDs...),
		},
	}, nil
}

func (s Service) rotateSession(ctx context.Context, user userRecord, session sessionRecord) (LoginResult, error) {
	refreshToken, err := GenerateRefreshToken()
	if err != nil {
		return LoginResult{}, err
	}

	now := s.now().UTC()
	refreshExpiresAt := now.Add(s.refreshTokenTTL())
	accessToken, accessExpiresAt, err := s.tokenIssuer().IssueAccessToken(AccessTokenClaims{
		UserID:            user.ID,
		DeviceID:          session.DeviceID.String,
		SessionID:         session.ID,
		PermissionVersion: 1,
	})
	if err != nil {
		return LoginResult{}, fmt.Errorf("issue access token: %w", err)
	}

	if _, err := s.db().ExecContext(
		ctx,
		`UPDATE sessions
		SET refresh_token_hash = $2, expires_at = $3
		WHERE id = $1`,
		session.ID,
		HashRefreshToken(refreshToken),
		refreshExpiresAt,
	); err != nil {
		return LoginResult{}, fmt.Errorf("update session refresh token: %w", err)
	}

	return LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int(accessExpiresAt.Sub(now).Seconds()),
		User: LoginUser{
			ID:          user.ID,
			Username:    user.Username,
			DisplayName: user.DisplayName,
			Roles:       append([]string(nil), user.RoleIDs...),
		},
	}, nil
}

func (s Service) loadUser(ctx context.Context, userID string) (userRecord, error) {
	row := s.db().QueryRowContext(
		ctx,
		`SELECT id, username, display_name, password_hash, status
		FROM users
		WHERE id = $1 AND deleted_at IS NULL`,
		userID,
	)

	var user userRecord
	if err := row.Scan(&user.ID, &user.Username, &user.DisplayName, &user.PasswordHash, &user.Status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return userRecord{}, &ServiceError{
				StatusCode:  http.StatusNotFound,
				Code:        contracts.ErrorCodeUserNotFound,
				Message:     "user not found",
				UserMessage: "用户不存在",
			}
		}
		return userRecord{}, fmt.Errorf("query user by id: %w", err)
	}

	if user.Status != "enabled" {
		return userRecord{}, &ServiceError{
			StatusCode:  http.StatusForbidden,
			Code:        contracts.ErrorCodeUserDisabled,
			Message:     "user is disabled",
			UserMessage: "账号已被禁用",
		}
	}

	roleRows, err := s.db().QueryContext(
		ctx,
		`SELECT r.id
		FROM roles r
		INNER JOIN user_roles ur ON ur.role_id = r.id
		WHERE ur.user_id = $1
		ORDER BY r.id ASC`,
		user.ID,
	)
	if err != nil {
		return userRecord{}, fmt.Errorf("query user roles: %w", err)
	}
	defer roleRows.Close()

	for roleRows.Next() {
		var roleID string
		if err := roleRows.Scan(&roleID); err != nil {
			return userRecord{}, fmt.Errorf("scan user role: %w", err)
		}
		user.RoleIDs = append(user.RoleIDs, roleID)
	}

	if err := roleRows.Err(); err != nil {
		return userRecord{}, fmt.Errorf("iterate user roles: %w", err)
	}

	return user, nil
}

func (s Service) loadSessionByRefreshToken(ctx context.Context, refreshToken string) (sessionRecord, error) {
	row := s.db().QueryRowContext(
		ctx,
		`SELECT id, user_id, device_id, refresh_token_hash, status, expires_at
		FROM sessions
		WHERE refresh_token_hash = $1`,
		HashRefreshToken(refreshToken),
	)

	return s.scanSessionRecord(row)
}

func (s Service) loadSessionByID(ctx context.Context, sessionID string) (sessionRecord, error) {
	row := s.db().QueryRowContext(
		ctx,
		`SELECT id, user_id, device_id, refresh_token_hash, status, expires_at
		FROM sessions
		WHERE id = $1`,
		sessionID,
	)

	return s.scanSessionRecord(row)
}

func (s Service) scanSessionRecord(row *sql.Row) (sessionRecord, error) {
	var record sessionRecord
	if err := row.Scan(&record.ID, &record.UserID, &record.DeviceID, &record.RefreshTokenHash, &record.Status, &record.ExpiresAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return sessionRecord{}, &ServiceError{
				StatusCode:  http.StatusUnauthorized,
				Code:        contracts.ErrorCodeAuthRefreshTokenInvalid,
				Message:     "session not found",
				UserMessage: "登录状态已失效，请重新登录",
			}
		}
		return sessionRecord{}, fmt.Errorf("scan session: %w", err)
	}

	if record.Status == "revoked" {
		return sessionRecord{}, &ServiceError{
			StatusCode:  http.StatusUnauthorized,
			Code:        contracts.ErrorCodeAuthSessionRevoked,
			Message:     "session is revoked",
			UserMessage: "当前会话已被管理员终止",
		}
	}

	if s.now().UTC().After(record.ExpiresAt.UTC()) {
		return sessionRecord{}, &ServiceError{
			StatusCode:  http.StatusUnauthorized,
			Code:        contracts.ErrorCodeAuthSessionExpired,
			Message:     "session is expired",
			UserMessage: "登录已过期，请重新登录",
		}
	}

	return record, nil
}

func (s Service) loadClientPrincipal(ctx context.Context, accessToken string) (clientPrincipal, error) {
	claims, err := s.tokenIssuer().VerifyAccessToken(accessToken)
	if err != nil {
		return clientPrincipal{}, mapTokenError(err)
	}

	return s.loadClientPrincipalFromClaims(ctx, claims)
}

func (s Service) loadClientPrincipalFromClaims(ctx context.Context, claims AccessTokenClaims) (clientPrincipal, error) {
	session, err := s.loadSessionByID(ctx, claims.SessionID)
	if err != nil {
		return clientPrincipal{}, err
	}
	if session.Status != "active" {
		return clientPrincipal{}, &ServiceError{
			StatusCode:  http.StatusUnauthorized,
			Code:        contracts.ErrorCodeAuthSessionRevoked,
			Message:     "session is revoked",
			UserMessage: "当前会话已被管理员终止",
		}
	}

	if claims.DeviceID != "" {
		if err := s.ensureTrustedDevice(ctx, claims.UserID, claims.DeviceID, ""); err != nil {
			return clientPrincipal{}, err
		}
	}

	user, err := s.loadUser(ctx, claims.UserID)
	if err != nil {
		return clientPrincipal{}, err
	}

	return clientPrincipal{Claims: claims, User: user}, nil
}

func mapTokenError(err error) error {
	if errors.Is(err, ErrExpiredToken) {
		return &ServiceError{
			StatusCode:  http.StatusUnauthorized,
			Code:        contracts.ErrorCodeAuthInvalidToken,
			Message:     "access token is expired",
			UserMessage: "登录状态已失效，请重新登录",
		}
	}

	if errors.Is(err, ErrInvalidToken) {
		return &ServiceError{
			StatusCode:  http.StatusUnauthorized,
			Code:        contracts.ErrorCodeAuthInvalidToken,
			Message:     "access token is invalid",
			UserMessage: "登录状态已失效，请重新登录",
		}
	}

	return err
}
