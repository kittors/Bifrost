# Phase 9 Acceptance Stability Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 Phase 9 的联调验收从人工口头确认推进为可重复执行的 Docker E2E、服务端集成测试和安全审查记录。

**Architecture:** 继续复用根目录 Playwright 多容器环境，不新增第二套测试世界。E2E 只通过公开 Admin / Client / Proxy API 驱动真实 Gateway、PostgreSQL 和 mock 上游；手工安全项以代码审查记录、命令输出和 checklist 时间戳闭环。测试辅助函数继续集中在 `tests/e2e/fixtures/client-api.ts`，避免 spec 文件重复拼接响应结构。

**Tech Stack:** Playwright 1.59.1、Vitest 4.1.4、Go 1.26.2、PostgreSQL 18.x、Docker Compose v2、Node.js 24.x、pnpm 10.33.0。

---

### Task 1: Add Phase 9 E2E Helper Coverage

**Files:**
- Modify: `tests/e2e/fixtures/client-api.ts`

- [x] **Step 1: Add admin device and user override helpers**

```ts
export async function listAdminDevices(request, accessToken, userID) {
  const { payload, response } = await requestJSON(request, {
    accessToken,
    path: `/api/v1/admin/devices?userId=${encodeURIComponent(userID)}`,
  });
  expect(response.ok()).toBeTruthy();
  return payload.data.items;
}

export async function setAdminDeviceStatus(request, accessToken, deviceID, status) {
  const { payload, response } = await requestJSON(request, {
    accessToken,
    body: { status },
    method: "POST",
    path: `/api/v1/admin/devices/${deviceID}/status`,
  });
  expect(response.ok()).toBeTruthy();
  return payload.data;
}

export async function replaceUserServiceOverrides(request, accessToken, userID, input) {
  const { payload, response } = await requestJSON(request, {
    accessToken,
    body: input,
    method: "PUT",
    path: `/api/v1/admin/users/${userID}/service-overrides`,
  });
  expect(response.ok()).toBeTruthy();
  return payload.data.items;
}
```

- [x] **Step 2: Add audit query helper**

```ts
export async function listAdminAuditEvents(request, accessToken, filters = {}) {
  const query = new URLSearchParams(filters);
  const { payload, response } = await requestJSON(request, {
    accessToken,
    path: `/api/v1/admin/audit-events?${query.toString()}`,
  });
  expect(response.ok()).toBeTruthy();
  return payload.data.items;
}
```

- [x] **Step 3: Run fixture type verification**

Run: `pnpm exec tsc --noEmit --project apps/admin/tsconfig.json`

Expected: PASS。

### Task 2: Add Policy And Device Acceptance E2E

**Files:**
- Create: `tests/e2e/policy-live-update.spec.ts`

- [x] **Step 1: Add user deny acceptance scenario**

```ts
test("user-level deny takes effect immediately for proxy access", async ({ request }) => {
  const admin = await adminLogin(request, "admin", seedPassword);
  const client = await bootstrapClientDevice(request, "alice", seedPassword);

  await replaceUserServiceOverrides(request, admin.session.accessToken, client.session.user.id, {
    allowServiceIds: [],
    denyServiceIds: ["service_gitlab"],
  });

  const denied = await proxyServiceRequest(request, client.session.accessToken, "gitlab");
  expect(denied.response.status()).toBe(403);
  expect(denied.payload.error?.code).toBe("POLICY_ACCESS_DENIED");
});
```

- [x] **Step 2: Add device disabled acceptance scenario**

```ts
test("device disable takes effect immediately for proxy access", async ({ request }) => {
  const admin = await adminLogin(request, "admin", seedPassword);
  const client = await bootstrapClientDevice(request, "alice", seedPassword);

  await setAdminDeviceStatus(request, admin.session.accessToken, client.device.deviceId, "disabled");

  const denied = await proxyServiceRequest(request, client.session.accessToken, "gitlab");
  expect(denied.response.status()).toBe(403);
  expect(denied.payload.error?.code).toBe("DEVICE_DISABLED");
});
```

- [x] **Step 3: Restore changed policy/device state in `finally` blocks**

