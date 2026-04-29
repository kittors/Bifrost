# Bifrost

Bifrost 是一个企业私有 Web 服务访问网关。

第一阶段聚焦于通过以下能力，受控访问企业内部 `HTTP/HTTPS` 服务：

- Go Gateway 服务端
- Electron 桌面客户端
- React 管理后台
- 共享设计系统与 API 契约
- 基于 Docker 的本地多容器测试环境

桌面客户端刻意保持轻量：它是一个卡片式访问入口，不是 VPN，不是系统代理，也不是管理后台。

## 文档入口

建议先阅读 [docs/00-overview/README.md](./docs/00-overview/README.md)。

项目实现清单维护在 [docs/08-roadmap/development-checklist.md](./docs/08-roadmap/development-checklist.md)。已完成事项需要记录完成时间。

## 本地开发

### 安装依赖

```bash
corepack enable
corepack prepare pnpm@10.33.0 --activate
pnpm install --frozen-lockfile
```

项目依赖统一使用 `pnpm` 安装；`npm install`、`yarn install` 和 `bun install` 会被 `preinstall` 守卫拒绝。

首次安装会自动执行 Electron、esbuild 等桌面端必需的构建脚本。如果你之前看到过 `Ignored build scripts` 或启动时报 `Electron uninstall`，拉取最新配置后重新执行 `pnpm install --frozen-lockfile` 即可修复。

### 启动命令速查

| 目标 | 启动命令 | 默认访问地址 | 停止命令 |
|---|---|---|---|
| 只启动后端服务 | `pnpm dev:backend` | Gateway：`http://127.0.0.1:18080` | `pnpm dev:backend:down` |
| 启动后台 Admin Web | `pnpm test:e2e:up` | Admin Web：`http://127.0.0.1:15173` | `pnpm test:e2e:down` |
| 启动桌面客户端 | `pnpm --filter @bifrost/desktop dev` | Electron 开发服务：`http://127.0.0.1:22473` | 关闭 Electron 窗口，必要时再停后端 |

### 启动后端服务

后端服务包含 PostgreSQL、Gateway 和多个 mock 私有服务，适合调试 Gateway API、客户端登录、服务列表和代理链路。

启动：

```bash
pnpm dev:backend
```

默认地址：

```text
Gateway:   http://127.0.0.1:18080
Postgres:  127.0.0.1:15432
```

健康检查：

```bash
curl http://127.0.0.1:18080/healthz
```

停止：

```bash
pnpm dev:backend:down
```

### 启动后台 Admin Web

后台 Admin Web 需要 Gateway 和数据库一起运行。最省心的方式是启动完整本地联调环境：

```bash
pnpm test:e2e:up
```

默认地址：

```text
Admin Web: http://127.0.0.1:15173
Gateway:   http://127.0.0.1:18080
Postgres:  127.0.0.1:15432
```

后台登录账号：

```text
用户名：admin
密码：  ChangeMe123!
```

停止：

```bash
pnpm test:e2e:down
```

注意：`pnpm dev:backend` 只启动后端环境，不启动 Admin Web。如果你要看后台页面，请使用 `pnpm test:e2e:up`。

### 启动桌面客户端

桌面客户端是 Electron 应用。它可以单独启动，但要完成登录、拉取服务列表和打开受控服务，需要先启动后端服务。

先启动后端：

```bash
pnpm dev:backend
```

如果你已经通过 `pnpm test:e2e:up` 启动了完整联调环境，可以跳过这一步，直接启动桌面客户端。

另开一个终端，启动桌面客户端：

```bash
pnpm --filter @bifrost/desktop dev
```

桌面客户端开发服务器固定监听 `http://127.0.0.1:22473`，不会再使用 Vite 默认的 `5173/5174`。如果你的机器上 `22473` 临时被占用，可以显式指定另一个空闲端口：

```bash
BIFROST_DESKTOP_DEV_PORT=22474 pnpm --filter @bifrost/desktop dev
```

客户端登录信息：

```text
Server URL: http://127.0.0.1:18080
Username:   alice
Password:   ChangeMe123!
```

登录成功后，客户端会在 `127.0.0.1:18080` 到 `127.0.0.1:18099` 中选择一个可用端口启动 Bifrost 专用本地回环代理。因为本地 Gateway 默认占用 `18080`，客户端通常会自动避让到下一个可用端口；实际服务入口以客户端界面显示的地址为准。

停止后端：

```bash
pnpm dev:backend:down
```

### 运行完整本地 E2E

如果你只是想跑一遍完整回归，不需要手动打开页面或客户端，执行：

```bash
pnpm test:e2e
```

`pnpm test:e2e` 会自动清理旧容器、启动 Docker Compose 测试环境、运行 Playwright E2E，并在结束后回收容器。

### 常用本地入口

- Gateway 健康检查：`http://127.0.0.1:18080/healthz`
- Admin Web：`http://127.0.0.1:15173`（由 `pnpm test:e2e` / `pnpm test:e2e:up` 启动）
- E2E PostgreSQL：`127.0.0.1:15432`

## 发布命令

```bash
pnpm build:gateway:image
pnpm build:admin:image
pnpm --filter @bifrost/desktop build
CSC_IDENTITY_AUTO_DISCOVERY=false pnpm --filter @bifrost/desktop exec electron-builder --config electron-builder.yml --dir
```

三端桌面安装包由 [desktop-packages.yml](./.github/workflows/desktop-packages.yml) 在 macOS、Windows 和 Linux runner 上构建并上传构建产物。

## 部署文档

- [服务端运行参数说明](./docs/07-deployment/service-runtime-parameters.md)
- [Admin Web 启动说明](./docs/07-deployment/admin-startup-guide.md)
- [Desktop Client 启动说明](./docs/07-deployment/desktop-startup-guide.md)
- [数据库 Migration 执行说明](./docs/07-deployment/database-migration-guide.md)
- [数据库备份与恢复说明](./docs/07-deployment/database-backup-and-restore.md)
- [TLS 与证书配置说明](./docs/07-deployment/tls-and-certificates.md)
- [日志与监控接入说明](./docs/07-deployment/logging-and-monitoring.md)
- [第一版内部试用发布说明](./docs/07-deployment/internal-trial-release-notes.md)
- [已知问题列表](./docs/07-deployment/known-issues.md)
- [回滚方案](./docs/07-deployment/rollback-guide.md)
- [requestId 排障入口](./docs/07-deployment/request-id-troubleshooting.md)

## 开发状态

当前实现工作在 `dev` 分支推进。Phase 9 验收已完成，Phase 10/11 的发布与 CI 工作仍在推进中。
