package auth

import (
	"database/sql"
	"time"

	"github.com/kittors/bifrost/apps/gateway/internal/contracts"
)

type LoginResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
	User         LoginUser
}

type LoginUser struct {
	ID          string
	Username    string
	DisplayName string
	Roles       []string
}

type DeviceResult struct {
	ID     string
	Status string
}

type ClientBootstrapResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
	User         LoginUser
	Device       DeviceResult
}

type DeviceChallengeResult struct {
	ID        string
	Challenge string
	ExpiresIn int
}

type DeviceChallengeVerificationResult struct {
	Verified bool
}

type ClientService struct {
	ID           string
	Key          string
	Name         string
	Description  string
	Group        string
	Status       string
	PublicPath   string
	AccessSource string
}

type ServiceAccessURLResult struct {
	PublicPath   string
	ExpiresIn    int
	AccessTicket string
}

type ResolveProxyRequestResult struct {
	ServiceID    string
	ServiceKey   string
	ServiceName  string
	PublicPath   string
	UpstreamURL  string
	UserID       string
	DeviceID     string
	AccessSource string
}

type AdminUser struct {
	ID          string
	Username    string
	DisplayName string
	Email       string
	Status      string
	Roles       []string
}

type AdminUserListResult struct {
	Items      []AdminUser
	Pagination contracts.Pagination
}

type AdminRole struct {
	ID          string
	Name        string
	DisplayName string
	Description string
}

type AdminRoleListResult struct {
	Items      []AdminRole
	Pagination contracts.Pagination
}

type AdminService struct {
	ID          string
	Key         string
	Name        string
	Description string
	Group       string
	Protocol    string
	UpstreamURL string
	PublicPath  string
	Status      string
}

type AdminServiceListResult struct {
	Items      []AdminService
	Pagination contracts.Pagination
}

type AdminDevice struct {
	ID                   string
	UserID               string
	UserUsername         string
	Name                 string
	OS                   string
	ClientVersion        string
	PublicKeyFingerprint string
	Status               string
}

type AdminDeviceListResult struct {
	Items      []AdminDevice
	Pagination contracts.Pagination
}

type AdminAuditEvent struct {
	ID          string
	RequestID   string
	Type        string
	ActorUserID string
	TargetType  string
	TargetID    string
	ServiceID   string
	Result      string
	Summary     string
}

type AdminAuditEventListResult struct {
	Items      []AdminAuditEvent
	Pagination contracts.Pagination
}

type UserServiceOverride struct {
	ServiceID string
	Effect    string
}

type ServiceError struct {
	StatusCode  int
	Code        contracts.ErrorCode
	Message     string
	UserMessage string
}

func (e *ServiceError) Error() string {
	return e.Message
}

// 内部 record 类型单独集中，减少 service_auth.go 里的样板噪音。
type userRecord struct {
	ID           string
	Username     string
	DisplayName  string
	PasswordHash string
	Status       string
	RoleIDs      []string
}

type sessionRecord struct {
	ID               string
	UserID           string
	DeviceID         sql.NullString
	RefreshTokenHash string
	Status           string
	ExpiresAt        time.Time
}

type clientPrincipal struct {
	Claims AccessTokenClaims
	User   userRecord
}
