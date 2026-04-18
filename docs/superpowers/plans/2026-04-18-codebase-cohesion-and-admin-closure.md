# Codebase Cohesion And Admin Closure Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在不破坏现有网关闭环的前提下，继续压缩大文件职责边界，并补齐后台管理端缺失的用户、服务、设备管理闭环。

**Architecture:** 先处理已经明显过大的 Go 鉴权与代理相关文件，把“会一起变更”的逻辑拆到更聚合的职责文件中；随后只补与 Phase 6 直接相关的后台 API 和 UI，不提前扩展到桌面客户端。后台前端继续沿用 entity API + feature 组件 + page 编排的结构，后端继续沿用 auth service + server handler 分层。

**Tech Stack:** Go 1.26.2、PostgreSQL 18.x、React 19.2.5、TypeScript 6.0.3、TanStack Query 5、Vite 8.0.8、Tailwind CSS v4、shadcn/ui、pnpm、Vitest。

---

### Task 1: Shrink Oversized Gateway Auth Files

**Files:**
- Modify: `apps/gateway/internal/auth/service.go`
- Modify: `apps/gateway/internal/auth/service_auth.go`
- Modify: `apps/gateway/internal/auth/auth.go`
- Modify: `apps/gateway/internal/server/proxy_test.go`
- Create: `apps/gateway/internal/auth/service_core.go`
- Create: `apps/gateway/internal/auth/service_sessions.go`
- Create: `apps/gateway/internal/auth/passwords.go`
- Create: `apps/gateway/internal/auth/tokens.go`
- Create: `apps/gateway/internal/server/proxy_access_test.go`
- Create: `apps/gateway/internal/server/proxy_error_test.go`
- Test: `apps/gateway/internal/auth/service_auth_test.go`
- Test: `apps/gateway/internal/server/proxy_test.go`

- [x] **Step 1: Confirm existing seam tests cover split safety**

```go
func TestServiceIssueRefreshTokenPersistsSession(t *testing.T) {
	service, cleanup := newTestService(t)
	defer cleanup()

	result, err := service.issueRefreshToken(context.Background(), issueRefreshTokenInput{
		SessionID: "session-123",
		UserID:    adminUserID,
	})
	require.NoError(t, err)
	require.NotEmpty(t, result.RawToken)
}
```

- [x] **Step 2: Run targeted tests before refactor**

Run: `BIFROST_DATABASE_TEST_URL='postgres://bifrost:bifrost@127.0.0.1:15432/postgres?sslmode=disable' GOTOOLCHAIN=local go test ./internal/auth ./internal/server -run 'TestServiceIssueRefreshTokenPersistsSession|TestHandleServiceProxy'`

Expected: PASS，作为拆分前基线。

- [x] **Step 3: Move constructors, clock/id helpers, refresh-session persistence, password helpers, token helpers into focused files**

```go
// service_core.go
func NewService(options ServiceOptions) Service { ... }
func (s Service) now() time.Time { ... }
func (s Service) newSessionID() (string, error) { ... }

// service_sessions.go
type issueRefreshTokenInput struct { ... }
func (s Service) issueRefreshToken(ctx context.Context, input issueRefreshTokenInput) (issueRefreshTokenResult, error) { ... }
func (s Service) persistRefreshSession(ctx context.Context, input persistRefreshSessionInput) error { ... }

// passwords.go
func HashPassword(password string) (string, error) { ... }
func VerifyPassword(hash string, password string) error { ... }

// tokens.go
func IssueAccessToken(input IssueAccessTokenInput) (string, error) { ... }
func ParseAccessToken(token string) (AccessTokenClaims, error) { ... }
```

- [x] **Step 4: Split proxy tests by intent and keep original assertions unchanged**

```go
// proxy_access_test.go
func TestHandleServiceProxyForwardsAuthorizedRequest(t *testing.T) { ... }

// proxy_error_test.go
func TestHandleServiceProxyReturnsBadGatewayForUpstreamFailure(t *testing.T) { ... }
func TestHandleServiceProxyReturnsGatewayTimeoutForTimeout(t *testing.T) { ... }
```

