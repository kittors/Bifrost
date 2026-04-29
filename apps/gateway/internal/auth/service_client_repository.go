package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/kittors/bifrost/apps/gateway/internal/contracts"
)

type proxyServiceRecord struct {
	ID          string
	Key         string
	Name        string
	PublicPath  string
	UpstreamURL string
	Status      string
}

// 客户端服务读取集中在这里，保持上层 service 方法只编排业务流程。
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

func (s Service) loadProxyServiceByKey(ctx context.Context, serviceKey string) (proxyServiceRecord, error) {
	row := s.db().QueryRowContext(
		ctx,
		`SELECT id, key, name, public_path, upstream_url, status
		FROM services
		WHERE key = $1`,
		serviceKey,
	)

	var service proxyServiceRecord
	if err := row.Scan(&service.ID, &service.Key, &service.Name, &service.PublicPath, &service.UpstreamURL, &service.Status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return proxyServiceRecord{}, &ServiceError{
				StatusCode:  http.StatusNotFound,
				Code:        contracts.ErrorCodeServiceNotFound,
				Message:     "service not found",
				UserMessage: "服务不存在",
			}
		}
		return proxyServiceRecord{}, fmt.Errorf("query proxy service by key: %w", err)
	}

	return service, nil
}
