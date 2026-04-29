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

// 设备目录相关逻辑单独放置，便于后续继续扩展设备吊销、风险标记等能力。
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
