package auth

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/kittors/bifrost/apps/gateway/internal/contracts"
)

// 后台策略覆盖与筛选辅助工具单独放置，避免查询入口文件继续膨胀。
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

func insertUserServiceOverrideTx(ctx context.Context, tx *sql.Tx, userID string, serviceID string, effect string, createdBy string) error {
	if _, err := tx.ExecContext(ctx, `INSERT INTO user_service_overrides (user_id, service_id, effect, reason, created_by) VALUES ($1, $2, $3, '', $4)`, userID, serviceID, effect, createdBy); err != nil {
		return fmt.Errorf("insert user service override: %w", err)
	}
	return nil
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
