package contracts

type ErrorCode string

const (
	ErrorCodeCommonBadRequest           ErrorCode = "COMMON_BAD_REQUEST"
	ErrorCodeCommonValidationFailed     ErrorCode = "COMMON_VALIDATION_FAILED"
	ErrorCodeCommonNotFound             ErrorCode = "COMMON_NOT_FOUND"
	ErrorCodeCommonConflict             ErrorCode = "COMMON_CONFLICT"
	ErrorCodeCommonRateLimited          ErrorCode = "COMMON_RATE_LIMITED"
	ErrorCodeCommonInternalError        ErrorCode = "COMMON_INTERNAL_ERROR"
	ErrorCodeAuthInvalidCredentials     ErrorCode = "AUTH_INVALID_CREDENTIALS"
	ErrorCodeAuthInvalidToken           ErrorCode = "AUTH_INVALID_TOKEN"
	ErrorCodeAuthRefreshTokenInvalid    ErrorCode = "AUTH_REFRESH_TOKEN_INVALID"
	ErrorCodeAuthSessionExpired         ErrorCode = "AUTH_SESSION_EXPIRED"
	ErrorCodeAuthSessionRevoked         ErrorCode = "AUTH_SESSION_REVOKED"
	ErrorCodeAuthPasswordRequired       ErrorCode = "AUTH_PASSWORD_REQUIRED"
	ErrorCodeAuthPasswordTooWeak        ErrorCode = "AUTH_PASSWORD_TOO_WEAK"
	ErrorCodeUserDisabled               ErrorCode = "USER_DISABLED"
	ErrorCodeUserNotFound               ErrorCode = "USER_NOT_FOUND"
	ErrorCodeUserAlreadyExists          ErrorCode = "USER_ALREADY_EXISTS"
	ErrorCodeUserCannotDisableSelf      ErrorCode = "USER_CANNOT_DISABLE_SELF"
	ErrorCodeUserCannotDeleteSelf       ErrorCode = "USER_CANNOT_DELETE_SELF"
	ErrorCodeUserLastAdminRequired      ErrorCode = "USER_LAST_ADMIN_REQUIRED"
	ErrorCodeRoleNotFound               ErrorCode = "ROLE_NOT_FOUND"
	ErrorCodeRoleAlreadyExists          ErrorCode = "ROLE_ALREADY_EXISTS"
	ErrorCodeRoleInUse                  ErrorCode = "ROLE_IN_USE"
	ErrorCodeRoleSystemRoleLocked       ErrorCode = "ROLE_SYSTEM_ROLE_LOCKED"
	ErrorCodeRolePolicyInvalid          ErrorCode = "ROLE_POLICY_INVALID"
	ErrorCodeDeviceNotFound             ErrorCode = "DEVICE_NOT_FOUND"
	ErrorCodeDeviceNotTrusted           ErrorCode = "DEVICE_NOT_TRUSTED"
	ErrorCodeDeviceDisabled             ErrorCode = "DEVICE_DISABLED"
	ErrorCodeDeviceKeyInvalid           ErrorCode = "DEVICE_KEY_INVALID"
	ErrorCodeDeviceChallengeExpired     ErrorCode = "DEVICE_CHALLENGE_EXPIRED"
	ErrorCodeDeviceAlreadyBound         ErrorCode = "DEVICE_ALREADY_BOUND"
	ErrorCodeDeviceLimitExceeded        ErrorCode = "DEVICE_LIMIT_EXCEEDED"
	ErrorCodeServiceNotFound            ErrorCode = "SERVICE_NOT_FOUND"
	ErrorCodeServiceDisabled            ErrorCode = "SERVICE_DISABLED"
	ErrorCodeServiceAlreadyExists       ErrorCode = "SERVICE_ALREADY_EXISTS"
	ErrorCodeServiceUpstreamInvalid     ErrorCode = "SERVICE_UPSTREAM_INVALID"
	ErrorCodeServiceUpstreamUnreachable ErrorCode = "SERVICE_UPSTREAM_UNREACHABLE"
	ErrorCodeServiceRouteInvalid        ErrorCode = "SERVICE_ROUTE_INVALID"
	ErrorCodePolicyAccessDenied         ErrorCode = "POLICY_ACCESS_DENIED"
	ErrorCodePolicyUserDenied           ErrorCode = "POLICY_USER_DENIED"
	ErrorCodePolicyRoleDenied           ErrorCode = "POLICY_ROLE_DENIED"
	ErrorCodePolicyDeviceDenied         ErrorCode = "POLICY_DEVICE_DENIED"
	ErrorCodePolicyRuleInvalid          ErrorCode = "POLICY_RULE_INVALID"
	ErrorCodeAuditEventNotFound         ErrorCode = "AUDIT_EVENT_NOT_FOUND"
	ErrorCodeAuditQueryInvalid          ErrorCode = "AUDIT_QUERY_INVALID"
	ErrorCodeAuditExportTooLarge        ErrorCode = "AUDIT_EXPORT_TOO_LARGE"
	ErrorCodeGatewayBadUpstream         ErrorCode = "GATEWAY_BAD_UPSTREAM"
	ErrorCodeGatewayUpstreamTimeout     ErrorCode = "GATEWAY_UPSTREAM_TIMEOUT"
	ErrorCodeGatewayRouteNotFound       ErrorCode = "GATEWAY_ROUTE_NOT_FOUND"
	ErrorCodeGatewayRequestTooLarge     ErrorCode = "GATEWAY_REQUEST_TOO_LARGE"
)

type AuditEventType string

const (
	AuditEventTypeAuthLoginSucceeded         AuditEventType = "auth.login.succeeded"
	AuditEventTypeAuthLoginFailed            AuditEventType = "auth.login.failed"
	AuditEventTypeAuthLogoutSucceeded        AuditEventType = "auth.logout.succeeded"
	AuditEventTypeAuthRefreshSucceeded       AuditEventType = "auth.refresh.succeeded"
	AuditEventTypeDeviceRegistered           AuditEventType = "device.registered"
	AuditEventTypeDeviceChallengeRequested   AuditEventType = "device.challenge.requested"
	AuditEventTypeDeviceChallengeVerified    AuditEventType = "device.challenge.verified"
	AuditEventTypeDeviceDisabled             AuditEventType = "device.disabled"
	AuditEventTypeServiceAccessGranted       AuditEventType = "service.access.granted"
	AuditEventTypeServiceAccessDenied        AuditEventType = "service.access.denied"
	AuditEventTypeServiceAccessUpstreamError AuditEventType = "service.access.upstream_error"
	AuditEventTypeAdminUserCreated           AuditEventType = "admin.user.created"
	AuditEventTypeAdminUserUpdated           AuditEventType = "admin.user.updated"
	AuditEventTypeAdminRoleUpdated           AuditEventType = "admin.role.updated"
	AuditEventTypeAdminServiceUpdated        AuditEventType = "admin.service.updated"
)

type Pagination struct {
	Page       int64 `json:"page"`
	PageSize   int64 `json:"pageSize"`
	Total      int64 `json:"total"`
	TotalPages int64 `json:"totalPages"`
}

type Meta struct {
	RequestID  string      `json:"requestId"`
	Timestamp  string      `json:"timestamp"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

type APIError struct {
	Code        ErrorCode      `json:"code"`
	Message     string         `json:"message"`
	UserMessage string         `json:"userMessage"`
	Details     map[string]any `json:"details"`
}
