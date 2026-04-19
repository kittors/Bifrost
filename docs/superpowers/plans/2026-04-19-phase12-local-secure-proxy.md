# Phase 12 Local Secure Proxy Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让桌面客户端登录并绑定设备后，在 `127.0.0.1` 上启动项目专用本地代理，使用户可以用自己的浏览器/API 工具访问被防火墙保护的 Web/API 服务。

**Architecture:** 客户端 Main Process 启动只绑定 loopback 的 HTTP 代理，不修改系统代理、DNS、路由或 VPN。Renderer 只通过白名单 IPC 读取代理状态、打开服务入口和展示本地访问地址；本地代理把浏览器请求转发到 Gateway 的 `/s/{serviceKey}`，并在 Main Process 内注入设备会话访问令牌。Gateway 仍然负责最终用户、设备、服务、角色和策略鉴权。

**Tech Stack:** Electron 41.2.1、Node.js 24.11.1、TypeScript 6.0.3、React 19.2.5、Tailwind CSS v4、Vitest 4.1.4、Go 1.26.2、Playwright 1.59.1。

---

### Task 1: Add Desktop Local Proxy Core

**Files:**
- Create: `apps/desktop/electron/main/local-proxy.ts`
- Create: `apps/desktop/electron/main/local-proxy.test.ts`
- Modify: `apps/desktop/vitest.config.ts`

- [x] **Step 1: Write failing proxy route tests**

Add tests that start a fake Gateway, start the local proxy with a desktop session, call `http://127.0.0.1:<port>/s/gitlab/whoami`, and assert:

- local proxy binds `127.0.0.1`
- request is forwarded to `/s/gitlab/whoami`
- `Authorization: Bearer <accessToken>` is injected
- `proxyManagedByBifrost` remains a diagnostic flag only and no system proxy API is used

Run: `pnpm --filter @bifrost/desktop test -- electron/main/local-proxy.test.ts`

Expected: FAIL because `local-proxy.ts` does not exist.

- [x] **Step 2: Implement minimal local proxy**

Create `createLocalProxyController()` with:

- `start(session)`
- `stop()`
- `status()`
- `openService(publicPath)`

Implementation constraints:

- Listen host is always `127.0.0.1`
- Preferred port is `18080`
- If busy, try up to `18099`
- Only forward paths starting with `/s/`
- Inject `Authorization` inside Main Process
- Strip hop-by-hop headers
- Never call OS proxy, DNS, route, VPN, PAC or TUN APIs

Run: `pnpm --filter @bifrost/desktop test -- electron/main/local-proxy.test.ts`

Expected: PASS.

### Task 2: Wire Local Proxy Through Safe IPC

**Files:**
- Modify: `apps/desktop/electron/shared/ipc.ts`
- Modify: `apps/desktop/electron/shared/types.ts`
- Modify: `apps/desktop/electron/main/ipc.ts`
- Modify: `apps/desktop/electron/preload/index.ts`
- Modify: `apps/desktop/renderer/src/desktop-bridge.d.ts`

- [x] **Step 1: Write failing IPC bridge tests through existing renderer mocks**

Extend renderer tests so `window.bifrostDesktop.localProxy` is required and service cards call `localProxy.openService(...)` instead of `openExternal(...)`.

Run: `pnpm --filter @bifrost/desktop test -- renderer/src/features/services/services-card.test.tsx`

Expected: FAIL because `localProxy` bridge does not exist.

- [x] **Step 2: Add IPC contracts**

Add these channels:

- `bifrost:local-proxy:start`
- `bifrost:local-proxy:stop`
- `bifrost:local-proxy:status`
- `bifrost:local-proxy:open-service`

Expose only typed methods:

- `start(session)`
- `stop()`
- `status()`
- `openService(publicPath)`

Run: `pnpm --filter @bifrost/desktop check`

Expected: PASS.

### Task 3: Start And Stop Proxy With Session Lifecycle

**Files:**
- Modify: `apps/desktop/renderer/src/entities/session/store.ts`
- Modify: `apps/desktop/renderer/src/entities/session/types.ts`
- Modify: `apps/desktop/renderer/src/entities/session/store.test.ts`
- Modify: `apps/desktop/renderer/src/features/account/account-card.tsx`

- [x] **Step 1: Write failing session lifecycle tests**

Assert:

- `saveSession()` starts the local proxy
- `hydrateFromSecureStore()` refreshes session and starts the local proxy
- `clearSession()` stops the local proxy
- failure to start the proxy leaves the session saved but exposes a user-readable error

Run: `pnpm --filter @bifrost/desktop test -- renderer/src/entities/session/store.test.ts`

Expected: FAIL until store calls the new bridge.

- [x] **Step 2: Implement lifecycle integration**

Add `localProxyStatus` to store state and ensure:

- login success starts proxy
- app restore starts proxy after refresh
- logout stops proxy before clearing UI state
- no system network setting is modified

Run: `pnpm --filter @bifrost/desktop test -- renderer/src/entities/session/store.test.ts`

Expected: PASS.

### Task 4: Update Service Card And Diagnostics UI

**Files:**
- Modify: `apps/desktop/renderer/src/features/services/services-card.tsx`
- Modify: `apps/desktop/renderer/src/features/services/services-card.test.tsx`
- Modify: `apps/desktop/renderer/src/features/diagnostics/diagnostics-card.tsx`
- Modify: `docs/02-architecture/client-architecture.md`
- Modify: `docs/07-deployment/desktop-startup-guide.md`

- [x] **Step 1: Write failing service UI tests**

Assert service card:

- shows a local proxy entry URL like `http://127.0.0.1:18080/s/gitlab/`
- opens the service through `localProxy.openService("/s/gitlab/")`
- still states that Bifrost does not modify system proxy/DNS/routes

Run: `pnpm --filter @bifrost/desktop test -- renderer/src/features/services/services-card.test.tsx`

Expected: FAIL until UI is wired.

- [x] **Step 2: Implement compact UI**

Update copy to:

- “本地入口已启用”
- “浏览器访问本机地址，不接管系统网络”
- Keep compact card layout and Tailwind v4 utilities only

Run: `pnpm --filter @bifrost/desktop test -- renderer/src/features/services/services-card.test.tsx`

Expected: PASS.

### Task 5: E2E And Checklist Closure

**Files:**
- Create: `tests/e2e/local-proxy-access.spec.ts`
- Modify: `docs/08-roadmap/development-checklist.md`

- [x] **Step 1: Add E2E coverage for final user scenario**

Add an E2E scenario that:

- bootstraps a client device
- obtains allowed services
- calls Gateway proxy with a bearer token to prove the server path still works
- verifies disabled service / unauthorized service remains denied

Because Playwright E2E does not launch packaged Electron in this repo, renderer and Main Process tests cover the actual local loopback server, while E2E covers Gateway policy correctness.

Run: `pnpm test:e2e`

Expected: PASS.

- [x] **Step 2: Update checklist**

Append Phase 12 checklist items and mark each item complete only after local tests and E2E pass.

Run:

- `pnpm lint`
- `pnpm check`
- `pnpm test`
- `pnpm test:infra`
- `pnpm test:e2e`
- `pnpm --filter @bifrost/desktop build`

Expected: PASS.
