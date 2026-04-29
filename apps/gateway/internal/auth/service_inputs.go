package auth

import "github.com/kittors/bifrost/apps/gateway/internal/contracts"

// 输入模型集中描述外部用例，避免业务实现文件被大量 DTO 淹没。
type AdminLoginInput struct {
	Username  string
	Password  string
	RequestID string
}

type ClientLoginInput struct {
	Username      string
	Password      string
	DeviceID      string
	ClientVersion string
	RequestID     string
}

type BootstrapClientDeviceInput struct {
	Username             string
	Password             string
	DeviceName           string
	DeviceOS             string
	ClientVersion        string
	PublicKey            string
	PublicKeyFingerprint string
	RequestID            string
}

type RefreshInput struct {
	RefreshToken string
	DeviceID     string
}

type LogoutInput struct {
	AccessToken string
}

type CurrentUserInput struct {
	AccessToken string
}

type RegisterDeviceInput struct {
	AccessToken          string
	Name                 string
	OS                   string
	ClientVersion        string
	PublicKey            string
	PublicKeyFingerprint string
}

type CreateDeviceChallengeInput struct {
	AccessToken string
	DeviceID    string
}

type VerifyDeviceChallengeInput struct {
	AccessToken string
	ChallengeID string
	Signature   string
}

type ListClientServicesInput struct {
	AccessToken string
	Keyword     string
	Group       string
}

type GetClientServiceInput struct {
	AccessToken string
	ServiceID   string
}

type CreateServiceAccessURLInput struct {
	AccessToken string
	ServiceID   string
}

type ResolveProxyRequestInput struct {
	AccessToken  string
	AccessTicket string
	RequestID    string
	ServiceKey   string
}

type RecordProxyAccessEventInput struct {
	RequestID string
	Type      contracts.AuditEventType
	UserID    string
	DeviceID  string
	ServiceID string
	Result    string
	Summary   string
}

type ListAdminUsersInput struct {
	AccessToken string
	Page        int
	PageSize    int
	Keyword     string
	Status      string
	RoleID      string
}

type CreateAdminUserInput struct {
	AccessToken string
	RequestID   string
	Username    string
	DisplayName string
	Email       string
	Password    string
	RoleIDs     []string
}

type UpdateAdminUserInput struct {
	AccessToken string
	UserID      string
	DisplayName string
	Email       string
	RoleIDs     []string
}

type GetAdminUserInput struct {
	AccessToken string
	UserID      string
}

type ResetAdminUserPasswordInput struct {
	AccessToken string
	RequestID   string
	UserID      string
	Password    string
}

type SetAdminUserStatusInput struct {
	AccessToken string
	RequestID   string
	UserID      string
	Status      string
}

type ListAdminRolesInput struct {
	AccessToken string
	Page        int
	PageSize    int
	Keyword     string
}

type CreateAdminRoleInput struct {
	AccessToken string
	Name        string
	DisplayName string
	Description string
}

type UpdateAdminRoleInput struct {
	AccessToken string
	RoleID      string
	DisplayName string
	Description string
}

type GetAdminServiceInput struct {
	AccessToken string
	ServiceID   string
}

type ListAdminServicesInput struct {
	AccessToken string
	Page        int
	PageSize    int
	Keyword     string
	Status      string
	Group       string
}

type UpdateAdminServiceInput struct {
	AccessToken string
	ServiceID   string
	Name        string
	Description string
	Group       string
	Protocol    string
	UpstreamURL string
	PublicPath  string
}

type SetAdminServiceStatusInput struct {
	AccessToken string
	RequestID   string
	ServiceID   string
	Status      string
}

type CreateAdminServiceInput struct {
	AccessToken string
	Key         string
	Name        string
	Description string
	Group       string
	Protocol    string
	UpstreamURL string
	PublicPath  string
	Enabled     bool
}

type GetAdminDeviceInput struct {
	AccessToken string
	DeviceID    string
}

type ListAdminDevicesInput struct {
	AccessToken string
	Page        int
	PageSize    int
	Keyword     string
	Status      string
	UserID      string
}

type SetAdminDeviceStatusInput struct {
	AccessToken string
	RequestID   string
	DeviceID    string
	Status      string
}

type ListAdminAuditEventsInput struct {
	AccessToken string
	Page        int
	PageSize    int
	Type        string
	ActorUserID string
	TargetType  string
	TargetID    string
	ServiceID   string
	Result      string
}

type ReplaceRoleServicesInput struct {
	AccessToken string
	RoleID      string
	ServiceIDs  []string
}

type ReplaceUserServiceOverridesInput struct {
	AccessToken     string
	UserID          string
	AllowServiceIDs []string
	DenyServiceIDs  []string
}

type ListUserServiceOverridesInput struct {
	AccessToken string
	UserID      string
}
