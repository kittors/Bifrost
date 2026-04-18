package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/kittors/bifrost/apps/gateway/internal/contracts"
)

// 认证入口仅保留登录、刷新、登出和凭证校验，不再混入会话持久化细节。
func (s Service) AdminLogin(ctx context.Context, input AdminLoginInput) (LoginResult, error) {
	user, err := s.authenticateUser(ctx, input.Username, input.Password)
	if err != nil {
		if auditErr := s.recordLoginFailedEvent(ctx, input.RequestID, input.Username); auditErr != nil {
			return LoginResult{}, auditErr
		}
		return LoginResult{}, err
	}

	if !slices.Contains(user.RoleIDs, adminRoleID) {
		return LoginResult{}, &ServiceError{
			StatusCode:  http.StatusForbidden,
			Code:        contracts.ErrorCodePolicyAccessDenied,
			Message:     "admin role is required",
			UserMessage: "当前账号没有后台访问权限",
		}
	}

	result, err := s.createSession(ctx, user, nil)
	if err != nil {
		return LoginResult{}, err
	}

	if err := s.recordAuditEvent(ctx, auditEventInput{
		RequestID:   input.RequestID,
		Type:        contracts.AuditEventTypeAuthLoginSucceeded,
		ActorUserID: user.ID,
		TargetType:  "user",
		TargetID:    user.ID,
		Result:      "success",
		Summary:     "admin login succeeded",
	}); err != nil {
		return LoginResult{}, err
	}

	return result, nil
}

func (s Service) ClientLogin(ctx context.Context, input ClientLoginInput) (LoginResult, error) {
	user, err := s.authenticateUser(ctx, input.Username, input.Password)
	if err != nil {
		if auditErr := s.recordLoginFailedEvent(ctx, input.RequestID, input.Username); auditErr != nil {
			return LoginResult{}, auditErr
		}
		return LoginResult{}, err
	}

	if input.DeviceID == "" {
		return LoginResult{}, &ServiceError{
			StatusCode:  http.StatusForbidden,
			Code:        contracts.ErrorCodeDeviceNotTrusted,
			Message:     "device id is required",
			UserMessage: "当前设备未被信任",
		}
	}

	if err := s.ensureTrustedDevice(ctx, user.ID, input.DeviceID, input.ClientVersion); err != nil {
		return LoginResult{}, err
	}

	return s.createSession(ctx, user, &input.DeviceID)
}

func (s Service) RefreshSession(ctx context.Context, input RefreshInput) (LoginResult, error) {
	if input.RefreshToken == "" {
		return LoginResult{}, &ServiceError{
			StatusCode:  http.StatusUnauthorized,
			Code:        contracts.ErrorCodeAuthRefreshTokenInvalid,
			Message:     "refresh token is required",
			UserMessage: "登录状态已失效，请重新登录",
		}
	}

	record, err := s.loadSessionByRefreshToken(ctx, input.RefreshToken)
	if err != nil {
		return LoginResult{}, err
	}

	if record.DeviceID.Valid && input.DeviceID != "" && record.DeviceID.String != input.DeviceID {
		return LoginResult{}, &ServiceError{
			StatusCode:  http.StatusUnauthorized,
			Code:        contracts.ErrorCodeAuthRefreshTokenInvalid,
			Message:     "refresh token device mismatch",
			UserMessage: "登录状态已失效，请重新登录",
		}
	}

	user, err := s.loadUser(ctx, record.UserID)
	if err != nil {
		return LoginResult{}, err
	}

	if record.DeviceID.Valid {
		if input.DeviceID == "" {
			return LoginResult{}, &ServiceError{
				StatusCode:  http.StatusUnauthorized,
				Code:        contracts.ErrorCodeAuthRefreshTokenInvalid,
				Message:     "refresh token device id is required",
				UserMessage: "登录状态已失效，请重新登录",
			}
		}

		if err := s.ensureTrustedDevice(ctx, user.ID, record.DeviceID.String, ""); err != nil {
			return LoginResult{}, err
		}
	}

	return s.rotateSession(ctx, user, record)
}

