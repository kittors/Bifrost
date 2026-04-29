package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// 客户端服务访问策略集中在这里：用户级 deny 优先，其次用户 allow，最后角色授权。
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
