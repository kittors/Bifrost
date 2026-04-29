# Phase 9 Manual Security Review

> 本文记录 Phase 9 中无法完全依赖自动化断言的安全核查项。核查结论以代码审查、聚焦 E2E 和现有单元测试为证据，后续若实现系统级代理、DNS、路由或更复杂的 Electron 能力，需要重新审阅本文。

**审查时间：** 2026-04-18 13:27 CST

**审查范围：**

- Electron 客户端是否修改系统代理、DNS、路由。
- Electron IPC 是否只暴露受控白名单能力。
- Gateway 日志是否避免输出密码、令牌、Cookie 等敏感值。
- Gateway 是否拒绝未配置的服务 key。
- 服务访问 Cookie 是否具备安全属性。

## 结论

- [x] 客户端未修改系统代理。
- [x] 客户端未修改系统 DNS。
- [x] 客户端未修改系统路由。
- [x] Gateway 访问日志未输出密码或令牌。
- [x] Gateway 不会代理未配置上游。
- [x] Electron IPC 未暴露任意 channel 或危险执行能力。
- [x] 服务访问 Cookie 使用 `HttpOnly`、`SameSite=Lax`、服务路径隔离和 HTTPS Secure 自适应。

## 审查证据

### 客户端不接管系统网络

审查命令：

```bash
rg -n "proxy|dns|route|networksetup|scutil|setProxy|systemProxy|HTTP_PROXY|HTTPS_PROXY|route add|route delete" apps/desktop/electron apps/desktop/renderer -S
```

结果只命中诊断展示和类型字段，不存在 `networksetup`、`scutil`、`route add`、`route delete` 或系统代理写入逻辑。

关键文件：

- `apps/desktop/electron/main/diagnostics.ts` 明确返回 `proxyManagedByBifrost: false`、`dnsManagedByBifrost: false`、`routeManagedByBifrost: false`。
- `apps/desktop/renderer/src/features/diagnostics/diagnostics-card.tsx` 只展示诊断状态，不执行系统修改。

### Electron IPC 白名单

关键文件：

- `apps/desktop/electron/shared/ipc.ts` 只定义固定 `bifrost:*` channel。
- `apps/desktop/electron/preload/index.ts` 通过 `contextBridge.exposeInMainWorld("bifrostDesktop", ...)` 暴露受控 API，没有暴露 `ipcRenderer`。
- `apps/desktop/electron/main/ipc.ts` 逐个 `ipcMain.handle(...)` 注册白名单 channel。
- `apps/desktop/electron/main/window.ts` 使用 `contextIsolation: true`、`nodeIntegration: false`、`sandbox: true`。

外部打开能力只允许 `http:` 和 `https:` URL，并拒绝任意新窗口导航。

### 日志不输出敏感信息

关键文件：

- `apps/gateway/internal/server/middleware.go`

访问日志字段只包含 `request_id`、`method`、`path`、`status` 和 `duration_ms`，不记录请求体、`Authorization`、Cookie、访问令牌、刷新令牌或密码。

### 未配置服务 key 不代理

关键文件：

- `apps/gateway/internal/auth/service_client_proxy.go`
- `tests/e2e/service-proxy-failures.spec.ts`

`loadProxyServiceByKey` 在服务 key 不存在时返回 `404 SERVICE_NOT_FOUND`，代理解析在查到服务前不会组装上游请求。

验证命令：

```bash
pnpm exec playwright test tests/e2e/service-proxy-failures.spec.ts --project=api-chromium
```

验证结果：3 个用例通过，其中 `unknown service key is rejected before any upstream proxying` 覆盖未配置服务 key。

### Cookie 安全属性

关键文件：

- `apps/gateway/internal/server/client_service_handlers.go`
- `apps/gateway/internal/server/client_service_routes_test.go`

服务访问 Cookie 使用：

- `HttpOnly: true`
- `SameSite: http.SameSiteLaxMode`
- `Path: result.PublicPath`
- `Secure: requestIsSecure(request)`

已有单元测试覆盖普通 HTTP 下不设置 Secure，以及 `X-Forwarded-Proto: https` 下设置 Secure。

