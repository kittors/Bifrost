# Admin Phase 6 Closure Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 补齐后台管理端 Phase 6 剩余功能，让管理员能编辑角色、读取并编辑用户级服务覆盖，并验证主题与关键筛选链路。

**Architecture:** 后端继续沿用 `auth.Service` 作为业务边界、`server.App` 作为 HTTP handler 边界；前端继续沿用 `entities/admin/api.ts` 做接口封装、`features/admin-*` 做高聚合 UI 组件、`pages/*` 只做查询和区块编排。所有新增行为先写失败测试，再实现最小代码通过测试。

**Tech Stack:** Go 1.26.2、PostgreSQL 18.x、React 19.2.5、TypeScript 6.0.3、TanStack Query 5.99.0、Vite 8.0.8、Tailwind CSS v4、Vitest 4.1.4、Biome。

---

### Task 1: Add Role Edit Backend APIs

**Files:**
- Modify: `apps/gateway/internal/auth/service_inputs.go`
- Modify: `apps/gateway/internal/auth/service_admin_catalog.go`
- Modify: `apps/gateway/internal/auth/service_admin_test.go`
- Modify: `apps/gateway/internal/server/server.go`
- Modify: `apps/gateway/internal/server/request_helpers.go`
- Modify: `apps/gateway/internal/server/admin_catalog_handlers.go`
- Modify: `apps/gateway/internal/server/auth_test_helpers_test.go`
- Modify: `apps/gateway/internal/server/admin_routes_test.go`

- [x] **Step 1: Add failing tests**

```go
func TestServiceAdminRoleUpdate(t *testing.T) {
  updated, err := service.UpdateAdminRole(ctx, auth.UpdateAdminRoleInput{
    AccessToken: loginResult.AccessToken,
    RoleID: "role_developer",
    DisplayName: "研发团队",
    Description: "研发私有服务访问角色",
  })
  if err != nil { t.Fatalf("update admin role: %v", err) }
  if updated.DisplayName != "研发团队" { t.Fatalf("unexpected role: %#v", updated) }
}
```

Run: `BIFROST_DATABASE_TEST_URL='postgres://bifrost:bifrost@127.0.0.1:15432/postgres?sslmode=disable' GOTOOLCHAIN=local go test ./internal/auth ./internal/server -run 'AdminRole|AdminConfigRoutes'`

Expected: FAIL because `UpdateAdminRole` and route wiring do not exist.

- [x] **Step 2: Implement input type and auth service method**

```go
type UpdateAdminRoleInput struct {
  AccessToken string
  RoleID string
  DisplayName string
  Description string
}

func (s Service) UpdateAdminRole(ctx context.Context, input UpdateAdminRoleInput) (AdminRole, error) { ... }
```

- [x] **Step 3: Wire HTTP route**

```text
PATCH /api/v1/admin/roles/{roleId}
```

- [x] **Step 4: Run backend tests**

Run: `BIFROST_DATABASE_TEST_URL='postgres://bifrost:bifrost@127.0.0.1:15432/postgres?sslmode=disable' GOTOOLCHAIN=local go test ./internal/auth ./internal/server`

Expected: PASS。

### Task 2: Add User Service Overrides Read API

**Files:**
- Modify: `apps/gateway/internal/auth/service_inputs.go`
- Modify: `apps/gateway/internal/auth/service_admin_policy.go`
- Modify: `apps/gateway/internal/auth/service_admin_test.go`
- Modify: `apps/gateway/internal/server/admin_user_handlers.go`
- Modify: `apps/gateway/internal/server/auth_test_helpers_test.go`
- Modify: `apps/gateway/internal/server/admin_routes_test.go`

- [x] **Step 1: Add failing tests**

```go
func TestServiceListUserServiceOverrides(t *testing.T) {
  overrides, err := service.ListUserServiceOverrides(ctx, auth.ListUserServiceOverridesInput{
    AccessToken: loginResult.AccessToken,
    UserID: "user_alice",
  })
  if err != nil { t.Fatalf("list user service overrides: %v", err) }
  if len(overrides) != 2 { t.Fatalf("expected 2 overrides, got %d", len(overrides)) }
}
```

Run: `BIFROST_DATABASE_TEST_URL='postgres://bifrost:bifrost@127.0.0.1:15432/postgres?sslmode=disable' GOTOOLCHAIN=local go test ./internal/auth ./internal/server -run 'UserServiceOverrides|AdminConfigRoutes'`

Expected: FAIL because GET service override support does not exist.

- [x] **Step 2: Implement service read method**

```go
type ListUserServiceOverridesInput struct {
  AccessToken string
  UserID string
}

func (s Service) ListUserServiceOverrides(ctx context.Context, input ListUserServiceOverridesInput) ([]UserServiceOverride, error) { ... }
```

