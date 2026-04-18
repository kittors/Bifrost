# Desktop Phase 7 MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 落地 Electron 小卡片式客户端 MVP，并补齐首台设备首次登录所需的安全 bootstrap 能力，让用户可以从桌面端完成登录、设备绑定、服务列表查看和打开受控服务。

**Architecture:** 后端新增“客户端首登设备 bootstrap” 入口，仍由 `auth.Service` 承载认证与设备信任逻辑、`server.App` 承载 HTTP 路由；桌面端采用 `electron-vite` 的 `main + preload + renderer` 结构，Main 负责窗口、安全存储、外部浏览器与 IPC 白名单，Preload 仅暴露受控桥接 API，Renderer 负责小卡片式 React UI 与业务状态。敏感数据通过 Main 层安全存储封装保存，不修改系统代理、DNS 或路由。

**Tech Stack:** Electron 41.2.1、electron-vite 6.0.0-beta.1（当前支持 Vite 8 的 registry 版本）、React 19.2.5、react-dom 19.2.5、TypeScript 6.0.3、Vite 8.0.8、Tailwind CSS v4、@vitejs/plugin-react 6.0.1、TanStack Query 5.99.0、Zustand 5.0.12、Lucide React 1.8.0、Vitest 4.1.4、Go 1.26.2。

---

### Task 1: Add Client Device Bootstrap Backend

**Files:**
- Modify: `apps/gateway/internal/auth/service_inputs.go`
- Create: `apps/gateway/internal/auth/service_client_bootstrap.go`
- Modify: `apps/gateway/internal/auth/service_auth.go`
- Modify: `apps/gateway/internal/auth/service_results.go`
- Modify: `apps/gateway/internal/auth/service_devices.go`
- Modify: `apps/gateway/internal/auth/service_admin_test.go`
- Modify: `apps/gateway/internal/auth/service_auth_test.go`
- Modify: `apps/gateway/internal/server/auth_session_handlers.go`
- Modify: `apps/gateway/internal/server/server.go`
- Modify: `apps/gateway/internal/server/auth_routes_test.go`
- Modify: `apps/gateway/internal/contracts/generated.go`
- Modify: `docs/05-api/client-api-design.md`

- [x] **Step 1: Write failing backend tests**

```go
func TestServiceClientBootstrapDevice(t *testing.T) {
  result, err := service.BootstrapClientDevice(ctx, auth.BootstrapClientDeviceInput{
    Username: "alice",
    Password: "correct horse battery staple",
    DeviceName: "Alice MacBook Pro",
    DeviceOS: "macOS",
    ClientVersion: "0.1.0",
    PublicKey: publicKey,
    PublicKeyFingerprint: fingerprint,
  })
  if err != nil { t.Fatalf("bootstrap device: %v", err) }
  if result.Device.ID == "" || result.User.Username != "alice" { t.Fatalf("unexpected bootstrap result: %#v", result) }
}
```

Run: `BIFROST_DATABASE_TEST_URL='postgres://bifrost:bifrost@127.0.0.1:15432/postgres?sslmode=disable' GOTOOLCHAIN=local go test ./internal/auth ./internal/server -run 'Bootstrap|ClientLogin'`

Expected: FAIL because bootstrap input, service, and route do not exist.

- [x] **Step 2: Implement bootstrap service**

```go
type BootstrapClientDeviceInput struct {
  Username string
  Password string
  DeviceName string
  DeviceOS string
  ClientVersion string
  PublicKey string
  PublicKeyFingerprint string
  RequestID string
}

func (s Service) BootstrapClientDevice(ctx context.Context, input BootstrapClientDeviceInput) (ClientBootstrapResult, error) { ... }
```

- [x] **Step 3: Wire HTTP route and update API docs**

```text
POST /api/v1/client/devices/bootstrap
```

- [x] **Step 4: Run backend verification**

Run: `BIFROST_DATABASE_TEST_URL='postgres://bifrost:bifrost@127.0.0.1:15432/postgres?sslmode=disable' GOTOOLCHAIN=local go test ./...`

