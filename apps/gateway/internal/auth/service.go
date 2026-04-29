package auth

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/kittors/bifrost/apps/gateway/internal/contracts"
)

const (
	adminRoleID = "role_admin"
)

type Service struct {
	DB                 *sql.DB
	PasswordHasher     PasswordHasher
	TokenIssuer        TokenIssuer
	Now                func() time.Time
	RefreshTokenTTL    time.Duration
	SessionIDFactory   func() (string, error)
	DeviceIDFactory    func() (string, error)
	ChallengeIDFactory func() (string, error)
	UserIDFactory      func() (string, error)
	RoleIDFactory      func() (string, error)
	ServiceIDFactory   func() (string, error)
	ChallengeTTL       time.Duration
}

type auditEventInput struct {
	RequestID   string
	Type        contracts.AuditEventType
	ActorUserID string
	TargetType  string
	TargetID    string
	ServiceID   string
	Result      string
	Summary     string
}

// 基础设施与 ID 工厂保留在主文件，作为整个 Service 的稳定根节点。
func (s Service) db() *sql.DB {
	return s.DB
}

func (s Service) passwordHasher() PasswordHasher {
	if s.PasswordHasher == (PasswordHasher{}) {
		return DefaultPasswordHasher()
	}
	return s.PasswordHasher
}

func (s Service) tokenIssuer() TokenIssuer {
	issuer := s.TokenIssuer
	if issuer.Now == nil {
		issuer.Now = s.now
	}
	return issuer
}

func (s Service) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}

func (s Service) refreshTokenTTL() time.Duration {
	if s.RefreshTokenTTL > 0 {
		return s.RefreshTokenTTL
	}
	return 7 * 24 * time.Hour
}

func (s Service) newSessionID() (string, error) {
	if s.SessionIDFactory != nil {
		return s.SessionIDFactory()
	}

	token, err := GenerateRefreshToken()
	if err != nil {
		return "", fmt.Errorf("generate session id: %w", err)
	}
	return "sess_" + token[:20], nil
}

func (s Service) newDeviceID() (string, error) {
	if s.DeviceIDFactory != nil {
		return s.DeviceIDFactory()
	}

	token, err := GenerateRefreshToken()
	if err != nil {
		return "", fmt.Errorf("generate device id: %w", err)
	}
	return "dev_" + token[:20], nil
}

func (s Service) newChallengeID() (string, error) {
	if s.ChallengeIDFactory != nil {
		return s.ChallengeIDFactory()
	}

	token, err := GenerateRefreshToken()
	if err != nil {
		return "", fmt.Errorf("generate challenge id: %w", err)
	}
	return "ch_" + token[:20], nil
}

func (s Service) newUserID() (string, error) {
	if s.UserIDFactory != nil {
		return s.UserIDFactory()
	}

	token, err := GenerateRefreshToken()
	if err != nil {
		return "", fmt.Errorf("generate user id: %w", err)
	}
	return "user_" + token[:20], nil
}

func (s Service) newRoleID() (string, error) {
	if s.RoleIDFactory != nil {
		return s.RoleIDFactory()
	}

	token, err := GenerateRefreshToken()
	if err != nil {
		return "", fmt.Errorf("generate role id: %w", err)
	}
	return "role_" + token[:20], nil
}

func (s Service) newServiceID() (string, error) {
	if s.ServiceIDFactory != nil {
		return s.ServiceIDFactory()
	}

	token, err := GenerateRefreshToken()
	if err != nil {
		return "", fmt.Errorf("generate service id: %w", err)
	}
	return "service_" + token[:20], nil
}

func (s Service) newAuditEventID() (string, error) {
	token, err := GenerateRefreshToken()
	if err != nil {
		return "", fmt.Errorf("generate audit event id: %w", err)
	}
	return "audit_" + token[:20], nil
}

func (s Service) recordAuditEvent(ctx context.Context, input auditEventInput) error {
	eventID, err := s.newAuditEventID()
	if err != nil {
		return err
	}

	requestID := strings.TrimSpace(input.RequestID)
	if requestID == "" {
		requestID = fmt.Sprintf("req_internal_%d", s.now().UTC().UnixNano())
	}

	var serviceID any
	if strings.TrimSpace(input.ServiceID) != "" {
		serviceID = input.ServiceID
	}

	if _, err := s.db().ExecContext(
		ctx,
		`INSERT INTO audit_events (id, request_id, type, actor_user_id, target_type, target_id, service_id, result, summary)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		eventID,
		requestID,
		string(input.Type),
		nullIfEmpty(input.ActorUserID),
		input.TargetType,
		nullIfEmpty(input.TargetID),
		serviceID,
		input.Result,
		input.Summary,
	); err != nil {
		return fmt.Errorf("insert audit event: %w", err)
	}

	return nil
}

func isUniqueViolation(err error) bool {
	return strings.Contains(err.Error(), "SQLSTATE 23505")
}

func nullIfEmpty(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
