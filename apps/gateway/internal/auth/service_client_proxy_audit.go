package auth

import (
	"context"
	"strings"
)

// 代理访问审计单独收敛，普通代理编排不需要关心审计事件落库细节。
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