Expected: PASS。

### Task 2: Set Up Electron Build And Secure Main/Preload Skeleton

**Files:**
- Modify: `apps/desktop/package.json`
- Modify: `apps/desktop/tsconfig.json`
- Create: `apps/desktop/electron.vite.config.ts`
- Create: `apps/desktop/electron/shared/ipc.ts`
- Create: `apps/desktop/electron/main/index.ts`
- Create: `apps/desktop/electron/main/window.ts`
- Create: `apps/desktop/electron/main/ipc.ts`
- Create: `apps/desktop/electron/main/security-store.ts`
- Create: `apps/desktop/electron/main/device-identity.ts`
- Create: `apps/desktop/electron/main/diagnostics.ts`
- Create: `apps/desktop/electron/preload/index.ts`
- Modify: `apps/desktop/renderer/src/globals.d.ts`

- [x] **Step 1: Add failing desktop type tests**

```ts
it("exposes only the whitelisted desktop bridge API", () => {
  expectTypeOf(window.bifrostDesktop.session.load).toBeFunction();
  expectTypeOf(window.bifrostDesktop.openExternal).toBeFunction();
});
```

Run: `pnpm --filter @bifrost/desktop check`

Expected: FAIL because bridge types and electron-vite entry files are incomplete.

- [x] **Step 2: Add Electron dependencies and config**

```json
{
  "dependencies": {
    "electron": "41.2.1",
    "react": "19.2.5",
    "react-dom": "19.2.5",
    "@tanstack/react-query": "5.99.0",
    "zustand": "5.0.12",
    "lucide-react": "1.8.0"
  },
  "devDependencies": {
    "electron-vite": "6.0.0-beta.1",
    "@vitejs/plugin-react": "6.0.1",
    "@tailwindcss/vite": "4.2.2"
  }
}
```

- [x] **Step 3: Implement secure BrowserWindow and IPC whitelist**

```ts
webPreferences: {
  preload,
  contextIsolation: true,
  nodeIntegration: false,
  sandbox: true,
}
```

- [x] **Step 4: Run desktop type verification**

Run: `pnpm --filter @bifrost/desktop check`

Expected: PASS。

### Task 3: Build Desktop Data Layer For Session, Device, Services, Diagnostics

**Files:**
- Create: `apps/desktop/renderer/src/shared/config/env.ts`
- Create: `apps/desktop/renderer/src/shared/lib/http.ts`
- Create: `apps/desktop/renderer/src/shared/lib/base64url.ts`
- Create: `apps/desktop/renderer/src/entities/session/api.ts`
- Create: `apps/desktop/renderer/src/entities/session/store.ts`
- Create: `apps/desktop/renderer/src/entities/session/types.ts`
- Create: `apps/desktop/renderer/src/entities/device/api.ts`
- Create: `apps/desktop/renderer/src/entities/device/types.ts`
- Create: `apps/desktop/renderer/src/entities/services/api.ts`
- Create: `apps/desktop/renderer/src/entities/services/types.ts`
- Create: `apps/desktop/renderer/src/entities/diagnostics/types.ts`
- Create: `apps/desktop/renderer/src/entities/session/api.test.ts`
- Create: `apps/desktop/renderer/src/entities/device/api.test.ts`
- Create: `apps/desktop/renderer/src/entities/services/api.test.ts`

- [x] **Step 1: Write failing renderer API tests**

```ts
it("bootstraps a new client device when no device id exists", async () => { ... });
it("refreshes a client session with secure-stored refresh token", async () => { ... });
it("requests a service access url and opens it through preload", async () => { ... });
```

Run: `pnpm --filter @bifrost/desktop test`

Expected: FAIL because API helpers and store flows do not exist.

- [x] **Step 2: Implement API modules and Zustand session store**

```ts
export async function bootstrapClientDevice(...) { ... }
export async function clientLogin(...) { ... }
export async function listClientServices(...) { ... }
export const useDesktopSessionStore = create<DesktopSessionState>()(...)
```

