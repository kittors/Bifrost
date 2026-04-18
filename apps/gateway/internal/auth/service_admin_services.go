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

// 服务目录管理单独收敛，后续新增协议校验或上游探测时只影响本文件。
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