```ts
finally {
  await replaceUserServiceOverrides(request, admin.session.accessToken, client.session.user.id, {
    allowServiceIds: [],
    denyServiceIds: [],
  });
  await setAdminDeviceStatus(request, admin.session.accessToken, client.device.deviceId, "enabled");
}
```

- [x] **Step 4: Run focused E2E**

Run: `pnpm exec playwright test tests/e2e/policy-live-update.spec.ts --project=api-chromium`

Expected: PASS。

### Task 3: Add Upstream Timeout And Audit Acceptance E2E

**Files:**
- Create: `tests/e2e/stability-audit.spec.ts`
- Modify: `tests/e2e/service-proxy-failures.spec.ts`

- [x] **Step 1: Add upstream slow response timeout scenario**

```ts
test("slow upstream returns gateway timeout", async ({ request }) => {
  const client = await bootstrapClientDevice(request, "alice", seedPassword);
  const failed = await proxyServiceRequest(request, client.session.accessToken, "gitlab", "slow?delayMs=7000");
  expect(failed.response.status()).toBe(504);
  expect(failed.payload.error?.code).toBe("GATEWAY_UPSTREAM_TIMEOUT");
});
```

- [x] **Step 2: Add audit success and denial query scenario**

```ts
test("audit list contains login and proxy outcomes", async ({ request }) => {
  const admin = await adminLogin(request, "admin", seedPassword);
  const client = await bootstrapClientDevice(request, "alice", seedPassword);
  await proxyServiceRequest(request, client.session.accessToken, "gitlab");
  await proxyServiceRequest(request, client.session.accessToken, "jenkins");

  const loginEvents = await listAdminAuditEvents(request, admin.session.accessToken, {
    type: "auth.login.succeeded",
    result: "success",
  });
  const deniedEvents = await listAdminAuditEvents(request, admin.session.accessToken, {
    type: "service.access.denied",
    result: "failure",
  });

  expect(loginEvents.length).toBeGreaterThan(0);
  expect(deniedEvents.length).toBeGreaterThan(0);
});
```

- [x] **Step 3: Run focused E2E**

Run: `pnpm exec playwright test tests/e2e/service-proxy-failures.spec.ts tests/e2e/stability-audit.spec.ts --project=api-chromium`

Expected: PASS。

### Task 4: Complete Phase 9 Checklist And Verification

**Files:**
- Modify: `docs/08-roadmap/development-checklist.md`

- [x] **Step 1: Mark automated acceptance items complete with CST timestamps**

```md
- [x] 验证角色授权可立即生效（完成时间：YYYY-MM-DD HH:mm CST）
- [x] 验证用户级 deny 可立即生效（完成时间：YYYY-MM-DD HH:mm CST）
- [x] 验证设备禁用可立即生效（完成时间：YYYY-MM-DD HH:mm CST）
- [x] 验证上游慢响应后返回 `504`（完成时间：YYYY-MM-DD HH:mm CST）
```

- [x] **Step 2: Run full verification**

Run:
- `pnpm lint`
- `pnpm check`
- `pnpm test`
- `pnpm test:infra`
- `BIFROST_DATABASE_TEST_URL='postgres://bifrost:bifrost@127.0.0.1:15432/postgres?sslmode=disable' go test ./...`
- `pnpm test:e2e:up`
- `pnpm test:e2e`
- `pnpm test:e2e:down`

Expected: PASS。

**Verification result:** 2026-04-18 13:19 CST 已执行 `pnpm lint`、`pnpm check`、`pnpm test`、`pnpm test:infra`、`BIFROST_DATABASE_TEST_URL='postgres://bifrost:bifrost@127.0.0.1:15432/postgres?sslmode=disable' go test ./...`、`pnpm test:e2e`，全部通过；Playwright E2E 共 13 个场景通过。

## Self-Review

- Spec coverage: 覆盖 Phase 9 中可自动化的配置可见性、授权即时生效、用户 deny、设备禁用、服务禁用、上游 502/504 和审计查询项；手工安全检查项会在后续 Task 中通过代码审查记录补齐。
- Placeholder scan: 无 `TBD` / `TODO` / “稍后实现” 占位。
- Type consistency: E2E helper 命名与现有 `client-api.ts` 风格一致，响应断言沿用统一 `payload.success/error.code` 结构。
