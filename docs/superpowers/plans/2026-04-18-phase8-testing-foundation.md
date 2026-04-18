# Phase 8 Testing Foundation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 建立可重复执行的测试基础设施，让 Bifrost 可以用 Docker 驱动真实联调环境，稳定执行 Go、前端与 Playwright E2E 测试。

**Architecture:** 根目录新增统一测试编排层，负责启动/等待/清理 Docker 联调环境、迁移数据库并回填种子数据；Playwright 以 root 级配置驱动 Admin Web 与 Gateway 真服务；前端 Vitest 继续保留应用内测试，但补足共享测试配置与关键页面回归。服务端沿用 Go `testing`，优先复用现有 `internal/auth` 与 `internal/server` 的真实逻辑，避免重复搭建第二套测试世界。

**Tech Stack:** Playwright 1.59.1、Vitest 4.1.4、Node.js 24.x、pnpm 10.33.0、Go 1.26.2、PostgreSQL 18.x、Docker Compose v2。

---

### Task 1: Add Root E2E Infrastructure Scripts

**Files:**
- Create: `scripts/testing/e2e-up.mjs`
- Create: `scripts/testing/e2e-down.mjs`
- Create: `scripts/testing/e2e-seed.mjs`
- Create: `scripts/testing/wait-for-http.mjs`
- Modify: `package.json`
- Modify: `docker-compose.yml`
- Modify: `docs/06-engineering/testing-strategy.md`

- [x] **Step 1: Write the failing infra smoke test**

```js
test("root e2e scripts are declared and point to dedicated helpers", () => {
  const packageJson = JSON.parse(readFileSync(new URL("../../package.json", import.meta.url), "utf8"));

  assert.equal(packageJson.scripts["test:e2e:up"], "node ./scripts/testing/e2e-up.mjs");
  assert.equal(packageJson.scripts["test:e2e:seed"], "node ./scripts/testing/e2e-seed.mjs");
  assert.equal(packageJson.scripts["test:e2e:down"], "node ./scripts/testing/e2e-down.mjs");
  assert.equal(packageJson.scripts["test:e2e"], "playwright test");
});
```

Run: `node --test tests/infra/*.test.mjs`

Expected: FAIL because these scripts and helpers do not exist yet.

- [x] **Step 2: Implement dedicated Docker orchestration helpers**

```js
await execFile("docker", ["compose", "up", "-d", "postgres", "mock-gitlab", "mock-jenkins", "mock-docs", "mock-internal-admin"]);
await execFile("pnpm", ["db:migrate"], { env: { ...process.env, BIFROST_DATABASE_URL: testDatabaseURL } });
await execFile("pnpm", ["db:seed"], { env: { ...process.env, BIFROST_DATABASE_URL: testDatabaseURL } });
await execFile("docker", ["compose", "up", "-d", "gateway", "admin-web"]);
```

- [x] **Step 3: Add health waiting and teardown**

```js
await waitForHTTP("http://127.0.0.1:8080/readyz");
await waitForHTTP("http://127.0.0.1:5173/health");
await execFile("docker", ["compose", "down", "-v", "--remove-orphans"]);
```

- [x] **Step 4: Run infra verification**

Run: `node --test tests/infra/*.test.mjs`

Expected: PASS。

### Task 2: Add Playwright Root Config And Shared Fixtures

**Files:**
- Create: `playwright.config.ts`
- Create: `tests/e2e/fixtures/env.ts`
- Create: `tests/e2e/fixtures/admin-session.ts`
- Create: `tests/e2e/fixtures/client-api.ts`
- Modify: `package.json`
- Modify: `pnpm-lock.yaml`

- [ ] **Step 1: Write the failing config smoke test**

```ts
test("playwright config uses local admin and gateway base urls", async () => {
  const config = (await import("../../playwright.config.ts")).default;
  assert.equal(config.use?.baseURL, "http://127.0.0.1:5173");
  assert.equal(config.webServer, undefined);
});
```

