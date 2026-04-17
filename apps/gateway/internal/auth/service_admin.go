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

// 后台管理能力统一收口，方便继续按用户、角色、服务、审计逐步细化。
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
				StatusCode:  http.StatusConflict,
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

func (s Service) ListAdminServices(ctx context.Context, input ListAdminServicesInput) (AdminServiceListResult, error) {
	if _, err := s.ensureAdminPrincipal(ctx, input.AccessToken); err != nil {
		return AdminServiceListResult{}, err
	}
	page, pageSize := normalizePagination(input.Page, input.PageSize)
	where, args := buildAdminServiceFilters(input)

	var total int64
	if err := s.db().QueryRowContext(ctx, "SELECT COUNT(*) FROM services s "+where, args...).Scan(&total); err != nil {
		return AdminServiceListResult{}, fmt.Errorf("count services: %w", err)
	}

	queryArgs := append([]any{}, args...)
	queryArgs = append(queryArgs, pageSize, (page-1)*pageSize)
	query := `SELECT id, key, name, description, group_name, protocol, upstream_url, public_path, status FROM services s ` + where + fmt.Sprintf(" ORDER BY name ASC LIMIT $%d OFFSET $%d", len(queryArgs)-1, len(queryArgs))
	rows, err := s.db().QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return AdminServiceListResult{}, fmt.Errorf("query admin services: %w", err)
	}
	defer rows.Close()

	items := []AdminService{}
	for rows.Next() {
		var service AdminService
		if err := rows.Scan(&service.ID, &service.Key, &service.Name, &service.Description, &service.Group, &service.Protocol, &service.UpstreamURL, &service.PublicPath, &service.Status); err != nil {
			return AdminServiceListResult{}, fmt.Errorf("scan admin service: %w", err)
		}
		items = append(items, service)
	}
	if err := rows.Err(); err != nil {
		return AdminServiceListResult{}, fmt.Errorf("iterate admin services: %w", err)
	}

	return AdminServiceListResult{
		Items: items,
		Pagination: contracts.Pagination{
			Page:       int64(page),
			PageSize:   int64(pageSize),
			Total:      total,
			TotalPages: totalPages(total, pageSize),
		},
	}, nil
}

