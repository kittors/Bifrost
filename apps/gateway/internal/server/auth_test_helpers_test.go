package server_test

import (
	"context"
	"encoding/json"

	"github.com/kittors/bifrost/apps/gateway/internal/auth"
)

// 共享 stub 与响应解析结构统一放在 helper 文件，避免每个路由测试重复定义。

type stubAuthService struct {
	adminLoginInput  auth.AdminLoginInput
	adminLoginResult auth.LoginResult
	adminLoginError  error

	clientLoginInput  auth.ClientLoginInput
	clientLoginResult auth.LoginResult
	clientLoginError  error

	refreshInput  auth.RefreshInput
	refreshResult auth.LoginResult
	refreshError  error

	logoutInput auth.LogoutInput
	logoutError error

	currentUserInput auth.CurrentUserInput
	currentUser      auth.LoginUser
	currentUserError error

	registerDeviceInput  auth.RegisterDeviceInput
	registerDeviceResult auth.DeviceResult
	registerDeviceError  error

	deviceChallengeInput  auth.CreateDeviceChallengeInput
	deviceChallengeResult auth.DeviceChallengeResult
	deviceChallengeError  error

	verifyDeviceChallengeInput  auth.VerifyDeviceChallengeInput
	verifyDeviceChallengeResult auth.DeviceChallengeVerificationResult
	verifyDeviceChallengeError  error

	listClientServicesInput auth.ListClientServicesInput
	clientServices          []auth.ClientService
	clientServicesError     error

	getClientServiceInput auth.GetClientServiceInput
	clientService         auth.ClientService
	clientServiceError    error

	createServiceAccessURLInput auth.CreateServiceAccessURLInput
	serviceAccessURL            auth.ServiceAccessURLResult
	serviceAccessURLError       error

	resolveProxyInput  auth.ResolveProxyRequestInput
	resolveProxyResult auth.ResolveProxyRequestResult
	resolveProxyError  error

	recordProxyAccessEventInput auth.RecordProxyAccessEventInput
	recordProxyAccessEventError error

	listAdminUsersInput auth.ListAdminUsersInput
	adminUsers          auth.AdminUserListResult
	adminUsersError     error

	createAdminUserInput auth.CreateAdminUserInput
	createdAdminUser     auth.AdminUser
	createAdminUserError error

	updateAdminUserInput auth.UpdateAdminUserInput
	updatedAdminUser     auth.AdminUser
	updateAdminUserError error

	getAdminUserInput auth.GetAdminUserInput
	adminUser         auth.AdminUser
	getAdminUserError error

	resetAdminUserPasswordInput auth.ResetAdminUserPasswordInput
	resetAdminUserPasswordError error

	setAdminUserStatusInput auth.SetAdminUserStatusInput
	statusAdminUser         auth.AdminUser
	setAdminUserStatusError error

	listAdminRolesInput auth.ListAdminRolesInput
	adminRoles          auth.AdminRoleListResult
	adminRolesError     error

	createAdminRoleInput auth.CreateAdminRoleInput
	createdAdminRole     auth.AdminRole
	createAdminRoleError error

	listAdminServicesInput auth.ListAdminServicesInput
	adminServices          auth.AdminServiceListResult
	adminServicesError     error

	createAdminServiceInput auth.CreateAdminServiceInput
	createdAdminService     auth.AdminService
	createAdminServiceError error

	listAdminDevicesInput auth.ListAdminDevicesInput
	adminDevices          auth.AdminDeviceListResult
	adminDevicesError     error

	listAdminAuditEventsInput auth.ListAdminAuditEventsInput
	adminAuditEvents          auth.AdminAuditEventListResult
	adminAuditEventsError     error

	replaceRoleServicesInput auth.ReplaceRoleServicesInput
	replaceRoleServicesError error

	replaceUserServiceOverridesInput auth.ReplaceUserServiceOverridesInput
	userServiceOverrides             []auth.UserServiceOverride
	userServiceOverridesError        error
}

func (s *stubAuthService) AdminLogin(_ context.Context, input auth.AdminLoginInput) (auth.LoginResult, error) {
	s.adminLoginInput = input
	return s.adminLoginResult, s.adminLoginError
}

func (s *stubAuthService) ClientLogin(_ context.Context, input auth.ClientLoginInput) (auth.LoginResult, error) {
	s.clientLoginInput = input
	return s.clientLoginResult, s.clientLoginError
}

func (s *stubAuthService) RefreshSession(_ context.Context, input auth.RefreshInput) (auth.LoginResult, error) {
	s.refreshInput = input
	return s.refreshResult, s.refreshError
}

func (s *stubAuthService) Logout(_ context.Context, input auth.LogoutInput) error {
	s.logoutInput = input
	return s.logoutError
}

func (s *stubAuthService) CurrentUser(_ context.Context, input auth.CurrentUserInput) (auth.LoginUser, error) {
	s.currentUserInput = input
	return s.currentUser, s.currentUserError
}

func (s *stubAuthService) RegisterDevice(_ context.Context, input auth.RegisterDeviceInput) (auth.DeviceResult, error) {
	s.registerDeviceInput = input
	return s.registerDeviceResult, s.registerDeviceError
}

func (s *stubAuthService) CreateDeviceChallenge(_ context.Context, input auth.CreateDeviceChallengeInput) (auth.DeviceChallengeResult, error) {
	s.deviceChallengeInput = input
	return s.deviceChallengeResult, s.deviceChallengeError
}

func (s *stubAuthService) VerifyDeviceChallenge(_ context.Context, input auth.VerifyDeviceChallengeInput) (auth.DeviceChallengeVerificationResult, error) {
	s.verifyDeviceChallengeInput = input
	return s.verifyDeviceChallengeResult, s.verifyDeviceChallengeError
}

