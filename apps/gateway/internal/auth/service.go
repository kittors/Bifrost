package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"time"

	"github.com/kittors/bifrost/apps/gateway/internal/contracts"
)

const (
	adminRoleID = "role_admin"
)

type Service struct {
	DB               *sql.DB
	PasswordHasher   PasswordHasher
	TokenIssuer      TokenIssuer
	Now              func() time.Time
	RefreshTokenTTL  time.Duration
	SessionIDFactory func() (string, error)
}

type AdminLoginInput struct {
	Username string
	Password string
}

type ClientLoginInput struct {
	Username      string
	Password      string
	DeviceID      string
	ClientVersion string
}

type RefreshInput struct {
	RefreshToken string
	DeviceID     string
}

type LogoutInput struct {
	AccessToken string
}

type CurrentUserInput struct {
	AccessToken string
}

type LoginResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
	User         LoginUser
}

type LoginUser struct {
	ID          string
	Username    string
	DisplayName string
	Roles       []string
}

type ServiceError struct {
	StatusCode  int
	Code        contracts.ErrorCode
	Message     string
	UserMessage string
}

func (e *ServiceError) Error() string {
	return e.Message
}

func (s Service) AdminLogin(ctx context.Context, input AdminLoginInput) (LoginResult, error) {
	user, err := s.authenticateUser(ctx, input.Username, input.Password)
	if err != nil {
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

	return s.createSession(ctx, user, nil)
}

func (s Service) ClientLogin(ctx context.Context, input ClientLoginInput) (LoginResult, error) {
	user, err := s.authenticateUser(ctx, input.Username, input.Password)
	if err != nil {
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

	refreshed, err := s.rotateSession(ctx, user, record)
	if err != nil {
		return LoginResult{}, err
	}

	return refreshed, nil
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

	db := s.db()
	row := db.QueryRowContext(
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

	roleRows, err := db.QueryContext(
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

func (s Service) ensureTrustedDevice(ctx context.Context, userID string, deviceID string, clientVersion string) error {
	db := s.db()
	row := db.QueryRowContext(
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

	if _, err := db.ExecContext(
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

func invalidCredentialsError() error {
	return &ServiceError{
		StatusCode:  http.StatusUnauthorized,
		Code:        contracts.ErrorCodeAuthInvalidCredentials,
		Message:     "invalid credentials",
		UserMessage: "账号或密码不正确",
	}
}

type userRecord struct {
	ID           string
	Username     string
	DisplayName  string
	PasswordHash string
	Status       string
	RoleIDs      []string
}

type sessionRecord struct {
	ID               string
	UserID           string
	DeviceID         sql.NullString
	RefreshTokenHash string
	Status           string
	ExpiresAt        time.Time
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
	db := s.db()
	row := db.QueryRowContext(
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

	roleRows, err := db.QueryContext(
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

func (s Service) db() *sql.DB {
	return s.DB
}

func (s Service) passwordHasher() PasswordHasher {
	if s.PasswordHasher == (PasswordHasher{}) {
		return DefaultPasswordHasher()
	}
	return s.PasswordHasher
}

func (s Service) tokenIssuer() TokenIssuer {
	issuer := s.TokenIssuer
	if issuer.Now == nil {
		issuer.Now = s.now
	}
	return issuer
}

func (s Service) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}

func (s Service) refreshTokenTTL() time.Duration {
	if s.RefreshTokenTTL > 0 {
		return s.RefreshTokenTTL
	}
	return 7 * 24 * time.Hour
}

func (s Service) newSessionID() (string, error) {
	if s.SessionIDFactory != nil {
		return s.SessionIDFactory()
	}

	token, err := GenerateRefreshToken()
	if err != nil {
		return "", fmt.Errorf("generate session id: %w", err)
	}
	return "sess_" + token[:20], nil
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