func (s Service) CreateAdminService(ctx context.Context, input CreateAdminServiceInput) (AdminService, error) {
	if _, err := s.ensureAdminPrincipal(ctx, input.AccessToken); err != nil {
		return AdminService{}, err
	}

	serviceID, err := s.newServiceID()
	if err != nil {
		return AdminService{}, err
	}

	status := "disabled"
	if input.Enabled {
		status = "enabled"
	}

	if _, err := s.db().ExecContext(
		ctx,
		`INSERT INTO services (id, key, name, description, group_name, protocol, upstream_url, public_path, status) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		serviceID,
		input.Key,
		input.Name,
		input.Description,
		input.Group,
		input.Protocol,
		input.UpstreamURL,
		input.PublicPath,
		status,
	); err != nil {
		if isUniqueViolation(err) {
			return AdminService{}, &ServiceError{
				StatusCode:  http.StatusConflict,
				Code:        contracts.ErrorCodeServiceAlreadyExists,
				Message:     "service already exists",
				UserMessage: "服务已存在",
			}
		}
		return AdminService{}, fmt.Errorf("insert service: %w", err)
	}

	return AdminService{
		ID:          serviceID,
		Key:         input.Key,
		Name:        input.Name,
		Description: input.Description,
		Group:       input.Group,
		Protocol:    input.Protocol,
		UpstreamURL: input.UpstreamURL,
		PublicPath:  input.PublicPath,
		Status:      status,
	}, nil
}

func (s Service) ListAdminDevices(ctx context.Context, input ListAdminDevicesInput) (AdminDeviceListResult, error) {
	if _, err := s.ensureAdminPrincipal(ctx, input.AccessToken); err != nil {
		return AdminDeviceListResult{}, err
	}
	page, pageSize := normalizePagination(input.Page, input.PageSize)
	where, args := buildAdminDeviceFilters(input)

	var total int64
	if err := s.db().QueryRowContext(ctx, "SELECT COUNT(*) FROM devices d INNER JOIN users u ON u.id = d.user_id "+where, args...).Scan(&total); err != nil {
		return AdminDeviceListResult{}, fmt.Errorf("count devices: %w", err)
	}

	queryArgs := append([]any{}, args...)
	queryArgs = append(queryArgs, pageSize, (page-1)*pageSize)
	query := `SELECT d.id, d.user_id, u.username, d.name, d.os, d.client_version, d.public_key_fingerprint, d.status
		FROM devices d INNER JOIN users u ON u.id = d.user_id ` + where + fmt.Sprintf(" ORDER BY d.created_at DESC LIMIT $%d OFFSET $%d", len(queryArgs)-1, len(queryArgs))
	rows, err := s.db().QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return AdminDeviceListResult{}, fmt.Errorf("query admin devices: %w", err)
	}
	defer rows.Close()

	items := []AdminDevice{}
	for rows.Next() {
		var device AdminDevice
		if err := rows.Scan(&device.ID, &device.UserID, &device.UserUsername, &device.Name, &device.OS, &device.ClientVersion, &device.PublicKeyFingerprint, &device.Status); err != nil {
			return AdminDeviceListResult{}, fmt.Errorf("scan admin device: %w", err)
		}
		items = append(items, device)
	}
	if err := rows.Err(); err != nil {
		return AdminDeviceListResult{}, fmt.Errorf("iterate admin devices: %w", err)
	}

	return AdminDeviceListResult{
		Items: items,
		Pagination: contracts.Pagination{
			Page:       int64(page),
			PageSize:   int64(pageSize),
			Total:      total,
			TotalPages: totalPages(total, pageSize),
		},
	}, nil
}

func (s Service) ListAdminAuditEvents(ctx context.Context, input ListAdminAuditEventsInput) (AdminAuditEventListResult, error) {
	if _, err := s.ensureAdminPrincipal(ctx, input.AccessToken); err != nil {
		return AdminAuditEventListResult{}, err
	}
	page, pageSize := normalizePagination(input.Page, input.PageSize)
	where, args := buildAdminAuditFilters(input)

	var total int64
	if err := s.db().QueryRowContext(ctx, "SELECT COUNT(*) FROM audit_events a "+where, args...).Scan(&total); err != nil {
		return AdminAuditEventListResult{}, fmt.Errorf("count audit events: %w", err)
	}

	queryArgs := append([]any{}, args...)
	queryArgs = append(queryArgs, pageSize, (page-1)*pageSize)
	query := `SELECT id, request_id, type, COALESCE(actor_user_id, ''), target_type, COALESCE(target_id, ''), COALESCE(service_id, ''), result, summary
		FROM audit_events a ` + where + fmt.Sprintf(" ORDER BY occurred_at DESC LIMIT $%d OFFSET $%d", len(queryArgs)-1, len(queryArgs))
	rows, err := s.db().QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return AdminAuditEventListResult{}, fmt.Errorf("query audit events: %w", err)
	}
	defer rows.Close()

	items := []AdminAuditEvent{}
	for rows.Next() {
		var event AdminAuditEvent
		if err := rows.Scan(&event.ID, &event.RequestID, &event.Type, &event.ActorUserID, &event.TargetType, &event.TargetID, &event.ServiceID, &event.Result, &event.Summary); err != nil {
			return AdminAuditEventListResult{}, fmt.Errorf("scan audit event: %w", err)
		}
		items = append(items, event)
	}
	if err := rows.Err(); err != nil {
		return AdminAuditEventListResult{}, fmt.Errorf("iterate audit events: %w", err)
	}

	return AdminAuditEventListResult{
		Items: items,
		Pagination: contracts.Pagination{
			Page:       int64(page),
			PageSize:   int64(pageSize),
			Total:      total,
			TotalPages: totalPages(total, pageSize),
		},
	}, nil
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

func (s Service) ReplaceUserServiceOverrides(ctx context.Context, input ReplaceUserServiceOverridesInput) ([]UserServiceOverride, error) {
	principal, err := s.ensureAdminPrincipal(ctx, input.AccessToken)
	if err != nil {
		return nil, err
	}
	if hasIntersection(input.AllowServiceIDs, input.DenyServiceIDs) {
		return nil, &ServiceError{
			StatusCode:  http.StatusUnprocessableEntity,
			Code:        contracts.ErrorCodePolicyRuleInvalid,
			Message:     "service override has conflicting allow and deny entries",
			UserMessage: "访问策略配置无效",
		}
	}

	tx, err := s.db().BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin user service override transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `DELETE FROM user_service_overrides WHERE user_id = $1`, input.UserID); err != nil {
		return nil, fmt.Errorf("delete user service overrides: %w", err)
	}

	overrides := []UserServiceOverride{}
	for _, serviceID := range input.AllowServiceIDs {
		if err := insertUserServiceOverrideTx(ctx, tx, input.UserID, serviceID, "allow", principal.User.ID); err != nil {
			return nil, err
		}
		overrides = append(overrides, UserServiceOverride{ServiceID: serviceID, Effect: "allow"})
	}
	for _, serviceID := range input.DenyServiceIDs {
		if err := insertUserServiceOverrideTx(ctx, tx, input.UserID, serviceID, "deny", principal.User.ID); err != nil {
			return nil, err
		}
		overrides = append(overrides, UserServiceOverride{ServiceID: serviceID, Effect: "deny"})
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit user service override transaction: %w", err)
	}
	return overrides, nil
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

func insertUserServiceOverrideTx(ctx context.Context, tx *sql.Tx, userID string, serviceID string, effect string, createdBy string) error {
	if _, err := tx.ExecContext(ctx, `INSERT INTO user_service_overrides (user_id, service_id, effect, reason, created_by) VALUES ($1, $2, $3, '', $4)`, userID, serviceID, effect, createdBy); err != nil {
		return fmt.Errorf("insert user service override: %w", err)
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

func buildAdminServiceFilters(input ListAdminServicesInput) (string, []any) {
	conditions := []string{"WHERE true"}
	args := []any{}
	if input.Keyword != "" {
		args = append(args, "%"+strings.ToLower(input.Keyword)+"%")
		conditions = append(conditions, fmt.Sprintf("(LOWER(s.key) LIKE $%d OR LOWER(s.name) LIKE $%d OR LOWER(s.description) LIKE $%d)", len(args), len(args), len(args)))
	}
	if input.Status != "" {
		args = append(args, input.Status)
		conditions = append(conditions, fmt.Sprintf("s.status = $%d", len(args)))
	}
	if input.Group != "" {
		args = append(args, input.Group)
		conditions = append(conditions, fmt.Sprintf("s.group_name = $%d", len(args)))
	}
	return strings.Join(conditions, " AND "), args
}

func buildAdminDeviceFilters(input ListAdminDevicesInput) (string, []any) {
	conditions := []string{"WHERE true"}
	args := []any{}
	if input.Keyword != "" {
		args = append(args, "%"+strings.ToLower(input.Keyword)+"%")
		conditions = append(conditions, fmt.Sprintf("(LOWER(d.name) LIKE $%d OR LOWER(u.username) LIKE $%d OR LOWER(d.public_key_fingerprint) LIKE $%d)", len(args), len(args), len(args)))
	}
	if input.Status != "" {
		args = append(args, input.Status)
		conditions = append(conditions, fmt.Sprintf("d.status = $%d", len(args)))
	}
	if input.UserID != "" {
		args = append(args, input.UserID)
		conditions = append(conditions, fmt.Sprintf("d.user_id = $%d", len(args)))
	}
	return strings.Join(conditions, " AND "), args
}

func buildAdminAuditFilters(input ListAdminAuditEventsInput) (string, []any) {
	conditions := []string{"WHERE true"}
	args := []any{}
	add := func(column string, value string) {
		if value == "" {
			return
		}
		args = append(args, value)
		conditions = append(conditions, fmt.Sprintf("%s = $%d", column, len(args)))
	}
	add("a.type", input.Type)
	add("a.actor_user_id", input.ActorUserID)
	add("a.target_type", input.TargetType)
	add("a.target_id", input.TargetID)
	add("a.service_id", input.ServiceID)
	add("a.result", input.Result)
	return strings.Join(conditions, " AND "), args
}

func buildSimpleKeywordFilter(alias string, columns []string, keyword string, extraCondition string) (string, []any) {
	conditions := []string{"WHERE true"}
	args := []any{}
	if extraCondition != "" {
		conditions = append(conditions, extraCondition)
	}
	if keyword != "" {
		args = append(args, "%"+strings.ToLower(keyword)+"%")
		parts := make([]string, 0, len(columns))
		for _, column := range columns {
			parts = append(parts, fmt.Sprintf("LOWER(%s.%s) LIKE $%d", alias, column, len(args)))
		}
		conditions = append(conditions, "("+strings.Join(parts, " OR ")+")")
	}
	return strings.Join(conditions, " AND "), args
}

func hasIntersection(left []string, right []string) bool {
	seen := map[string]bool{}
	for _, item := range left {
		seen[item] = true
	}
	for _, item := range right {
		if seen[item] {
			return true
		}
	}
	return false
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