Run: `node --test tests/infra/*.test.mjs`

Expected: FAIL because root Playwright config and fixtures are missing.

- [x] **Step 2: Add Playwright dependency and root config**

```ts
export default defineConfig({
  testDir: "./tests/e2e",
  timeout: 45_000,
  use: {
    baseURL: "http://127.0.0.1:5173",
    trace: "retain-on-failure",
  },
  projects: [
    { name: "admin-chromium", use: { browserName: "chromium" } },
  ],
});
```

- [x] **Step 3: Add reusable admin/client fixtures**

```ts
export async function loginAdmin(page: Page) {
  await page.goto("/login");
  await page.getByLabel("用户名").fill("admin");
  await page.getByLabel("密码").fill("ChangeMe123!");
  await page.getByRole("button", { name: "登录后台" }).click();
}
```

- [x] **Step 4: Run config verification**

Run: `pnpm exec playwright test --list`

Expected: PASS。

### Task 3: Backfill Go Integration Test Harness

**Files:**
- Create: `apps/gateway/internal/database/integration_test.go`
- Create: `apps/gateway/internal/auth/integration_policy_test.go`
- Create: `apps/gateway/internal/auth/integration_device_session_test.go`
- Create: `apps/gateway/internal/auth/integration_audit_test.go`
- Modify: `apps/gateway/internal/auth/service_test_helpers_test.go`
- Modify: `apps/gateway/internal/database/migrate_test.go`

- [ ] **Step 1: Write the failing integration regression**

```go
func TestIntegrationUserServiceDenyOverridesRoleAllow(t *testing.T) {
	service := newIntegrationService(t)

	allowed, err := service.ResolveProxyRequest(ctx, auth.ResolveProxyRequestInput{
		AccessToken: bobAccessToken,
		ServiceKey:  "jenkins",
		RequestID:   "req_integration_policy",
	})
	if err == nil {
		t.Fatal("expected deny override to block jenkins access")
	}
}
```

Run: `BIFROST_DATABASE_TEST_URL='postgres://bifrost:bifrost@127.0.0.1:15432/postgres?sslmode=disable' GOTOOLCHAIN=local go test ./internal/database ./internal/auth -run 'Integration|Migrate'`

Expected: FAIL because reusable integration bootstrap helpers and focused integration suites do not exist.

- [ ] **Step 2: Build reusable integration bootstrap helper**

```go
func newIntegrationService(t *testing.T) Service {
	t.Helper()
	dsn := testDatabaseURL(t)
	seedPhase1(t, dsn)
	return newServiceForDatabase(t, dsn)
}
```

- [ ] **Step 3: Add policy, device/session, audit focused suites**

```go
func TestIntegrationDeviceDisabledBlocksRefresh(t *testing.T) { ... }
func TestIntegrationAuditListReturnsNewestEventsFirst(t *testing.T) { ... }
```

- [ ] **Step 4: Run Go integration verification**

Run: `BIFROST_DATABASE_TEST_URL='postgres://bifrost:bifrost@127.0.0.1:15432/postgres?sslmode=disable' GOTOOLCHAIN=local go test ./internal/database ./internal/auth -run 'Integration|Migrate'`

Expected: PASS。

### Task 4: Add First Real Playwright Scenarios

**Files:**
- Create: `tests/e2e/admin-login.spec.ts`
- Create: `tests/e2e/admin-user-create.spec.ts`
- Create: `tests/e2e/client-service-access.spec.ts`
- Create: `tests/e2e/request-id-smoke.spec.ts`
- Modify: `tests/e2e/fixtures/admin-session.ts`
- Modify: `tests/e2e/fixtures/client-api.ts`

- [x] **Step 1: Write the failing end-to-end scenarios**

