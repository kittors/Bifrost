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

// 后台用户管理和管理员身份校验集中维护，减少权限与用户写操作的耦合扩散。
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
	principal, err := s.ensureAdminPrincipal(ctx, input.AccessToken)
	if err != nil {
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

	if err := s.recordAuditEvent(ctx, auditEventInput{
		RequestID:   input.RequestID,
		Type:        contracts.AuditEventTypeAdminUserCreated,
		ActorUserID: principal.User.ID,
		TargetType:  "user",
		TargetID:    userID,
		Result:      "success",
		Summary:     "admin user created",
	}); err != nil {
		return AdminUser{}, err
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

func (s Service) GetAdminUser(ctx context.Context, input GetAdminUserInput) (AdminUser, error) {
	if _, err := s.ensureAdminPrincipal(ctx, input.AccessToken); err != nil {
		return AdminUser{}, err
	}

	return s.loadAdminUser(ctx, input.UserID)
}

func (s Service) ResetAdminUserPassword(ctx context.Context, input ResetAdminUserPasswordInput) error {
	principal, err := s.ensureAdminPrincipal(ctx, input.AccessToken)
	if err != nil {
		return err
	}

	if strings.TrimSpace(input.Password) == "" {
		return &ServiceError{
			StatusCode:  http.StatusBadRequest,
			Code:        contracts.ErrorCodeCommonBadRequest,
			Message:     "password is required",
			UserMessage: "请求参数不正确",
		}
	}

	passwordHash, err := s.passwordHasher().Hash(input.Password)
	if err != nil {
		return fmt.Errorf("hash reset password: %w", err)
	}

	tx, err := s.db().BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin reset password transaction: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(
		ctx,
		`UPDATE users
		SET password_hash = $2, updated_at = $3
		WHERE id = $1 AND deleted_at IS NULL`,
		input.UserID,
		passwordHash,
		s.now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("update admin user password: %w", err)
	}
	if err := ensureUserMutationAffected(result); err != nil {
		return err
	}

	if _, err := tx.ExecContext(
		ctx,
		`UPDATE sessions
		SET status = 'revoked', revoked_at = $2
		WHERE user_id = $1 AND status = 'active'`,
		input.UserID,
		s.now().UTC(),
	); err != nil {
		return fmt.Errorf("revoke user sessions after password reset: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit reset password transaction: %w", err)
	}

	return s.recordAuditEvent(ctx, auditEventInput{
		RequestID:   input.RequestID,
		Type:        contracts.AuditEventTypeAdminUserUpdated,
		ActorUserID: principal.User.ID,
		TargetType:  "user",
		TargetID:    input.UserID,
		Result:      "success",
		Summary:     "admin user password reset",
	})
}

func (s Service) SetAdminUserStatus(ctx context.Context, input SetAdminUserStatusInput) (AdminUser, error) {
	principal, err := s.ensureAdminPrincipal(ctx, input.AccessToken)
	if err != nil {
		return AdminUser{}, err
	}

	status := strings.TrimSpace(input.Status)
	if status != "enabled" && status != "disabled" {
		return AdminUser{}, &ServiceError{
			StatusCode:  http.StatusBadRequest,
			Code:        contracts.ErrorCodeCommonBadRequest,
			Message:     "status must be enabled or disabled",
			UserMessage: "请求参数不正确",
		}
	}

	targetUser, err := s.loadAdminUser(ctx, input.UserID)
	if err != nil {
		return AdminUser{}, err
	}

	if status == "disabled" {
		if principal.User.ID == input.UserID {
			return AdminUser{}, &ServiceError{
				StatusCode:  http.StatusUnprocessableEntity,
				Code:        contracts.ErrorCodeUserCannotDisableSelf,
				Message:     "cannot disable current admin user",
				UserMessage: "不能禁用当前登录账号",
			}
		}

		if err := s.ensureLastEnabledAdmin(ctx, targetUser); err != nil {
			return AdminUser{}, err
		}
	}

	tx, err := s.db().BeginTx(ctx, nil)
	if err != nil {
		return AdminUser{}, fmt.Errorf("begin set user status transaction: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(
		ctx,
		`UPDATE users
		SET status = $2, updated_at = $3
		WHERE id = $1 AND deleted_at IS NULL`,
		input.UserID,
		status,
		s.now().UTC(),
	)
	if err != nil {
		return AdminUser{}, fmt.Errorf("update admin user status: %w", err)
	}
	if err := ensureUserMutationAffected(result); err != nil {
		return AdminUser{}, err
	}

	if status == "disabled" {
		if _, err := tx.ExecContext(
			ctx,
			`UPDATE sessions
			SET status = 'revoked', revoked_at = $2
			WHERE user_id = $1 AND status = 'active'`,
			input.UserID,
			s.now().UTC(),
		); err != nil {
			return AdminUser{}, fmt.Errorf("revoke user sessions after disable: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return AdminUser{}, fmt.Errorf("commit set user status transaction: %w", err)
	}

	if err := s.recordAuditEvent(ctx, auditEventInput{
		RequestID:   input.RequestID,
		Type:        contracts.AuditEventTypeAdminUserUpdated,
		ActorUserID: principal.User.ID,
		TargetType:  "user",
		TargetID:    input.UserID,
		Result:      "success",
		Summary:     "admin user status updated",
	}); err != nil {
		return AdminUser{}, err
	}

	return s.loadAdminUser(ctx, input.UserID)
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

func (s Service) ensureLastEnabledAdmin(ctx context.Context, targetUser AdminUser) error {
	if !slices.Contains(targetUser.Roles, adminRoleID) || targetUser.Status == "disabled" {
		return nil
	}

	var enabledAdminCount int
	if err := s.db().QueryRowContext(
		ctx,
		`SELECT COUNT(*)
		FROM users u
		INNER JOIN user_roles ur ON ur.user_id = u.id
		WHERE ur.role_id = $1 AND u.status = 'enabled' AND u.deleted_at IS NULL`,
		adminRoleID,
	).Scan(&enabledAdminCount); err != nil {
		return fmt.Errorf("count enabled admin users: %w", err)
	}

	if enabledAdminCount <= 1 {
		return &ServiceError{
			StatusCode:  http.StatusUnprocessableEntity,
			Code:        contracts.ErrorCodeUserLastAdminRequired,
			Message:     "last enabled admin user is required",
			UserMessage: "至少需要保留一个管理员账号",
		}
	}

	return nil
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

func ensureUserMutationAffected(result sql.Result) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("admin user mutation rows affected: %w", err)
	}
	if affected == 0 {
		return &ServiceError{
			StatusCode:  http.StatusNotFound,
			Code:        contracts.ErrorCodeUserNotFound,
			Message:     "user not found",
			UserMessage: "用户不存在",
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
