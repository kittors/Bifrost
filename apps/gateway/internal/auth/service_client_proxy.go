package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/kittors/bifrost/apps/gateway/internal/contracts"
)

// 客户端服务目录与网关代理授权都依赖同一套访问判断，因此集中在一个文件维护。
func (s Service) ListClientServices(ctx context.Context, input ListClientServicesInput) ([]ClientService, error) {
	principal, err := s.loadClientPrincipal(ctx, input.AccessToken)
	if err != nil {
		return nil, err
	}

	rows, err := s.db().QueryContext(
		ctx,
		`SELECT id, key, name, description, group_name, status, public_path
		FROM services
		WHERE status = 'enabled'
		ORDER BY name ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("query services: %w", err)
	}
	defer rows.Close()

	var services []ClientService
	for rows.Next() {
		var service ClientService
		if err := rows.Scan(&service.ID, &service.Key, &service.Name, &service.Description, &service.Group, &service.Status, &service.PublicPath); err != nil {
			return nil, fmt.Errorf("scan service: %w", err)
		}

		accessSource, allowed, err := s.resolveServiceAccess(ctx, principal.User.ID, principal.User.RoleIDs, service.ID)
		if err != nil {
			return nil, err
		}
		if !allowed {
			continue
		}

		if input.Keyword != "" && !strings.Contains(strings.ToLower(service.Name+" "+service.Key+" "+service.Description), strings.ToLower(input.Keyword)) {
			continue
		}

		if input.Group != "" && service.Group != input.Group {
			continue
		}

		service.AccessSource = accessSource
		services = append(services, service)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate services: %w", err)
	}

	return services, nil
}

func (s Service) GetClientService(ctx context.Context, input GetClientServiceInput) (ClientService, error) {
	principal, err := s.loadClientPrincipal(ctx, input.AccessToken)
	if err != nil {
		return ClientService{}, err
	}

	service, err := s.loadService(ctx, input.ServiceID)
	if err != nil {
		return ClientService{}, err
	}

	accessSource, allowed, err := s.resolveServiceAccess(ctx, principal.User.ID, principal.User.RoleIDs, service.ID)
	if err != nil {
		return ClientService{}, err
	}
	if !allowed {
		return ClientService{}, &ServiceError{
			StatusCode:  http.StatusForbidden,
			Code:        contracts.ErrorCodePolicyAccessDenied,
			Message:     "user is not allowed to access service",
			UserMessage: "你没有访问该服务的权限",
		}
	}

	service.AccessSource = accessSource
	return service, nil
}

func (s Service) CreateServiceAccessURL(ctx context.Context, input CreateServiceAccessURLInput) (ServiceAccessURLResult, error) {
	principal, err := s.loadClientPrincipal(ctx, input.AccessToken)
	if err != nil {
		return ServiceAccessURLResult{}, err
	}

	service, err := s.loadService(ctx, input.ServiceID)
	if err != nil {
		return ServiceAccessURLResult{}, err
	}

	_, allowed, err := s.resolveServiceAccess(ctx, principal.User.ID, principal.User.RoleIDs, service.ID)
	if err != nil {
		return ServiceAccessURLResult{}, err
	}
	if !allowed {
		return ServiceAccessURLResult{}, &ServiceError{
			StatusCode:  http.StatusForbidden,
			Code:        contracts.ErrorCodePolicyAccessDenied,
			Message:     "user is not allowed to access service",
			UserMessage: "你没有访问该服务的权限",
		}
	}

	ticketIssuer := s.tokenIssuer()
	ticketIssuer.TTL = 5 * time.Minute
	accessTicket, expiresAt, err := ticketIssuer.IssueServiceAccessTicket(ServiceAccessTicketClaims{
		UserID:    principal.User.ID,
		DeviceID:  principal.Claims.DeviceID,
		SessionID: principal.Claims.SessionID,
		ServiceID: service.ID,
	})
	if err != nil {
		return ServiceAccessURLResult{}, fmt.Errorf("issue service access ticket: %w", err)
	}

	return ServiceAccessURLResult{
		PublicPath:   service.PublicPath,
		ExpiresIn:    int(expiresAt.Sub(s.now().UTC()).Seconds()),
		AccessTicket: accessTicket,
	}, nil
}

func (s Service) ResolveProxyRequest(ctx context.Context, input ResolveProxyRequestInput) (ResolveProxyRequestResult, error) {
	service, err := s.loadProxyServiceByKey(ctx, input.ServiceKey)
	if err != nil {
		return ResolveProxyRequestResult{}, err
	}

	var principal clientPrincipal
	if strings.TrimSpace(input.AccessToken) != "" {
		principal, err = s.loadClientPrincipal(ctx, input.AccessToken)
		if err != nil {
			return ResolveProxyRequestResult{}, err
		}
	} else if strings.TrimSpace(input.AccessTicket) != "" {
		ticketClaims, err := s.tokenIssuer().VerifyServiceAccessTicket(input.AccessTicket)
		if err != nil {
			return ResolveProxyRequestResult{}, mapTokenError(err)
		}
		if ticketClaims.ServiceID != service.ID {
			return ResolveProxyRequestResult{}, &ServiceError{
				StatusCode:  http.StatusUnauthorized,
				Code:        contracts.ErrorCodeAuthInvalidToken,
				Message:     "service access ticket does not match requested service",
				UserMessage: "登录状态已失效，请重新登录",
			}
		}

		principal, err = s.loadClientPrincipalFromClaims(ctx, AccessTokenClaims{
			UserID:    ticketClaims.UserID,
			DeviceID:  ticketClaims.DeviceID,
			SessionID: ticketClaims.SessionID,
			IssuedAt:  ticketClaims.IssuedAt,
			ExpiresAt: ticketClaims.ExpiresAt,
		})
		if err != nil {
			return ResolveProxyRequestResult{}, err
		}
	} else {
		return ResolveProxyRequestResult{}, &ServiceError{
			StatusCode:  http.StatusUnauthorized,
			Code:        contracts.ErrorCodeAuthInvalidToken,
			Message:     "proxy access requires bearer token or service access ticket",
			UserMessage: "登录状态已失效，请重新登录",
		}
	}

	if service.Status != "enabled" {
		if err := s.RecordProxyAccessEvent(ctx, RecordProxyAccessEventInput{
			RequestID: input.RequestID,
			Type:      contracts.AuditEventTypeServiceAccessDenied,
			UserID:    principal.User.ID,
			DeviceID:  principal.Claims.DeviceID,
			ServiceID: service.ID,
			Result:    "failure",
			Summary:   "service is disabled",
		}); err != nil {
			return ResolveProxyRequestResult{}, err
		}
		return ResolveProxyRequestResult{}, &ServiceError{
			StatusCode:  http.StatusForbidden,
			Code:        contracts.ErrorCodeServiceDisabled,
			Message:     "service is disabled",
			UserMessage: "服务当前不可用",
		}
	}

	if strings.TrimSpace(service.UpstreamURL) == "" {
		return ResolveProxyRequestResult{}, &ServiceError{
			StatusCode:  http.StatusBadGateway,
			Code:        contracts.ErrorCodeServiceUpstreamInvalid,
			Message:     "service upstream url is empty",
			UserMessage: "服务暂时不可用，请稍后再试",
		}
	}

	accessSource, allowed, err := s.resolveServiceAccess(ctx, principal.User.ID, principal.User.RoleIDs, service.ID)
	if err != nil {
		return ResolveProxyRequestResult{}, err
	}
	if !allowed {
		if err := s.RecordProxyAccessEvent(ctx, RecordProxyAccessEventInput{
			RequestID: input.RequestID,
			Type:      contracts.AuditEventTypeServiceAccessDenied,
			UserID:    principal.User.ID,
			DeviceID:  principal.Claims.DeviceID,
			ServiceID: service.ID,
			Result:    "failure",
			Summary:   "service access denied",
		}); err != nil {
			return ResolveProxyRequestResult{}, err
		}
		return ResolveProxyRequestResult{}, &ServiceError{
			StatusCode:  http.StatusForbidden,
			Code:        contracts.ErrorCodePolicyAccessDenied,
			Message:     "user is not allowed to access service",
			UserMessage: "你没有访问该服务的权限",
		}
	}

	return ResolveProxyRequestResult{
		ServiceID:    service.ID,
		ServiceKey:   service.Key,
		ServiceName:  service.Name,
		PublicPath:   service.PublicPath,
		UpstreamURL:  service.UpstreamURL,
		UserID:       principal.User.ID,
		DeviceID:     principal.Claims.DeviceID,
		AccessSource: accessSource,
	}, nil
}

func (s Service) RecordProxyAccessEvent(ctx context.Context, input RecordProxyAccessEventInput) error {
	summary := strings.TrimSpace(input.Summary)
	if summary == "" {
		summary = "service access event"
	}

	return s.recordAuditEvent(ctx, auditEventInput{
		RequestID:   input.RequestID,
		Type:        input.Type,
		ActorUserID: input.UserID,
		TargetType:  "service",
		TargetID:    input.ServiceID,
		ServiceID:   input.ServiceID,
		Result:      input.Result,
		Summary:     summary,
	})
}

type proxyServiceRecord struct {
	ID          string
	Key         string
	Name        string
	PublicPath  string
	UpstreamURL string
	Status      string
}

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

func (s Service) resolveServiceAccess(ctx context.Context, userID string, roleIDs []string, serviceID string) (string, bool, error) {
	row := s.db().QueryRowContext(
		ctx,
		`SELECT effect
		FROM user_service_overrides
		WHERE user_id = $1 AND service_id = $2`,
		userID,
		serviceID,
	)

	var effect string
	if err := row.Scan(&effect); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", false, fmt.Errorf("query user service override: %w", err)
	}
	if effect == "deny" {
		return "deny", false, nil
	}
	if effect == "allow" {
		return "user", true, nil
	}

	if len(roleIDs) == 0 {
		return "", false, nil
	}

	placeholders := make([]string, 0, len(roleIDs))
	args := make([]any, 0, len(roleIDs)+1)
	args = append(args, serviceID)
	for index, roleID := range roleIDs {
		placeholders = append(placeholders, fmt.Sprintf("$%d", index+2))
		args = append(args, roleID)
	}

	query := fmt.Sprintf(
		`SELECT EXISTS (
			SELECT 1
			FROM role_services
			WHERE service_id = $1 AND role_id IN (%s)
		)`,
		strings.Join(placeholders, ","),
	)

	var roleAllowed bool
	if err := s.db().QueryRowContext(ctx, query, args...).Scan(&roleAllowed); err != nil {
		return "", false, fmt.Errorf("query role service access: %w", err)
	}
	if roleAllowed {
		return "role", true, nil
	}

	return "", false, nil
}
