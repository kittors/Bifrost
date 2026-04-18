package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/kittors/bifrost/apps/gateway/internal/contracts"
)

// 后台用户查询逻辑只负责读取与过滤，不掺杂写入和安全动作。
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

func (s Service) GetAdminUser(ctx context.Context, input GetAdminUserInput) (AdminUser, error) {
	if _, err := s.ensureAdminPrincipal(ctx, input.AccessToken); err != nil {
		return AdminUser{}, err
	}

	return s.loadAdminUser(ctx, input.UserID)
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
