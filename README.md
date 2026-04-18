# Bifrost

Bifrost is an enterprise private web service access gateway.

The first phase focuses on controlled access to internal `HTTP/HTTPS` services through:

- a Go gateway server
- an Electron desktop client
- a React admin web console
- a shared design system and API contract
- a local Docker-based multi-container test environment

The desktop client is intentionally small: it is a card-like access launcher, not a VPN, not a system proxy, and not an admin console.

## Documentation

Start with [docs/00-overview/README.md](./docs/00-overview/README.md).

The implementation checklist is maintained in [docs/08-roadmap/development-checklist.md](./docs/08-roadmap/development-checklist.md). Completed items must include a completion timestamp.

## Local Development

```bash
pnpm install
pnpm test:e2e
```

`pnpm test:e2e` 会自动清理旧容器、启动 Docker Compose 测试环境、运行 Playwright E2E，并在结束后回收容器。

常用入口：

- Gateway health: `http://127.0.0.1:18080/healthz`
- Admin Web: `http://127.0.0.1:15173`
- E2E PostgreSQL: `127.0.0.1:15432`

## Release Commands

```bash
pnpm build:gateway:image
pnpm build:admin:image
pnpm --filter @bifrost/desktop build
CSC_IDENTITY_AUTO_DISCOVERY=false pnpm --filter @bifrost/desktop exec electron-builder --config electron-builder.yml --dir
```

三端桌面安装包由 [desktop-packages.yml](./.github/workflows/desktop-packages.yml) 在 macOS、Windows 和 Linux runner 上构建并上传 artifact。

## Deployment Docs

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

## Development Status

Implementation is progressing on the `dev` branch with Phase 9 acceptance complete and Phase 10/11 release plus CI work in progress.
