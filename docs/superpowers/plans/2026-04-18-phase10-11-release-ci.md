# Phase 10-11 Release And CI Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 完成第一阶段发布准备与 CI 门禁，让 Bifrost 具备真实的服务端镜像、后台镜像、三端桌面安装包产物和 GitHub Actions 自动校验能力。

**Architecture:** 继续沿用 Monorepo 单仓工作流。Gateway 保持独立 Go 容器镜像，Admin Web 从当前占位静态服务切换到真实 Vite 构建产物镜像，Desktop 使用 `electron-builder` 输出三端安装包。CI 分成一个主质量门禁工作流和一个桌面产物工作流，避免把所有任务塞进单个 job。

**Tech Stack:** Go 1.26.2、Node.js 24.11.1、pnpm 10.33.0、Electron 41.2.1、electron-builder 26.8.1、Playwright 1.59.1、GitHub Actions、Docker BuildKit。

---

### Task 1: Replace Admin Preview Image With Real Build Image

**Files:**
- Modify: `docker/admin-web/Dockerfile`
- Create: `docker/admin-web/nginx.conf`
- Modify: `docker-compose.yml`
- Test: `tests/infra/docker-compose.test.mjs`

- [ ] **Step 1: 让 infra 测试约束后台镜像来自真实 Admin 应用目录**

更新 `tests/infra/docker-compose.test.mjs`，断言 `admin-web` 构建使用 `docker/admin-web/Dockerfile` 且运行健康检查仍指向 `/health`。同时增加对 `docker/admin-web/nginx.conf` 存在的断言，避免回到占位静态页。

- [ ] **Step 2: 运行 infra 测试确认当前失败**

Run: `pnpm test:infra`
Expected: FAIL，提示缺少 `nginx.conf` 或后台镜像仍是占位实现。

- [ ] **Step 3: 改造 Admin Dockerfile 为真实 Vite 构建镜像**

实现要点：

```dockerfile
FROM node:24.11.1-alpine AS builder
WORKDIR /workspace
RUN corepack enable && corepack prepare pnpm@10.33.0 --activate
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY apps/admin/package.json ./apps/admin/package.json
COPY packages ./packages
COPY apps/admin ./apps/admin
RUN pnpm install --frozen-lockfile
RUN pnpm --filter @bifrost/admin build

FROM nginx:1.29-alpine
COPY docker/admin-web/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=builder /workspace/apps/admin/dist /usr/share/nginx/html
EXPOSE 5173
```

要求：

- Nginx 监听 `5173`
- `/health` 返回 `200`
- 其他路径回退到 `index.html`

- [ ] **Step 4: 运行 infra 与 smoke 验证**

Run:
- `pnpm test:infra`
- `BIFROST_DEV_GATEWAY_PORT=28080 BIFROST_DEV_ADMIN_PORT=25173 BIFROST_DEV_POSTGRES_PORT=25432 pnpm test:e2e:up`
- `pnpm test:e2e:down`

Expected: PASS，真实 Admin 镜像能启动并通过 healthcheck。

### Task 2: Add Desktop Packaging Configuration

**Files:**
- Modify: `apps/desktop/package.json`
- Create: `apps/desktop/electron-builder.yml`
- Create: `apps/desktop/resources/README.md`
- Test: `tests/infra/docker-compose.test.mjs`

- [ ] **Step 1: 先补配置测试，约束桌面应用存在三端打包脚本**

在 `tests/infra/docker-compose.test.mjs` 新增断言：

- `@bifrost/desktop` 存在 `dist:mac`
- `@bifrost/desktop` 存在 `dist:win`
- `@bifrost/desktop` 存在 `dist:linux`
- `devDependencies.electron-builder === "26.8.1"`

- [ ] **Step 2: 跑测试确认失败**

Run: `pnpm test:infra`
Expected: FAIL，因为 desktop 还没有这些脚本和依赖。

- [ ] **Step 3: 增加 electron-builder 配置与脚本**

实现要点：

```yaml
appId: com.kittors.bifrost.desktop
productName: Bifrost
directories:
  output: release
files:
  - out/**
mac:
  target:
    - target: dmg
      arch: [universal]
    - target: zip
      arch: [universal]
win:
  target:
    - target: nsis
      arch: [x64]
linux:
  target:
    - AppImage
    - deb
    - tar.gz
```

并在 `apps/desktop/package.json` 中补：

- `main`
- `dist:mac`
- `dist:win`
- `dist:linux`
- `dist:dir`

- [ ] **Step 4: 本地验证桌面打包配置可解析**

Run:
- `pnpm install`
- `pnpm --filter @bifrost/desktop build`
- `pnpm --filter @bifrost/desktop exec electron-builder --config electron-builder.yml --dir`

Expected: PASS，本机至少能生成目录包。

### Task 3: Write Phase 10 Deployment And Operations Docs

