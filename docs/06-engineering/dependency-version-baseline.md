# 依赖版本基线

## 1. 说明

本文件记录 Bifrost 当前推荐的技术栈版本基线。目标是在实现前先锁定“推荐且稳定”的依赖组合，减少后续初始化工程时的版本漂移。

以下版本基线以 `2026-04-17` 为准，已结合官方文档、官方发布渠道或官方包注册表信息进行核对。

## 2. 运行时与工程基础

| 名称 | 版本 |
|---|---:|
| Go | `1.26.2` |
| Node.js LTS | `24.15.0` |
| pnpm | `10.33.0` |
| Turborepo | `2.9.6` |
| Biome | `2.4.12` |
| Changesets | `2.30.0` |
| Lefthook | `2.1.6` |

## 3. 桌面客户端与后台前端

| 名称 | 版本 |
|---|---:|
| Electron | `41.2.1` |
| electron-vite | `6.0.0-beta.1` |
| electron-builder | `26.8.1` |
| React | `19.2.5` |
| react-dom | `19.2.5` |
| TypeScript | `6.0.3` |
| Vite | `8.0.8` |
| @vitejs/plugin-react | `6.0.1` |
| Tailwind CSS | `4.2.2` |
| shadcn CLI | `4.3.0` |
| @tanstack/react-query | `5.99.0` |
| @tanstack/react-router | `1.168.22` |
| react-hook-form | `7.72.1` |
| @hookform/resolvers | `5.2.2` |
| zod | `4.3.6` |
| lucide-react | `1.8.0` |
| sonner | `2.0.7` |
| zustand | `5.0.12` |
| Vitest | `4.1.4` |
| Playwright | `1.59.1` |
| @types/node | `25.6.0` |

## 4. Go 服务端

| 名称 | 版本 |
|---|---:|
| github.com/go-chi/chi/v5 | `v5.2.5` |
| github.com/jackc/pgx/v5 | `v5.9.1` |
| github.com/pressly/goose/v3 | `v3.9.0` |
| github.com/oapi-codegen/oapi-codegen/v2 | `v2.6.0` |
| github.com/rs/zerolog | `v1.35.0` |

## 5. 数据存储建议

推荐：

- PostgreSQL：使用当前官方主线稳定版的 `18.x`

第一阶段默认不强制引入 Redis，原因：

- 当前业务核心是认证、授权、审计、服务目录与网关转发
- 引入 Redis 会增加部署和运维复杂度
- 初期完全可以先以 PostgreSQL 为主

在以下场景成熟后再引入 Redis：

- 高频限流
- 分布式会话缓存
- 实时在线状态
- 复杂事件消费

## 6. 版本策略

### 6.1 更新原则

- 核心基础设施优先使用官方稳定版本
- Node 优先选择 LTS 版本，而不是最新非 LTS
- Electron 优先使用当前稳定版本，并关注安全更新
- Tailwind、React、TypeScript 等核心前端依赖要成组评估升级

### 6.2 升级要求

升级以下依赖前必须做专项验证：

- Electron
- React
- Tailwind CSS
- TypeScript
- Vite
- Go

验证内容包括：

- 构建是否正常
- 共享组件是否兼容
- 暗色主题是否异常
- Electron 安全行为是否变化
- 契约生成链是否正常

## 7. 版本来源说明

本基线来源于以下官方渠道或官方注册表信息：

- Go 官方下载页
- Node.js 官方发布渠道
- Electron 官方文档与 npm 包版本
- React 官方文档与 npm 包版本
- Tailwind CSS 官方文档与 npm 包版本
- shadcn/ui 官方文档与 npm 包版本
- npm registry
- GitHub 官方仓库 tag

在进入真正的项目初始化阶段前，应再做一次版本复核，以确保版本仍为当时的稳定推荐值。