- [x] **Step 3: Wire HTTP route**

```text
GET /api/v1/admin/users/{userId}/service-overrides
```

- [x] **Step 4: Run backend tests**

Run: `BIFROST_DATABASE_TEST_URL='postgres://bifrost:bifrost@127.0.0.1:15432/postgres?sslmode=disable' GOTOOLCHAIN=local go test ./internal/auth ./internal/server`

Expected: PASS。

### Task 3: Finish Admin Role Edit And User Override UI

**Files:**
- Modify: `apps/admin/src/entities/admin/types.ts`
- Modify: `apps/admin/src/entities/admin/api.ts`
- Modify: `apps/admin/src/entities/admin/api.test.ts`
- Modify: `apps/admin/src/features/admin-roles/roles-table.tsx`
- Modify: `apps/admin/src/pages/roles-page.tsx`
- Modify: `apps/admin/src/features/admin-users/users-table.tsx`
- Modify: `apps/admin/src/pages/users-page.tsx`
- Create: `apps/admin/src/features/admin-roles/edit-role-dialog.tsx`
- Create: `apps/admin/src/features/admin-users/user-service-overrides-drawer.tsx`

- [x] **Step 1: Add failing API tests**

```ts
it("updates an admin role through the gateway API", async () => { ... });
it("lists user service overrides through the gateway API", async () => { ... });
it("replaces user service overrides through the gateway API", async () => { ... });
```

Run: `pnpm --filter @bifrost/admin test`

Expected: FAIL until API helpers exist.

- [x] **Step 2: Extend API helpers**

```ts
export async function updateAdminRole(input: { accessToken: string; roleID: string; displayName: string; description: string }) { ... }
export async function listUserServiceOverrides(input: { accessToken: string; userID: string }) { ... }
export async function replaceUserServiceOverrides(input: { accessToken: string; userID: string; allowServiceIDs: string[]; denyServiceIDs: string[] }) { ... }
```

- [x] **Step 3: Add focused UI components**

```tsx
<EditRoleDialog role={editingRole} />
<UserServiceOverridesDrawer user={overrideUser} services={services} />
```

- [x] **Step 4: Run frontend tests and checks**

Run: `pnpm --filter @bifrost/admin test && pnpm --filter @bifrost/admin check && pnpm lint`

Expected: PASS。

### Task 4: Update Checklist, Verify, Commit, Push

**Files:**
- Modify: `docs/08-roadmap/development-checklist.md`
- Modify: `docs/superpowers/plans/2026-04-18-admin-phase6-closure.md`

- [ ] **Step 1: Mark Phase 6 completion items**

当前仅已勾选以下真实完成项：

```md
- [x] 实现角色创建与编辑（完成时间：2026-04-18 11:17 CST）
- [x] 实现用户级服务覆盖编辑界面（完成时间：2026-04-18 11:17 CST）
```

`Light / Dark` 主题切换、关键页面分页与筛选行为，以及 `M7` 完成标记仍待浏览器级联调验证后再更新。

```md
- [x] 实现角色创建与编辑（完成时间：2026-04-18 HH:mm CST）
- [x] 实现用户级服务覆盖编辑界面（完成时间：2026-04-18 HH:mm CST）
- [x] 验证 Light / Dark 主题切换（完成时间：2026-04-18 HH:mm CST）
- [x] 验证所有关键页面分页与筛选行为（完成时间：2026-04-18 HH:mm CST）
- [x] M7：后台配置闭环完成（完成时间：2026-04-18 HH:mm CST）
```

- [x] **Step 2: Run full verification**

Run: `BIFROST_DATABASE_TEST_URL='postgres://bifrost:bifrost@127.0.0.1:15432/postgres?sslmode=disable' GOTOOLCHAIN=local go test ./...`

Run: `pnpm lint && pnpm check && pnpm test`

Expected: PASS。

- [ ] **Step 3: Commit and push**

```bash
git add apps/gateway apps/admin docs/08-roadmap/development-checklist.md docs/superpowers/plans/2026-04-18-admin-phase6-closure.md
git commit -m "feat: close admin phase six controls"
git push origin dev
```

## Self-Review

- Spec coverage: 覆盖 Phase 6 当前剩余的角色编辑、用户级服务覆盖 UI、主题验证和筛选验证；不覆盖 Phase 7 桌面客户端，下一计划单独处理。
- Placeholder scan: 无 `TBD`、`TODO` 或“以后补”占位。
- Type consistency: 后端使用 `UpdateAdminRoleInput`、`ListUserServiceOverridesInput`；前端使用 `updateAdminRole`、`listUserServiceOverrides`、`replaceUserServiceOverrides`。
