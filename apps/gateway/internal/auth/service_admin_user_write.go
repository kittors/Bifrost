package auth

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/kittors/bifrost/apps/gateway/internal/contracts"
)

// 后台用户写入逻辑只处理创建与资料编辑，安全动作放在独立文件里。
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
	if err := ensureUserMutationAffected(result); err != nil {
		return AdminUser{}, err
	}

	if err := replaceUserRoles(ctx, tx, input.UserID, input.RoleIDs); err != nil {
		return AdminUser{}, err
	}

	if err := tx.Commit(); err != nil {
		return AdminUser{}, fmt.Errorf("commit update user transaction: %w", err)
	}

	return s.loadAdminUser(ctx, input.UserID)
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