- [x] **Step 3: Implement secure session/device orchestration**

```ts
if (!deviceIdentity.deviceId) {
  const bootstrap = await bootstrapClientDevice(...)
  await window.bifrostDesktop.session.save(...)
}
```

- [x] **Step 4: Run desktop tests**

Run: `pnpm --filter @bifrost/desktop test && pnpm --filter @bifrost/desktop check`

Expected: PASS。

### Task 4: Build Compact Desktop Renderer UI

**Files:**
- Create: `apps/desktop/renderer/index.html`
- Create: `apps/desktop/renderer/src/main.tsx`
- Create: `apps/desktop/renderer/src/app/app.tsx`
- Create: `apps/desktop/renderer/src/app/providers.tsx`
- Create: `apps/desktop/renderer/src/app/layout/window-shell.tsx`
- Create: `apps/desktop/renderer/src/features/auth/login-card.tsx`
- Create: `apps/desktop/renderer/src/features/services/services-card.tsx`
- Create: `apps/desktop/renderer/src/features/account/account-card.tsx`
- Create: `apps/desktop/renderer/src/features/settings/settings-card.tsx`
- Create: `apps/desktop/renderer/src/features/diagnostics/diagnostics-card.tsx`
- Create: `apps/desktop/renderer/src/features/chrome/connection-banner.tsx`
- Create: `apps/desktop/renderer/src/features/chrome/section-tabs.tsx`
- Modify: `apps/desktop/renderer/src/main.ts`

- [x] **Step 1: Write failing renderer component smoke tests**

```ts
it("renders the compact desktop shell with login state", () => { ... });
it("renders services, account, settings and diagnostics tabs after login", () => { ... });
```

Run: `pnpm --filter @bifrost/desktop test`

Expected: FAIL because React renderer shell and cards do not exist.

- [x] **Step 2: Implement compact card UI**

```tsx
<WindowShell>
  <ConnectionBanner />
  <SectionTabs />
  <ServicesCard />
  <AccountCard />
  <SettingsCard />
  <DiagnosticsCard />
</WindowShell>
```

- [x] **Step 3: Wire actions**

```tsx
await openService(service.id)
await window.bifrostDesktop.openExternal(url)
```

- [x] **Step 4: Run desktop verification**

Run: `pnpm --filter @bifrost/desktop test && pnpm --filter @bifrost/desktop check && pnpm lint`

Expected: PASS。

### Task 5: Verify Non-Interference, Update Checklist, Commit, Push

**Files:**
- Modify: `docs/08-roadmap/development-checklist.md`
- Modify: `docs/superpowers/plans/2026-04-18-desktop-phase7-mvp.md`
- Modify: `docs/05-api/client-api-design.md`

- [x] **Step 1: Add diagnostics and non-interference evidence**

Run: `scutil --proxy 2>/dev/null || true`

Run: `networksetup -getwebproxy Wi-Fi 2>/dev/null || true`

Run: `route -n get default 2>/dev/null || ip route 2>/dev/null || true`

Expected: 客户端代码与验证脚本都未改写系统代理、DNS 或默认路由。

- [x] **Step 2: Run full verification**

Run: `BIFROST_DATABASE_TEST_URL='postgres://bifrost:bifrost@127.0.0.1:15432/postgres?sslmode=disable' GOTOOLCHAIN=local go test ./...`

Run: `pnpm lint && pnpm check && pnpm test`

Expected: PASS。

- [x] **Step 3: Update checklist and commit**

