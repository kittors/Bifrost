# Phase 13 Backend Validation Environment Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans or superpowers:subagent-driven-development to implement this plan task-by-task.

**Goal:** 不依赖客户端 UI，直接在本地构建一套可重复执行的后端验证环境，让开发者可以一键启动 PostgreSQL、Gateway 和多个 mock 私有服务，并完成服务层、接口层和代理链路回归。

**Architecture:** 继续复用现有 Docker Compose、数据库 migration / seed、Gateway Go 测试和 Playwright API/E2E 基座，但新增后端专用启动与回归脚本，不再强依赖 Admin Web 或 Desktop UI。

**Tech Stack:** Docker Compose、PostgreSQL 18.x、Go 1.26.2、Playwright 1.59.1、pnpm 10.33.0、Node.js 24。

---

### Task 1: Add Backend-Only Script Contracts

**Files:**
- Modify: `tests/infra/docker-compose.test.mjs`
- Modify: `package.json`

- [x] **Step 1: Write failing infra assertions**

Add failing infra assertions for:

- `pnpm dev:backend`
- `pnpm dev:backend:down`
- `pnpm test:backend`

Run: `pnpm test:infra`

Expected: FAIL until scripts and files exist.

- [x] **Step 2: Add script entries**

Add root script entries pointing to dedicated backend helpers.

Run: `pnpm test:infra`

Expected: PASS.

### Task 2: Implement Backend-Only Docker Startup

**Files:**
- Create: `scripts/testing/backend-up.mjs`

- [x] **Step 1: Implement startup flow**

Requirements:

- start `postgres`
- start `mock-gitlab`
- start `mock-jenkins`
- start `mock-docs`
- start `mock-internal-admin`
- run existing migration / seed helper
- start `gateway`
- wait for `readyz`

Run: `pnpm dev:backend`

Expected: environment stays up and `Gateway` is reachable on `http://127.0.0.1:18080`.

### Task 3: Implement Backend Full Validation Runner

**Files:**
- Create: `scripts/testing/backend-run.mjs`

- [x] **Step 1: Implement clean full-run flow**

Requirements:

- clean previous containers
- start backend environment
- run `pnpm test:infra`
- run `go test ./...` in `apps/gateway`
- run `pnpm exec playwright test`
- always clean containers on exit

Run: `pnpm test:backend`

Expected: PASS.

### Task 4: Update Engineering Docs And Checklist

**Files:**
- Modify: `docs/06-engineering/local-docker-development.md`
- Modify: `docs/06-engineering/testing-strategy.md`
- Modify: `docs/08-roadmap/development-checklist.md`

- [x] **Step 1: Document backend-only entrypoints**

Explain:

- what `dev:backend` starts
- what `dev:backend:down` cleans
- what `test:backend` validates
- why this path avoids UI dependency

Run:

- `pnpm lint`
- `pnpm check`

Expected: PASS.
