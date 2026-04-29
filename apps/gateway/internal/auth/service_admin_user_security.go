package auth

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/kittors/bifrost/apps/gateway/internal/contracts"
)

// 后台用户安全动作集中在这里，确保禁用、重置密码和管理员保护规则一致。
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
