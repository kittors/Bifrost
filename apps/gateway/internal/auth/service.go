package auth

import (
	"context"
	"crypto/ed25519"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/kittors/bifrost/apps/gateway/internal/contracts"
)

const (
	adminRoleID = "role_admin"
)

type Service struct {
	DB                 *sql.DB
	PasswordHasher     PasswordHasher
	TokenIssuer        TokenIssuer
	Now                func() time.Time
	RefreshTokenTTL    time.Duration
	SessionIDFactory   func() (string, error)
	DeviceIDFactory    func() (string, error)
	ChallengeIDFactory func() (string, error)
	UserIDFactory      func() (string, error)
	ChallengeTTL       time.Duration
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

type RegisterDeviceInput struct {
	AccessToken          string
	Name                 string
	OS                   string
	ClientVersion        string
	PublicKey            string
	PublicKeyFingerprint string
}

type CreateDeviceChallengeInput struct {
	AccessToken string
	DeviceID    string
}

type VerifyDeviceChallengeInput struct {
	AccessToken string
	ChallengeID string
	Signature   string
}

type ListClientServicesInput struct {
	AccessToken string
	Keyword     string
	Group       string
}

type GetClientServiceInput struct {
	AccessToken string
	ServiceID   string
}

type CreateServiceAccessURLInput struct {
	AccessToken string
	ServiceID   string
}

type ListAdminUsersInput struct {
	AccessToken string
	Page        int
	PageSize    int
	Keyword     string
	Status      string
	RoleID      string
}

type CreateAdminUserInput struct {
	AccessToken string
	Username    string
	DisplayName string
	Email       string
	Password    string
	RoleIDs     []string
}

type UpdateAdminUserInput struct {
	AccessToken string
	UserID      string
	DisplayName string
	Email       string
	RoleIDs     []string
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

type DeviceResult struct {
	ID     string
	Status string
}

type DeviceChallengeResult struct {
	ID        string
	Challenge string
	ExpiresIn int
}

type DeviceChallengeVerificationResult struct {
	Verified bool
}

type ClientService struct {
	ID           string
	Key          string
	Name         string
	Description  string
	Group        string
	Status       string
	PublicPath   string
	AccessSource string
}

type ServiceAccessURLResult struct {
	PublicPath string
	ExpiresIn  int
}

type AdminUser struct {
	ID          string
	Username    string
	DisplayName string
	Email       string
	Status      string
	Roles       []string
}

type AdminUserListResult struct {
	Items      []AdminUser
	Pagination contracts.Pagination
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

func (s Service) ListClientServices(ctx context.Context, input ListClientServicesInput) ([]ClientService, error) {
	principal, err := s.loadClientPrincipal(ctx, input.AccessToken)
	if err != nil {
		return nil, err
	}

	rows, err := s.db().QueryContext(
		ctx,
		`SELECT id, key, name, description, group_name, status, public_path
		FROM services
		WHERE status = 'enabled'
		ORDER BY name ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("query services: %w", err)
	}
	defer rows.Close()

	var services []ClientService
	for rows.Next() {
		var service ClientService
		if err := rows.Scan(&service.ID, &service.Key, &service.Name, &service.Description, &service.Group, &service.Status, &service.PublicPath); err != nil {
			return nil, fmt.Errorf("scan service: %w", err)
		}

		accessSource, allowed, err := s.resolveServiceAccess(ctx, principal.User.ID, principal.User.RoleIDs, service.ID)
		if err != nil {
			return nil, err
		}
		if !allowed {
			continue
		}

		if input.Keyword != "" && !strings.Contains(strings.ToLower(service.Name+" "+service.Key+" "+service.Description), strings.ToLower(input.Keyword)) {
			continue
		}

		if input.Group != "" && service.Group != input.Group {
			continue
		}

		service.AccessSource = accessSource
		services = append(services, service)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate services: %w", err)
	}

	return services, nil
}

func (s Service) GetClientService(ctx context.Context, input GetClientServiceInput) (ClientService, error) {
	principal, err := s.loadClientPrincipal(ctx, input.AccessToken)
	if err != nil {
		return ClientService{}, err
	}

	service, err := s.loadService(ctx, input.ServiceID)
	if err != nil {
		return ClientService{}, err
	}

	accessSource, allowed, err := s.resolveServiceAccess(ctx, principal.User.ID, principal.User.RoleIDs, service.ID)
	if err != nil {
		return ClientService{}, err
	}
	if !allowed {
		return ClientService{}, &ServiceError{
			StatusCode:  http.StatusForbidden,
			Code:        contracts.ErrorCodePolicyAccessDenied,
			Message:     "user is not allowed to access service",
			UserMessage: "你没有访问该服务的权限",
		}
	}

	service.AccessSource = accessSource
	return service, nil
}

func (s Service) CreateServiceAccessURL(ctx context.Context, input CreateServiceAccessURLInput) (ServiceAccessURLResult, error) {
	service, err := s.GetClientService(ctx, GetClientServiceInput{
		AccessToken: input.AccessToken,
		ServiceID:   input.ServiceID,
	})
	if err != nil {
		return ServiceAccessURLResult{}, err
	}

	return ServiceAccessURLResult{
		PublicPath: service.PublicPath,
		ExpiresIn:  300,
	}, nil
}

func (s Service) ListAdminUsers(ctx context.Context, input ListAdminUsersInput) (AdminUserListResult, error) {
	if _, err := s.ensureAdminPrincipal(ctx, input.AccessToken); err != nil {
		return AdminUserListResult{}, err
	}

	page, pageSize := normalizePagination(input.Page, input.PageSize)
	where, args := buildAdminUserFilters(input)

	var total int64
	countQuery := "SELECT COUNT(*) FROM users u " + where
	if err := s.db().QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return AdminUserListResult{}, fmt.Errorf("count admin users: %w", err)
	}

	queryArgs := append([]any{}, args...)
	queryArgs = append(queryArgs, pageSize, (page-1)*pageSize)
	query := `SELECT u.id, u.username, u.display_name, COALESCE(u.email, ''), u.status
		FROM users u ` + where + fmt.Sprintf(" ORDER BY u.created_at DESC LIMIT $%d OFFSET $%d", len(queryArgs)-1, len(queryArgs))

	rows, err := s.db().QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return AdminUserListResult{}, fmt.Errorf("query admin users: %w", err)
	}
	defer rows.Close()

	items := []AdminUser{}
	for rows.Next() {
		var user AdminUser
		if err := rows.Scan(&user.ID, &user.Username, &user.DisplayName, &user.Email, &user.Status); err != nil {
			return AdminUserListResult{}, fmt.Errorf("scan admin user: %w", err)
		}
		roles, err := s.loadUserRoleIDs(ctx, user.ID)
		if err != nil {
			return AdminUserListResult{}, err
		}
		user.Roles = roles
		items = append(items, user)
	}
	if err := rows.Err(); err != nil {
		return AdminUserListResult{}, fmt.Errorf("iterate admin users: %w", err)
	}

	return AdminUserListResult{
		Items: items,
		Pagination: contracts.Pagination{
			Page:       int64(page),
			PageSize:   int64(pageSize),
			Total:      total,
			TotalPages: totalPages(total, pageSize),
		},
	}, nil
}

func (s Service) CreateAdminUser(ctx context.Context, input CreateAdminUserInput) (AdminUser, error) {
	if _, err := s.ensureAdminPrincipal(ctx, input.AccessToken); err != nil {
		return AdminUser{}, err
	}

	if input.Username == "" || input.DisplayName == "" || input.Password == "" {
		return AdminUser{}, &ServiceError{
			StatusCode:  http.StatusBadRequest,
			Code:        contracts.ErrorCodeCommonBadRequest,
			Message:     "username, display name and password are required",
			UserMessage: "请求参数不正确",
		}
	}

	userID, err := s.newUserID()
	if err != nil {
		return AdminUser{}, err
	}

	passwordHash, err := s.passwordHasher().Hash(input.Password)
	if err != nil {
		return AdminUser{}, fmt.Errorf("hash new user password: %w", err)
	}

	tx, err := s.db().BeginTx(ctx, nil)
	if err != nil {
		return AdminUser{}, fmt.Errorf("begin create user transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO users (id, username, display_name, email, password_hash, status)
		VALUES ($1, $2, $3, $4, $5, 'enabled')`,
		userID,
		input.Username,
		input.DisplayName,
		input.Email,
		passwordHash,
	); err != nil {
		if isUniqueViolation(err) {
			return AdminUser{}, &ServiceError{
				StatusCode:  http.StatusConflict,
				Code:        contracts.ErrorCodeUserAlreadyExists,
				Message:     "user already exists",
				UserMessage: "用户已存在",
			}
		}
		return AdminUser{}, fmt.Errorf("insert admin user: %w", err)
	}

	if err := replaceUserRoles(ctx, tx, userID, input.RoleIDs); err != nil {
		return AdminUser{}, err
	}

	if err := tx.Commit(); err != nil {
		return AdminUser{}, fmt.Errorf("commit create user transaction: %w", err)
	}

	return s.loadAdminUser(ctx, userID)
}

func (s Service) UpdateAdminUser(ctx context.Context, input UpdateAdminUserInput) (AdminUser, error) {
	if _, err := s.ensureAdminPrincipal(ctx, input.AccessToken); err != nil {
		return AdminUser{}, err
	}

	tx, err := s.db().BeginTx(ctx, nil)
	if err != nil {
		return AdminUser{}, fmt.Errorf("begin update user transaction: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(
		ctx,
		`UPDATE users
		SET display_name = $2, email = $3, updated_at = $4
		WHERE id = $1 AND deleted_at IS NULL`,
		input.UserID,
		input.DisplayName,
		input.Email,
		s.now().UTC(),
	)
	if err != nil {
		return AdminUser{}, fmt.Errorf("update admin user: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return AdminUser{}, fmt.Errorf("update admin user rows affected: %w", err)
	}
	if affected == 0 {
		return AdminUser{}, &ServiceError{
			StatusCode:  http.StatusNotFound,
			Code:        contracts.ErrorCodeUserNotFound,
			Message:     "user not found",
			UserMessage: "用户不存在",
		}
	}

	if err := replaceUserRoles(ctx, tx, input.UserID, input.RoleIDs); err != nil {
		return AdminUser{}, err
	}

	if err := tx.Commit(); err != nil {
		return AdminUser{}, fmt.Errorf("commit update user transaction: %w", err)
	}

	return s.loadAdminUser(ctx, input.UserID)
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

type clientPrincipal struct {
	Claims AccessTokenClaims
	User   userRecord
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

func (s Service) loadClientPrincipal(ctx context.Context, accessToken string) (clientPrincipal, error) {
	claims, err := s.tokenIssuer().VerifyAccessToken(accessToken)
	if err != nil {
		return clientPrincipal{}, mapTokenError(err)
	}

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

func (s Service) loadService(ctx context.Context, serviceID string) (ClientService, error) {
	row := s.db().QueryRowContext(
		ctx,
		`SELECT id, key, name, description, group_name, status, public_path
		FROM services
		WHERE id = $1 AND status = 'enabled'`,
		serviceID,
	)

	var service ClientService
	if err := row.Scan(&service.ID, &service.Key, &service.Name, &service.Description, &service.Group, &service.Status, &service.PublicPath); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ClientService{}, &ServiceError{
				StatusCode:  http.StatusNotFound,
				Code:        contracts.ErrorCodeServiceNotFound,
				Message:     "service not found",
				UserMessage: "服务不存在",
			}
		}
		return ClientService{}, fmt.Errorf("query service: %w", err)
	}

	return service, nil
}

func (s Service) resolveServiceAccess(ctx context.Context, userID string, roleIDs []string, serviceID string) (string, bool, error) {
	row := s.db().QueryRowContext(
		ctx,
		`SELECT effect
		FROM user_service_overrides
		WHERE user_id = $1 AND service_id = $2`,
		userID,
		serviceID,
	)

	var effect string
	if err := row.Scan(&effect); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", false, fmt.Errorf("query user service override: %w", err)
	}
	if effect == "deny" {
		return "deny", false, nil
	}
	if effect == "allow" {
		return "user", true, nil
	}

	if len(roleIDs) == 0 {
		return "", false, nil
	}

	placeholders := make([]string, 0, len(roleIDs))
	args := make([]any, 0, len(roleIDs)+1)
	args = append(args, serviceID)
	for index, roleID := range roleIDs {
		placeholders = append(placeholders, fmt.Sprintf("$%d", index+2))
		args = append(args, roleID)
	}

	query := fmt.Sprintf(
		`SELECT EXISTS (
			SELECT 1
			FROM role_services
			WHERE service_id = $1 AND role_id IN (%s)
		)`,
		strings.Join(placeholders, ","),
	)

	var roleAllowed bool
	if err := s.db().QueryRowContext(ctx, query, args...).Scan(&roleAllowed); err != nil {
		return "", false, fmt.Errorf("query role service access: %w", err)
	}
	if roleAllowed {
		return "role", true, nil
	}

	return "", false, nil
}

func (s Service) ensureAdminPrincipal(ctx context.Context, accessToken string) (clientPrincipal, error) {
	principal, err := s.loadClientPrincipal(ctx, accessToken)
	if err != nil {
		return clientPrincipal{}, err
	}
	if !slices.Contains(principal.User.RoleIDs, adminRoleID) {
		return clientPrincipal{}, &ServiceError{
			StatusCode:  http.StatusForbidden,
			Code:        contracts.ErrorCodePolicyAccessDenied,
			Message:     "admin role is required",
			UserMessage: "当前账号没有后台访问权限",
		}
	}
	return principal, nil
}

func (s Service) loadAdminUser(ctx context.Context, userID string) (AdminUser, error) {
	row := s.db().QueryRowContext(
		ctx,
		`SELECT id, username, display_name, COALESCE(email, ''), status
		FROM users
		WHERE id = $1 AND deleted_at IS NULL`,
		userID,
	)

	var user AdminUser
	if err := row.Scan(&user.ID, &user.Username, &user.DisplayName, &user.Email, &user.Status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return AdminUser{}, &ServiceError{
				StatusCode:  http.StatusNotFound,
				Code:        contracts.ErrorCodeUserNotFound,
				Message:     "user not found",
				UserMessage: "用户不存在",
			}
		}
		return AdminUser{}, fmt.Errorf("query admin user: %w", err)
	}

	roles, err := s.loadUserRoleIDs(ctx, user.ID)
	if err != nil {
		return AdminUser{}, err
	}
	user.Roles = roles
	return user, nil
}

func (s Service) loadUserRoleIDs(ctx context.Context, userID string) ([]string, error) {
	rows, err := s.db().QueryContext(
		ctx,
		`SELECT role_id
		FROM user_roles
		WHERE user_id = $1
		ORDER BY role_id ASC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("query user role ids: %w", err)
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var roleID string
		if err := rows.Scan(&roleID); err != nil {
			return nil, fmt.Errorf("scan user role id: %w", err)
		}
		roles = append(roles, roleID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user role ids: %w", err)
	}
	return roles, nil
}

func replaceUserRoles(ctx context.Context, tx *sql.Tx, userID string, roleIDs []string) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM user_roles WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("delete user roles: %w", err)
	}

	for _, roleID := range roleIDs {
		if _, err := tx.ExecContext(ctx, `INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)`, userID, roleID); err != nil {
			return fmt.Errorf("insert user role %s: %w", roleID, err)
		}
	}

	return nil
}

func buildAdminUserFilters(input ListAdminUsersInput) (string, []any) {
	conditions := []string{"WHERE u.deleted_at IS NULL"}
	args := []any{}

	if input.Keyword != "" {
		args = append(args, "%"+strings.ToLower(input.Keyword)+"%")
		conditions = append(conditions, fmt.Sprintf("(LOWER(u.username) LIKE $%d OR LOWER(u.display_name) LIKE $%d OR LOWER(COALESCE(u.email, '')) LIKE $%d)", len(args), len(args), len(args)))
	}
	if input.Status != "" {
		args = append(args, input.Status)
		conditions = append(conditions, fmt.Sprintf("u.status = $%d", len(args)))
	}
	if input.RoleID != "" {
		args = append(args, input.RoleID)
		conditions = append(conditions, fmt.Sprintf("EXISTS (SELECT 1 FROM user_roles ur WHERE ur.user_id = u.id AND ur.role_id = $%d)", len(args)))
	}

	return strings.Join(conditions, " AND "), args
}

func normalizePagination(page int, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func totalPages(total int64, pageSize int) int64 {
	if total == 0 {
		return 0
	}
	return int64((total + int64(pageSize) - 1) / int64(pageSize))
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

func (s Service) newDeviceID() (string, error) {
	if s.DeviceIDFactory != nil {
		return s.DeviceIDFactory()
	}

	token, err := GenerateRefreshToken()
	if err != nil {
		return "", fmt.Errorf("generate device id: %w", err)
	}
	return "dev_" + token[:20], nil
}

func (s Service) newChallengeID() (string, error) {
	if s.ChallengeIDFactory != nil {
		return s.ChallengeIDFactory()
	}

	token, err := GenerateRefreshToken()
	if err != nil {
		return "", fmt.Errorf("generate challenge id: %w", err)
	}
	return "ch_" + token[:20], nil
}

func (s Service) newUserID() (string, error) {
	if s.UserIDFactory != nil {
		return s.UserIDFactory()
	}

	token, err := GenerateRefreshToken()
	if err != nil {
		return "", fmt.Errorf("generate user id: %w", err)
	}
	return "user_" + token[:20], nil
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

func isUniqueViolation(err error) bool {
	return strings.Contains(err.Error(), "SQLSTATE 23505")
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