```ts
test("admin can login and see dashboard shell", async ({ page }) => {
  await loginAdmin(page);
  await expect(page.getByText("系统概览")).toBeVisible();
});

test("alice can request gitlab access but not jenkins", async () => {
  const client = createClientAPI();
  const services = await client.loginAndListServices("alice", "ChangeMe123!");
  expect(services.map((item) => item.key)).toContain("gitlab");
  expect(services.map((item) => item.key)).not.toContain("jenkins");
});
```

Run: `pnpm exec playwright test tests/e2e/admin-login.spec.ts --project=admin-chromium`

Expected: FAIL because the specs and fixtures are not implemented yet.

- [x] **Step 2: Implement admin shell and client API scenarios**

```ts
const access = await client.createServiceAccessURL(session.accessToken, gitlab.id);
expect(access.url ?? access.publicPath).toContain("/s/gitlab");
```

- [x] **Step 3: Add requestId smoke assertion**

```ts
const response = await page.waitForResponse((candidate) => candidate.url().includes("/api/v1/admin/users"));
expect(response.headers()["x-request-id"] ?? "").not.toEqual("");
```

- [x] **Step 4: Run E2E verification**

Run: `pnpm exec playwright test`

Expected: PASS。

### Task 5: Update Checklist And Testing Docs

**Files:**
- Modify: `docs/08-roadmap/development-checklist.md`
- Modify: `docs/06-engineering/testing-strategy.md`
- Modify: `docs/06-engineering/local-docker-development.md`
- Modify: `docs/superpowers/plans/2026-04-18-phase8-testing-foundation.md`

- [x] **Step 1: Update checklist entries with real completion timestamps**

```md
- [x] 初始化 Playwright 配置（完成时间：2026-04-18 HH:mm CST）
- [x] 建立 Docker Compose 驱动的测试启动脚本（完成时间：2026-04-18 HH:mm CST）
- [x] 建立测试前数据库初始化脚本（完成时间：2026-04-18 HH:mm CST）
- [x] 建立测试后环境清理脚本（完成时间：2026-04-18 HH:mm CST）
```

- [x] **Step 2: Document exact local test workflow**

```md
1. `pnpm test:e2e:up`
2. `pnpm test:e2e:seed`
3. `pnpm exec playwright test`
4. `pnpm test:e2e:down`
```

- [x] **Step 3: Run final verification**

Run: `pnpm lint && pnpm check && pnpm test && pnpm test:infra`

Expected: PASS。

## Self-Review

- Spec coverage: 覆盖 Phase 8 中的 Docker 驱动测试启动、数据库初始化/清理、Go 集成测试基座、Playwright 配置与首批 E2E 场景。
- Placeholder scan: 所有任务都给出了明确文件、命令和示例代码，没有 `TODO` 或“后续补充”式占位。
- Type consistency: E2E 使用 `admin` 登录与 `alice` 客户端种子数据；数据库基座统一使用 `BIFROST_DATABASE_TEST_URL`；Playwright 场景依赖 root 级 `playwright.config.ts` 与 `tests/e2e/fixtures/*`。

## Execution Notes

- 2026-04-18 12:08 CST：完成 Docker E2E 编排脚本，默认使用 `15432/18080/15173` 测试端口，避免占用普通开发端口。
- 2026-04-18 12:08 CST：完成 Playwright root 配置与首批 4 条 API 级 E2E，覆盖管理员登录、客户端 bootstrap、服务列表、GitLab mock 真实访问、Jenkins deny 与 requestId。
- 2026-04-18 12:08 CST：执行 `pnpm test:e2e:down && pnpm test:e2e:up && pnpm test:e2e`，4 条 E2E 全部通过。
- 2026-04-18 12:08 CST：执行 `pnpm test:infra`，4 条 infra 测试全部通过。
- 2026-04-18 12:10 CST：执行 `pnpm lint`、`pnpm check`、`pnpm test`、`BIFROST_DATABASE_TEST_URL='postgres://bifrost:bifrost@127.0.0.1:15432/postgres?sslmode=disable' GOTOOLCHAIN=local go test ./...`，全部通过。