**Files:**
- Create: `docs/07-deployment/service-runtime-parameters.md`
- Create: `docs/07-deployment/admin-startup-guide.md`
- Create: `docs/07-deployment/desktop-startup-guide.md`
- Create: `docs/07-deployment/database-migration-guide.md`
- Create: `docs/07-deployment/database-backup-and-restore.md`
- Create: `docs/07-deployment/tls-and-certificates.md`
- Create: `docs/07-deployment/logging-and-monitoring.md`
- Create: `docs/07-deployment/internal-trial-release-notes.md`
- Create: `docs/07-deployment/known-issues.md`
- Create: `docs/07-deployment/rollback-guide.md`
- Create: `docs/07-deployment/request-id-troubleshooting.md`
- Modify: `README.md`
- Modify: `docs/08-roadmap/development-checklist.md`

- [ ] **Step 1: 梳理每份运维文档的责任边界并落盘**

文档必须分别覆盖：

- Gateway 环境变量、默认值、生产建议。
- Admin 镜像启动方式、反向代理示例、健康检查。
- Desktop 初始登录、设备信任、日志位置、平台差异。
- migration 执行顺序与失败回滚。
- PostgreSQL 备份、恢复、演练周期。
- TLS 证书申请、续期、轮换。
- 日志字段、审计保留、监控指标与告警建议。
- 内部试用发布说明、已知问题、回滚入口、通过 requestId 排障的操作步骤。

- [ ] **Step 2: 更新 README 目录索引**

在根 `README.md` 中加入：

- 本地联调入口
- 部署文档索引
- CI / 打包产物说明

- [ ] **Step 3: 用 checklist 标记 Phase 10 文档项**

等文档写完并经过最少一轮命令验证后，给 Phase 10 对应条目补时间戳。

### Task 4: Add GitHub Actions CI Gate

**Files:**
- Create: `.github/workflows/ci.yml`
- Create: `.github/workflows/desktop-packages.yml`
- Modify: `README.md`
- Modify: `docs/08-roadmap/development-checklist.md`

- [ ] **Step 1: 先补测试，约束仓库存在 CI 工作流**

在 `tests/infra/docker-compose.test.mjs` 增加断言：

- `.github/workflows/ci.yml` 存在
- `.github/workflows/desktop-packages.yml` 存在

- [ ] **Step 2: 跑 infra 测试确认失败**

Run: `pnpm test:infra`
Expected: FAIL，因为 workflow 还不存在。

- [ ] **Step 3: 实现主 CI 工作流**

`ci.yml` 至少包含：

- `push` 与 `pull_request`
- checkout / setup-node / pnpm / setup-go
- `pnpm lint`
- `pnpm check`
- `pnpm test`
- `pnpm test:infra`
- `pnpm test:e2e`
- `go test ./...`

要求：

- 在 Ubuntu runner 上执行
- Docker Compose 可用于 E2E
- 失败时上传 `test-results/playwright` 工件

- [ ] **Step 4: 实现桌面产物工作流**

`desktop-packages.yml` 使用 matrix：

- `macos-latest`
- `windows-latest`
- `ubuntu-latest`

每个 job：

- 安装 Node / pnpm
- 安装依赖
- 执行 `pnpm --filter @bifrost/desktop build`
- 执行对应 `dist:mac` / `dist:win` / `dist:linux`
- 上传 `apps/desktop/release/**`

- [ ] **Step 5: 推送后用 gh 检查工作流真实结果**

Run:
- `git push origin dev`
- `gh run list --limit 10`
- `gh run watch <run-id>`

Expected: CI 绿灯，桌面产物 workflow 成功产出三端构建工件。

### Task 5: Final Verification And Checklist Closure

**Files:**
- Modify: `docs/08-roadmap/development-checklist.md`

- [ ] **Step 1: 运行完整验证**

Run:
- `pnpm lint`
- `pnpm check`
- `pnpm test`
- `pnpm test:infra`
- `BIFROST_DATABASE_TEST_URL='postgres://bifrost:bifrost@127.0.0.1:15432/postgres?sslmode=disable' go test ./...`
- `pnpm test:e2e`
- `pnpm --filter @bifrost/desktop build`
- `pnpm --filter @bifrost/desktop exec electron-builder --config electron-builder.yml --dir`
- `docker build -f apps/gateway/Dockerfile .`
- `docker build -f docker/admin-web/Dockerfile .`

Expected: PASS。

- [ ] **Step 2: 更新 Phase 10 / Phase 11 清单与里程碑**

满足条件后补时间戳：

- Gateway/Admin 镜像、部署文档、三端桌面包、内部试用说明、已知问题、回滚、排障文档。
- CI 工作流、最低门禁、失败工件保留。

## Self-Review

- Spec coverage: 覆盖了 Phase 10 的镜像、文档、客户端安装包与发布资料，以及 Phase 11 的 CI、E2E、工件保留和门禁。
- Placeholder scan: 无 `TODO` / `TBD` / “后续补充” 占位。
- Type consistency: 所有新增文件路径、脚本名与当前仓库目录结构一致；Desktop 打包统一收敛到 `apps/desktop/release`。