- [x] **Step 5: Run focused refactor verification**

Run: `BIFROST_DATABASE_TEST_URL='postgres://bifrost:bifrost@127.0.0.1:15432/postgres?sslmode=disable' GOTOOLCHAIN=local go test ./internal/auth ./internal/server`

Expected: PASS，无行为回归。

- [ ] **Step 6: Commit refactor batch**

```bash
git add apps/gateway/internal/auth apps/gateway/internal/server docs/08-roadmap/development-checklist.md docs/superpowers/plans/2026-04-18-codebase-cohesion-and-admin-closure.md
git commit -m "refactor: split gateway auth core and proxy tests"
```

### Task 2: Add Admin User Detail, Password Reset, And Status Control APIs

**Files:**
- Modify: `apps/gateway/internal/auth/service_admin_users.go`
- Modify: `apps/gateway/internal/auth/service_auth_test.go`
- Modify: `apps/gateway/internal/server/admin_user_handlers.go`
- Modify: `apps/gateway/internal/server/auth_routes_test.go`
- Modify: `apps/gateway/internal/server/server.go`
- Modify: `apps/gateway/internal/server/response.go`

- [x] **Step 1: Add failing service tests for user detail, reset password, and enable/disable**

```go
func TestGetAdminUserReturnsUserWithRoles(t *testing.T) { ... }
func TestResetAdminUserPasswordReplacesStoredHash(t *testing.T) { ... }
func TestSetAdminUserStatusDisablesUser(t *testing.T) { ... }
```

- [x] **Step 2: Run the new service tests and confirm they fail**

Run: `BIFROST_DATABASE_TEST_URL='postgres://bifrost:bifrost@127.0.0.1:15432/postgres?sslmode=disable' GOTOOLCHAIN=local go test ./internal/auth -run 'TestGetAdminUserReturnsUserWithRoles|TestResetAdminUserPasswordReplacesStoredHash|TestSetAdminUserStatusDisablesUser'`

Expected: FAIL，提示缺少对应方法或断言不满足。

- [x] **Step 3: Add narrow auth service methods and audit events**

```go
func (s Service) GetAdminUser(ctx context.Context, input GetAdminUserInput) (AdminUser, error) { ... }
func (s Service) ResetAdminUserPassword(ctx context.Context, input ResetAdminUserPasswordInput) error { ... }
func (s Service) SetAdminUserStatus(ctx context.Context, input SetAdminUserStatusInput) (AdminUser, error) { ... }
```

- [x] **Step 4: Expose REST endpoints in server handlers**

```go
// GET /api/v1/admin/users/{id}
// POST /api/v1/admin/users/{id}/reset-password
// POST /api/v1/admin/users/{id}/status
```

- [x] **Step 5: Add route tests for success and not-found paths**

```go
func TestHandleAdminUserByIDReturnsUser(t *testing.T) { ... }
func TestHandleAdminUserResetPasswordReturnsNoContent(t *testing.T) { ... }
func TestHandleAdminUserStatusReturnsUpdatedUser(t *testing.T) { ... }
```

- [x] **Step 6: Run gateway verification**

Run: `BIFROST_DATABASE_TEST_URL='postgres://bifrost:bifrost@127.0.0.1:15432/postgres?sslmode=disable' GOTOOLCHAIN=local go test ./internal/auth ./internal/server -run 'AdminUser|ResetPassword|Status'`

Expected: PASS。

- [ ] **Step 7: Commit user-management backend batch**

```bash
git add apps/gateway/internal/auth apps/gateway/internal/server docs/08-roadmap/development-checklist.md
git commit -m "feat: add admin user detail and control endpoints"
```

### Task 3: Add Admin Service And Device Detail/Status APIs

**Files:**
- Modify: `apps/gateway/internal/auth/service_admin_catalog.go`
- Modify: `apps/gateway/internal/auth/service_devices.go`
- Modify: `apps/gateway/internal/auth/service_admin_test.go`
- Modify: `apps/gateway/internal/auth/service_devices_test.go`
- Modify: `apps/gateway/internal/server/admin_catalog_handlers.go`
- Modify: `apps/gateway/internal/server/device_handlers.go`
- Modify: `apps/gateway/internal/server/admin_routes_test.go`
- Modify: `apps/gateway/internal/server/device_routes_test.go`
- Modify: `apps/gateway/internal/server/server.go`