```md
- [x] 初始化 Electron 主进程（完成时间：2026-04-18 HH:mm CST）
- [x] 初始化 Preload（完成时间：2026-04-18 HH:mm CST）
- [x] 初始化 Renderer（完成时间：2026-04-18 HH:mm CST）
- [x] 禁用 Renderer 直接 Node.js 访问（完成时间：2026-04-18 HH:mm CST）
- [x] 启用 `contextIsolation`（完成时间：2026-04-18 HH:mm CST）
- [x] 关闭 `nodeIntegration`（完成时间：2026-04-18 HH:mm CST）
- [x] 建立 IPC 白名单（完成时间：2026-04-18 HH:mm CST）
- [x] 接入共享 token 与共享 UI（完成时间：2026-04-18 HH:mm CST）
- [x] 建立小卡片式主窗口（完成时间：2026-04-18 HH:mm CST）
- [x] 实现登录页（完成时间：2026-04-18 HH:mm CST）
- [x] 实现服务列表主视图（完成时间：2026-04-18 HH:mm CST）
- [x] 实现账号面板（完成时间：2026-04-18 HH:mm CST）
- [x] 实现设置面板（完成时间：2026-04-18 HH:mm CST）
- [x] 实现诊断面板（完成时间：2026-04-18 HH:mm CST）
- [x] 实现本地安全存储抽象（完成时间：2026-04-18 HH:mm CST）
- [x] 实现 macOS 安全存储适配（完成时间：2026-04-18 HH:mm CST）
- [x] 实现 Windows 安全存储适配（完成时间：2026-04-18 HH:mm CST）
- [x] 实现 Linux 安全存储适配（完成时间：2026-04-18 HH:mm CST）
- [x] 实现设备密钥生成（完成时间：2026-04-18 HH:mm CST）
- [x] 实现设备注册流程（完成时间：2026-04-18 HH:mm CST）
- [x] 实现登录与刷新流程（完成时间：2026-04-18 HH:mm CST）
- [x] 实现服务列表拉取（完成时间：2026-04-18 HH:mm CST）
- [x] 实现服务访问 URL 请求（完成时间：2026-04-18 HH:mm CST）
- [x] 实现打开系统浏览器（完成时间：2026-04-18 HH:mm CST）
- [x] 实现设备禁用状态提示（完成时间：2026-04-18 HH:mm CST）
- [x] 实现登录失效清理（完成时间：2026-04-18 HH:mm CST）
- [x] 验证客户端不修改系统代理（完成时间：2026-04-18 HH:mm CST）
- [x] 验证客户端不修改系统 DNS（完成时间：2026-04-18 HH:mm CST）
- [x] 验证客户端不修改系统路由（完成时间：2026-04-18 HH:mm CST）
```

Run: `git add apps/desktop apps/gateway docs/05-api/client-api-design.md docs/08-roadmap/development-checklist.md docs/superpowers/plans/2026-04-18-desktop-phase7-mvp.md`

Run: `git commit -m "feat: add desktop phase seven mvp"`

Run: `git push origin dev`

## Self-Review

- Spec coverage: 覆盖桌面客户端小卡片 MVP、首登设备 bootstrap、Main/Preload/Renderer 安全边界、安全存储、服务打开与不干扰系统网络约束。
- Placeholder scan: 无 `TBD`、`TODO` 或“后续补充”式占位。
- Type consistency: 后端使用 `BootstrapClientDeviceInput` / `ClientBootstrapResult`，桌面端使用 `window.bifrostDesktop` 白名单桥接与 `DesktopSessionState`。

## Execution Notes

- 2026-04-18 11:46 CST：完成桌面端错误提示映射收口，新增 `store.test.ts` 验证设备禁用时优先展示 `userMessage`。
- 2026-04-18 11:46 CST：重新执行 `pnpm exec biome check --write apps/desktop`、`pnpm --filter @bifrost/desktop test`、`pnpm --filter @bifrost/desktop check`，全部通过。
- 2026-04-18 11:46 CST：重新执行 `BIFROST_DATABASE_TEST_URL='postgres://bifrost:bifrost@127.0.0.1:15432/postgres?sslmode=disable' GOTOOLCHAIN=local go test ./...`、`pnpm lint`、`pnpm check`、`pnpm test`，全部通过。
- 2026-04-18 11:46 CST：通过 `scutil --proxy`、`networksetup -getwebproxy Wi-Fi`、`route -n get default` 与代码检索确认当前机器虽存在外部代理/VPN，但 Bifrost 未实现任何系统代理、DNS、路由改写逻辑。
