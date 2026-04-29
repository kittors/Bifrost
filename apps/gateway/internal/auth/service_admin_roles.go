package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/kittors/bifrost/apps/gateway/internal/contracts"
)

// 角色目录相关读写集中到单独文件，避免角色逻辑和服务、设备、审计互相缠绕。
func (s Service) ListAdminRoles(ctx context.Context, input ListAdminRolesInput) (AdminRoleListResult, error) {
	if _, err := s.ensureAdminPrincipal(ctx, input.AccessToken); err != nil {
		return AdminRoleListResult{}, err
	}
	page, pageSize := normalizePagination(input.Page, input.PageSize)
	where, args := buildSimpleKeywordFilter("r", []string{"name", "display_name", "description"}, input.Keyword, "")

	var total int64
	if err := s.db().QueryRowContext(ctx, "SELECT COUNT(*) FROM roles r "+where, args...).Scan(&total); err != nil {
		return AdminRoleListResult{}, fmt.Errorf("count roles: %w", err)
	}

	queryArgs := append([]any{}, args...)
	queryArgs = append(queryArgs, pageSize, (page-1)*pageSize)
	query := `SELECT id, name, display_name, description FROM roles r ` + where + fmt.Sprintf(" ORDER BY name ASC LIMIT $%d OFFSET $%d", len(queryArgs)-1, len(queryArgs))
	rows, err := s.db().QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return AdminRoleListResult{}, fmt.Errorf("query roles: %w", err)
	}
	defer rows.Close()

	items := []AdminRole{}
	for rows.Next() {
		var role AdminRole
		if err := rows.Scan(&role.ID, &role.Name, &role.DisplayName, &role.Description); err != nil {
			return AdminRoleListResult{}, fmt.Errorf("scan role: %w", err)
		}
		items = append(items, role)
	}
	if err := rows.Err(); err != nil {
		return AdminRoleListResult{}, fmt.Errorf("iterate roles: %w", err)
	}

	return AdminRoleListResult{
		Items: items,
		Pagination: contracts.Pagination{
			Page:       int64(page),
			PageSize:   int64(pageSize),
			Total:      total,
			TotalPages: totalPages(total, pageSize),
		},
	}, nil
}

func (s Service) CreateAdminRole(ctx context.Context, input CreateAdminRoleInput) (AdminRole, error) {
	if _, err := s.ensureAdminPrincipal(ctx, input.AccessToken); err != nil {
		return AdminRole{}, err
	}

	roleID, err := s.newRoleID()
	if err != nil {
		return AdminRole{}, err
	}

	if _, err := s.db().ExecContext(
		ctx,
		`INSERT INTO roles (id, name, display_name, description, is_system) VALUES ($1, $2, $3, $4, false)`,
		roleID,
		input.Name,
		input.DisplayName,
		input.Description,
	); err != nil {
		if isUniqueViolation(err) {
			return AdminRole{}, &ServiceError{
				StatusCode:  409,
				Code:        contracts.ErrorCodeRoleAlreadyExists,
				Message:     "role already exists",
				UserMessage: "角色已存在",
			}
		}
		return AdminRole{}, fmt.Errorf("insert role: %w", err)
	}

	return AdminRole{
		ID:          roleID,
		Name:        input.Name,
		DisplayName: input.DisplayName,
		Description: input.Description,
	}, nil
}

func (s Service) UpdateAdminRole(ctx context.Context, input UpdateAdminRoleInput) (AdminRole, error) {
	if _, err := s.ensureAdminPrincipal(ctx, input.AccessToken); err != nil {
		return AdminRole{}, err
	}

	result, err := s.db().ExecContext(
		ctx,
		`UPDATE roles
		SET display_name = $2, description = $3, updated_at = $4
		WHERE id = $1`,
		input.RoleID,
		input.DisplayName,
		input.Description,
		s.now().UTC(),
	)
	if err != nil {
		return AdminRole{}, fmt.Errorf("update admin role: %w", err)
	}
	if err := ensureRoleMutationAffected(result); err != nil {
		return AdminRole{}, err
	}

	row := s.db().QueryRowContext(
		ctx,
		`SELECT id, name, display_name, description FROM roles WHERE id = $1`,
		input.RoleID,
	)

	var role AdminRole
	if err := row.Scan(&role.ID, &role.Name, &role.DisplayName, &role.Description); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return AdminRole{}, &ServiceError{
				StatusCode:  http.StatusNotFound,
				Code:        contracts.ErrorCodeRoleNotFound,
				Message:     "role not found",
				UserMessage: "角色不存在",
			}
		}
		return AdminRole{}, fmt.Errorf("query updated role: %w", err)
	}

	return role, nil
}

func (s Service) ReplaceRoleServices(ctx context.Context, input ReplaceRoleServicesInput) error {
	if _, err := s.ensureAdminPrincipal(ctx, input.AccessToken); err != nil {
		return err
	}

	tx, err := s.db().BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin role services transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM role_services WHERE role_id = $1`, input.RoleID); err != nil {
		return fmt.Errorf("delete role services: %w", err)
	}
	for _, serviceID := range input.ServiceIDs {
		if _, err := tx.ExecContext(ctx, `INSERT INTO role_services (role_id, service_id) VALUES ($1, $2)`, input.RoleID, serviceID); err != nil {
			return fmt.Errorf("insert role service: %w", err)
		}
	}
	return tx.Commit()
}

func ensureRoleMutationAffected(result sql.Result) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("admin role mutation rows affected: %w", err)
	}
	if affected == 0 {
		return &ServiceError{
			StatusCode:  http.StatusNotFound,
			Code:        contracts.ErrorCodeRoleNotFound,
			Message:     "role not found",
			UserMessage: "角色不存在",
		}
	}
	return nil
}