- [x] **Step 1: Add failing tests for service/device detail and status updates**

```go
func TestGetAdminServiceReturnsCatalogEntry(t *testing.T) { ... }
func TestUpdateAdminServicePersistsEditableFields(t *testing.T) { ... }
func TestSetAdminServiceStatusDisablesService(t *testing.T) { ... }
func TestGetAdminDeviceReturnsTrustedDevice(t *testing.T) { ... }
func TestSetAdminDeviceStatusDisablesDevice(t *testing.T) { ... }
```

- [x] **Step 2: Run targeted backend tests**

Run: `BIFROST_DATABASE_TEST_URL='postgres://bifrost:bifrost@127.0.0.1:15432/postgres?sslmode=disable' GOTOOLCHAIN=local go test ./internal/auth ./internal/server -run 'AdminService|AdminDevice'`

Expected: FAIL。

- [x] **Step 3: Implement minimal service methods**

```go
func (s Service) GetAdminService(ctx context.Context, input GetAdminServiceInput) (AdminService, error) { ... }
func (s Service) UpdateAdminService(ctx context.Context, input UpdateAdminServiceInput) (AdminService, error) { ... }
func (s Service) SetAdminServiceStatus(ctx context.Context, input SetAdminServiceStatusInput) (AdminService, error) { ... }
func (s Service) GetAdminDevice(ctx context.Context, input GetAdminDeviceInput) (AdminDevice, error) { ... }
func (s Service) SetAdminDeviceStatus(ctx context.Context, input SetAdminDeviceStatusInput) (AdminDevice, error) { ... }
```

- [x] **Step 4: Wire server endpoints**

```go
// GET /api/v1/admin/services/{id}
// PATCH /api/v1/admin/services/{id}
// POST /api/v1/admin/services/{id}/status
// GET /api/v1/admin/devices/{id}
// POST /api/v1/admin/devices/{id}/status
```

- [x] **Step 5: Run backend verification**

Run: `BIFROST_DATABASE_TEST_URL='postgres://bifrost:bifrost@127.0.0.1:15432/postgres?sslmode=disable' GOTOOLCHAIN=local go test ./internal/auth ./internal/server`

Expected: PASS。

- [ ] **Step 6: Commit catalog/device backend batch**

```bash
git add apps/gateway/internal/auth apps/gateway/internal/server docs/08-roadmap/development-checklist.md
git commit -m "feat: add admin service and device control endpoints"
```

### Task 4: Finish Phase 6 Admin UI For Users, Services, And Devices

**Files:**
- Modify: `apps/admin/src/entities/admin/api.ts`
- Modify: `apps/admin/src/entities/admin/types.ts`
- Modify: `apps/admin/src/entities/admin/api.test.ts`
- Modify: `apps/admin/src/pages/users-page.tsx`
- Modify: `apps/admin/src/pages/services-page.tsx`
- Modify: `apps/admin/src/pages/devices-page.tsx`
- Modify: `apps/admin/src/features/admin-users/users-table.tsx`
- Modify: `apps/admin/src/features/admin-services/services-table.tsx`
- Create: `apps/admin/src/features/admin-users/user-detail-drawer.tsx`
- Create: `apps/admin/src/features/admin-users/reset-password-dialog.tsx`
- Create: `apps/admin/src/features/admin-services/edit-service-dialog.tsx`
- Create: `apps/admin/src/features/admin-devices/device-detail-drawer.tsx`
- Create: `apps/admin/src/features/admin-devices/device-status-action.tsx`

- [x] **Step 1: Extend API layer with exact response shapes**

```ts
export async function getAdminUser(...) { ... }
export async function resetAdminUserPassword(...) { ... }
export async function setAdminUserStatus(...) { ... }
export async function getAdminService(...) { ... }
export async function updateAdminService(...) { ... }
export async function setAdminServiceStatus(...) { ... }
export async function getAdminDevice(...) { ... }
export async function setAdminDeviceStatus(...) { ... }
```