func (s *stubAuthService) ListClientServices(_ context.Context, input auth.ListClientServicesInput) ([]auth.ClientService, error) {
	s.listClientServicesInput = input
	return s.clientServices, s.clientServicesError
}

func (s *stubAuthService) GetClientService(_ context.Context, input auth.GetClientServiceInput) (auth.ClientService, error) {
	s.getClientServiceInput = input
	return s.clientService, s.clientServiceError
}

func (s *stubAuthService) CreateServiceAccessURL(_ context.Context, input auth.CreateServiceAccessURLInput) (auth.ServiceAccessURLResult, error) {
	s.createServiceAccessURLInput = input
	return s.serviceAccessURL, s.serviceAccessURLError
}

func (s *stubAuthService) ResolveProxyRequest(_ context.Context, input auth.ResolveProxyRequestInput) (auth.ResolveProxyRequestResult, error) {
	s.resolveProxyInput = input
	return s.resolveProxyResult, s.resolveProxyError
}

func (s *stubAuthService) RecordProxyAccessEvent(_ context.Context, input auth.RecordProxyAccessEventInput) error {
	s.recordProxyAccessEventInput = input
	return s.recordProxyAccessEventError
}

func (s *stubAuthService) ListAdminUsers(_ context.Context, input auth.ListAdminUsersInput) (auth.AdminUserListResult, error) {
	s.listAdminUsersInput = input
	return s.adminUsers, s.adminUsersError
}

func (s *stubAuthService) CreateAdminUser(_ context.Context, input auth.CreateAdminUserInput) (auth.AdminUser, error) {
	s.createAdminUserInput = input
	return s.createdAdminUser, s.createAdminUserError
}

func (s *stubAuthService) UpdateAdminUser(_ context.Context, input auth.UpdateAdminUserInput) (auth.AdminUser, error) {
	s.updateAdminUserInput = input
	return s.updatedAdminUser, s.updateAdminUserError
}

func (s *stubAuthService) GetAdminUser(_ context.Context, input auth.GetAdminUserInput) (auth.AdminUser, error) {
	s.getAdminUserInput = input
	return s.adminUser, s.getAdminUserError
}

func (s *stubAuthService) ResetAdminUserPassword(_ context.Context, input auth.ResetAdminUserPasswordInput) error {
	s.resetAdminUserPasswordInput = input
	return s.resetAdminUserPasswordError
}

func (s *stubAuthService) SetAdminUserStatus(_ context.Context, input auth.SetAdminUserStatusInput) (auth.AdminUser, error) {
	s.setAdminUserStatusInput = input
	return s.statusAdminUser, s.setAdminUserStatusError
}

func (s *stubAuthService) ListAdminRoles(_ context.Context, input auth.ListAdminRolesInput) (auth.AdminRoleListResult, error) {
	s.listAdminRolesInput = input
	return s.adminRoles, s.adminRolesError
}

func (s *stubAuthService) CreateAdminRole(_ context.Context, input auth.CreateAdminRoleInput) (auth.AdminRole, error) {
	s.createAdminRoleInput = input
	return s.createdAdminRole, s.createAdminRoleError
}

func (s *stubAuthService) ListAdminServices(_ context.Context, input auth.ListAdminServicesInput) (auth.AdminServiceListResult, error) {
	s.listAdminServicesInput = input
	return s.adminServices, s.adminServicesError
}

func (s *stubAuthService) CreateAdminService(_ context.Context, input auth.CreateAdminServiceInput) (auth.AdminService, error) {
	s.createAdminServiceInput = input
	return s.createdAdminService, s.createAdminServiceError
}

func (s *stubAuthService) ListAdminDevices(_ context.Context, input auth.ListAdminDevicesInput) (auth.AdminDeviceListResult, error) {
	s.listAdminDevicesInput = input
	return s.adminDevices, s.adminDevicesError
}

func (s *stubAuthService) ListAdminAuditEvents(_ context.Context, input auth.ListAdminAuditEventsInput) (auth.AdminAuditEventListResult, error) {
	s.listAdminAuditEventsInput = input
	return s.adminAuditEvents, s.adminAuditEventsError
}

func (s *stubAuthService) ReplaceRoleServices(_ context.Context, input auth.ReplaceRoleServicesInput) error {
	s.replaceRoleServicesInput = input
	return s.replaceRoleServicesError
}

func (s *stubAuthService) ReplaceUserServiceOverrides(_ context.Context, input auth.ReplaceUserServiceOverridesInput) ([]auth.UserServiceOverride, error) {
	s.replaceUserServiceOverridesInput = input
	return s.userServiceOverrides, s.userServiceOverridesError
}

type apiEnvelope struct {
	Success bool              `json:"success"`
	Data    json.RawMessage   `json:"data"`
	Meta    envelopeMeta      `json:"meta"`
	Error   *envelopeAPIError `json:"error"`
}

type loginResponse struct {
	AccessToken  string         `json:"accessToken"`
	RefreshToken string         `json:"refreshToken"`
	ExpiresIn    int            `json:"expiresIn"`
	User         loginUserShape `json:"user"`
}

type loginUserShape struct {
	ID          string   `json:"id"`
	Username    string   `json:"username"`
	DisplayName string   `json:"displayName"`
	Roles       []string `json:"roles"`
}

type envelopeMeta struct {
	RequestID string `json:"requestId"`
	Timestamp string `json:"timestamp"`
}

type envelopeAPIError struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	UserMessage string `json:"userMessage"`
}
