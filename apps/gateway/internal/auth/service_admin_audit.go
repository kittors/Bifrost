package auth

import (
	"context"
	"fmt"

	"github.com/kittors/bifrost/apps/gateway/internal/contracts"
)

// 审计事件查询独立成文件，便于后续扩展更多过滤条件或详情读取能力。
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