- [x] **Step 2: Add failing API tests**

```ts
it("calls reset admin user password endpoint", async () => { ... });
it("calls update admin service endpoint", async () => { ... });
it("calls set admin device status endpoint", async () => { ... });
```

- [x] **Step 3: Build focused feature components with compact actions**

```tsx
<UserDetailDrawer userId={selectedUserId} />
<ResetPasswordDialog user={selectedUser} />
<EditServiceDialog service={selectedService} />
<DeviceDetailDrawer deviceId={selectedDeviceId} />
```

- [x] **Step 4: Keep page files orchestration-only**

```tsx
const [selectedUserId, setSelectedUserId] = useState<string | null>(null);
const [selectedService, setSelectedService] = useState<AdminService | null>(null);
const [selectedDeviceId, setSelectedDeviceId] = useState<string | null>(null);
```

- [x] **Step 5: Run frontend verification**

Run: `pnpm --filter @bifrost/admin test`

Expected: PASS。

Run: `pnpm lint && pnpm check`

Expected: PASS。

- [ ] **Step 6: Update checklist timestamps for completed Phase 6 items**

```md
- [x] 实现用户详情 Drawer（完成时间：2026-04-18 HH:mm CST）
- [x] 实现重置密码流程（完成时间：2026-04-18 HH:mm CST）
- [x] 实现用户启用 / 禁用（完成时间：2026-04-18 HH:mm CST）
- [x] 实现设备详情 Drawer（完成时间：2026-04-18 HH:mm CST）
- [x] 实现设备启用 / 禁用（完成时间：2026-04-18 HH:mm CST）
- [x] 实现服务创建与编辑（完成时间：2026-04-18 HH:mm CST）
- [x] 实现服务启用 / 禁用（完成时间：2026-04-18 HH:mm CST）
```

- [ ] **Step 7: Commit admin UI batch**

```bash
git add apps/admin docs/08-roadmap/development-checklist.md
git commit -m "feat: complete admin control drawers and status actions"
```

### Task 5: Full Verification And Push

**Files:**
- Modify: `docs/08-roadmap/development-checklist.md`

- [ ] **Step 1: Run full repo verification**

Run: `BIFROST_DATABASE_TEST_URL='postgres://bifrost:bifrost@127.0.0.1:15432/postgres?sslmode=disable' GOTOOLCHAIN=local go test ./...`

Expected: PASS in `apps/gateway`。

Run: `pnpm lint && pnpm check && pnpm test`

Expected: PASS at repo root。

- [ ] **Step 2: Push to dev**

```bash
git push origin dev
```

- [ ] **Step 3: Confirm clean worktree**

Run: `git status --short --branch`

Expected: `## dev...origin/dev`

## Self-Review

- Spec coverage: 本计划覆盖了用户详情、重置密码、用户启停、服务编辑、服务启停、设备详情、设备启停，以及继续拆分过大的网关文件；尚未覆盖角色编辑、用户级服务覆盖 UI、桌面客户端、Phase 8 完整自动化测试，这些保持在后续计划中处理。
- Placeholder scan: 已移除 `TODO/TBD` 类占位描述；未使用“类似 Task N”式引用。
- Type consistency: 计划内统一使用 `GetAdminUser`、`ResetAdminUserPassword`、`SetAdminUserStatus`、`GetAdminService`、`UpdateAdminService`、`SetAdminServiceStatus`、`GetAdminDevice`、`SetAdminDeviceStatus` 作为后端与前端 API 命名。

Plan complete and saved to `docs/superpowers/plans/2026-04-18-codebase-cohesion-and-admin-closure.md`. Two execution options:

1. Subagent-Driven (recommended) - I dispatch a fresh subagent per task, review between tasks, fast iteration
2. Inline Execution - Execute tasks in this session using executing-plans, batch execution with checkpoints

默认按用户已确认的 `Inline Execution` 继续执行。