func (s Service) Logout(ctx context.Context, input LogoutInput) error {
	claims, err := s.tokenIssuer().VerifyAccessToken(input.AccessToken)
	if err != nil {
		return mapTokenError(err)
	}

	result, err := s.db().ExecContext(
		ctx,
		`UPDATE sessions
		SET status = 'revoked', revoked_at = $2
		WHERE id = $1 AND status = 'active'`,
		claims.SessionID,
		s.now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("revoke session rows affected: %w", err)
	}
	if affected == 0 {
		return &ServiceError{
			StatusCode:  http.StatusUnauthorized,
			Code:        contracts.ErrorCodeAuthSessionRevoked,
			Message:     "session is not active",
			UserMessage: "当前会话已被管理员终止",
		}
	}

	return nil
}

func (s Service) CurrentUser(ctx context.Context, input CurrentUserInput) (LoginUser, error) {
	claims, err := s.tokenIssuer().VerifyAccessToken(input.AccessToken)
	if err != nil {
		return LoginUser{}, mapTokenError(err)
	}

	record, err := s.loadSessionByID(ctx, claims.SessionID)
	if err != nil {
		return LoginUser{}, err
	}
	if record.Status != "active" {
		return LoginUser{}, &ServiceError{
			StatusCode:  http.StatusUnauthorized,
			Code:        contracts.ErrorCodeAuthSessionRevoked,
			Message:     "session is revoked",
			UserMessage: "当前会话已被管理员终止",
		}
	}

	if s.now().UTC().After(record.ExpiresAt.UTC()) {
		return LoginUser{}, &ServiceError{
			StatusCode:  http.StatusUnauthorized,
			Code:        contracts.ErrorCodeAuthSessionExpired,
			Message:     "session is expired",
			UserMessage: "登录已过期，请重新登录",
		}
	}

	user, err := s.loadUser(ctx, claims.UserID)
	if err != nil {
		return LoginUser{}, err
	}

	return LoginUser{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Roles:       append([]string(nil), user.RoleIDs...),
	}, nil
}

func (s Service) authenticateUser(ctx context.Context, username string, password string) (userRecord, error) {
	if username == "" || password == "" {
		return userRecord{}, &ServiceError{
			StatusCode:  http.StatusUnauthorized,
			Code:        contracts.ErrorCodeAuthInvalidCredentials,
			Message:     "username and password are required",
			UserMessage: "账号或密码不正确",
		}
	}

	row := s.db().QueryRowContext(
		ctx,
		`SELECT id, username, display_name, password_hash, status
		FROM users
		WHERE username = $1 AND deleted_at IS NULL`,
		username,
	)

	var user userRecord
	if err := row.Scan(&user.ID, &user.Username, &user.DisplayName, &user.PasswordHash, &user.Status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return userRecord{}, invalidCredentialsError()
		}
		return userRecord{}, fmt.Errorf("query user: %w", err)
	}

	if user.Status != "enabled" {
		return userRecord{}, &ServiceError{
			StatusCode:  http.StatusForbidden,
			Code:        contracts.ErrorCodeUserDisabled,
			Message:     "user is disabled",
			UserMessage: "账号已被禁用",
		}
	}

	ok, err := s.passwordHasher().Verify(user.PasswordHash, password)
	if err != nil {
		return userRecord{}, fmt.Errorf("verify password: %w", err)
	}
	if !ok {
		return userRecord{}, invalidCredentialsError()
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

func invalidCredentialsError() error {
	return &ServiceError{
		StatusCode:  http.StatusUnauthorized,
		Code:        contracts.ErrorCodeAuthInvalidCredentials,
		Message:     "invalid credentials",
		UserMessage: "账号或密码不正确",
	}
}

func (s Service) recordLoginFailedEvent(ctx context.Context, requestID string, username string) error {
	targetID := strings.TrimSpace(username)

	return s.recordAuditEvent(ctx, auditEventInput{
		RequestID:  requestID,
		Type:       contracts.AuditEventTypeAuthLoginFailed,
		TargetType: "user",
		TargetID:   targetID,
		Result:     "failure",
		Summary:    "login failed",
	})
}
