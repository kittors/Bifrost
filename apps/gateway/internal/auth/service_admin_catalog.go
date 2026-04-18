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

// 后台目录类查询与写入聚合在这里，方便继续扩展角色、服务、设备、审计管理。
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

func (s Service) GetAdminService(ctx context.Context, input GetAdminServiceInput) (AdminService, error) {
	if _, err := s.ensureAdminPrincipal(ctx, input.AccessToken); err != nil {
		return AdminService{}, err
	}

	return s.loadAdminService(ctx, input.ServiceID)
}

func (s Service) UpdateAdminService(ctx context.Context, input UpdateAdminServiceInput) (AdminService, error) {
	if _, err := s.ensureAdminPrincipal(ctx, input.AccessToken); err != nil {
		return AdminService{}, err
	}

	result, err := s.db().ExecContext(
		ctx,
		`UPDATE services
		SET name = $2, description = $3, group_name = $4, protocol = $5, upstream_url = $6, public_path = $7, updated_at = $8
		WHERE id = $1`,
		input.ServiceID,
		input.Name,
		input.Description,
		input.Group,
		input.Protocol,
		input.UpstreamURL,
		input.PublicPath,
		s.now().UTC(),
	)
	if err != nil {
		return AdminService{}, fmt.Errorf("update admin service: %w", err)
	}
	if err := ensureServiceMutationAffected(result); err != nil {
		return AdminService{}, err
	}

	return s.loadAdminService(ctx, input.ServiceID)
}

func (s Service) SetAdminServiceStatus(ctx context.Context, input SetAdminServiceStatusInput) (AdminService, error) {
	principal, err := s.ensureAdminPrincipal(ctx, input.AccessToken)
	if err != nil {
		return AdminService{}, err
	}

	status := strings.TrimSpace(input.Status)
	if status != "enabled" && status != "disabled" {
		return AdminService{}, &ServiceError{
			StatusCode:  http.StatusBadRequest,
			Code:        contracts.ErrorCodeCommonBadRequest,
			Message:     "status must be enabled or disabled",
			UserMessage: "请求参数不正确",
		}
	}

	result, err := s.db().ExecContext(
		ctx,
		`UPDATE services
		SET status = $2, updated_at = $3
		WHERE id = $1`,
		input.ServiceID,
		status,
		s.now().UTC(),
	)
	if err != nil {
		return AdminService{}, fmt.Errorf("update admin service status: %w", err)
	}
	if err := ensureServiceMutationAffected(result); err != nil {
		return AdminService{}, err
	}

	if err := s.recordAuditEvent(ctx, auditEventInput{
		RequestID:   input.RequestID,
		Type:        contracts.AuditEventTypeAdminServiceUpdated,
		ActorUserID: principal.User.ID,
		TargetType:  "service",
		TargetID:    input.ServiceID,
		Result:      "success",
		Summary:     "admin service status updated",
	}); err != nil {
		return AdminService{}, err
	}

	return s.loadAdminService(ctx, input.ServiceID)
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
				StatusCode:  409,
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

func (s Service) GetAdminDevice(ctx context.Context, input GetAdminDeviceInput) (AdminDevice, error) {
	if _, err := s.ensureAdminPrincipal(ctx, input.AccessToken); err != nil {
		return AdminDevice{}, err
	}

	return s.loadAdminDevice(ctx, input.DeviceID)
}

func (s Service) SetAdminDeviceStatus(ctx context.Context, input SetAdminDeviceStatusInput) (AdminDevice, error) {
	if _, err := s.ensureAdminPrincipal(ctx, input.AccessToken); err != nil {
		return AdminDevice{}, err
	}

	status := strings.TrimSpace(input.Status)
	if status != "trusted" && status != "disabled" {
		return AdminDevice{}, &ServiceError{
			StatusCode:  http.StatusBadRequest,
			Code:        contracts.ErrorCodeCommonBadRequest,
			Message:     "status must be trusted or disabled",
			UserMessage: "请求参数不正确",
		}
	}

	result, err := s.db().ExecContext(
		ctx,
		`UPDATE devices
		SET status = $2, updated_at = $3
		WHERE id = $1`,
		input.DeviceID,
		status,
		s.now().UTC(),
	)
	if err != nil {
		return AdminDevice{}, fmt.Errorf("update admin device status: %w", err)
	}
	if err := ensureDeviceMutationAffected(result); err != nil {
		return AdminDevice{}, err
	}

	if status == "disabled" {
		if _, err := s.db().ExecContext(
			ctx,
			`UPDATE sessions
			SET status = 'revoked', revoked_at = $2
			WHERE device_id = $1 AND status = 'active'`,
			input.DeviceID,
			s.now().UTC(),
		); err != nil {
			return AdminDevice{}, fmt.Errorf("revoke sessions for disabled device: %w", err)
		}
	}

	return s.loadAdminDevice(ctx, input.DeviceID)
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

func (s Service) loadAdminService(ctx context.Context, serviceID string) (AdminService, error) {
	row := s.db().QueryRowContext(
		ctx,
		`SELECT id, key, name, description, group_name, protocol, upstream_url, public_path, status
		FROM services
		WHERE id = $1`,
		serviceID,
	)

	var service AdminService
	if err := row.Scan(&service.ID, &service.Key, &service.Name, &service.Description, &service.Group, &service.Protocol, &service.UpstreamURL, &service.PublicPath, &service.Status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return AdminService{}, &ServiceError{
				StatusCode:  http.StatusNotFound,
				Code:        contracts.ErrorCodeServiceNotFound,
				Message:     "service not found",
				UserMessage: "服务不存在",
			}
		}
		return AdminService{}, fmt.Errorf("query admin service: %w", err)
	}

	return service, nil
}

func (s Service) loadAdminDevice(ctx context.Context, deviceID string) (AdminDevice, error) {
	row := s.db().QueryRowContext(
		ctx,
		`SELECT d.id, d.user_id, u.username, d.name, d.os, d.client_version, d.public_key_fingerprint, d.status
		FROM devices d
		INNER JOIN users u ON u.id = d.user_id
		WHERE d.id = $1`,
		deviceID,
	)

	var device AdminDevice
	if err := row.Scan(&device.ID, &device.UserID, &device.UserUsername, &device.Name, &device.OS, &device.ClientVersion, &device.PublicKeyFingerprint, &device.Status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return AdminDevice{}, &ServiceError{
				StatusCode:  http.StatusNotFound,
				Code:        contracts.ErrorCodeDeviceNotFound,
				Message:     "device not found",
				UserMessage: "设备不存在",
			}
		}
		return AdminDevice{}, fmt.Errorf("query admin device: %w", err)
	}

	return device, nil
}

func ensureServiceMutationAffected(result sql.Result) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("admin service mutation rows affected: %w", err)
	}
	if affected == 0 {
		return &ServiceError{
			StatusCode:  http.StatusNotFound,
			Code:        contracts.ErrorCodeServiceNotFound,
			Message:     "service not found",
			UserMessage: "服务不存在",
		}
	}
	return nil
}

func ensureDeviceMutationAffected(result sql.Result) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("admin device mutation rows affected: %w", err)
	}
	if affected == 0 {
		return &ServiceError{
			StatusCode:  http.StatusNotFound,
			Code:        contracts.ErrorCodeDeviceNotFound,
			Message:     "device not found",
			UserMessage: "设备不存在",
		}
	}
	return nil
}
