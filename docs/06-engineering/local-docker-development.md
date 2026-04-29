# 本地多容器开发与联调环境

## 1. 目标

本文档定义 Bifrost 本地开发时如何通过多个 Docker 容器模拟真实部署环境，验证客户端、后台、网关、数据库和多个私有服务之间的联调效果。

目标是让开发者在本机就能验证：

- 后台配置服务目录
- 网关访问多个上游容器
- 客户端登录并打开服务
- 不同授权策略产生不同访问结果
- 审计日志能记录完整链路

## 2. 本地环境组件

建议本地开发环境包含：

```text
PostgreSQL
Gateway Server
Admin Web
Mock GitLab
Mock Jenkins
Mock Docs
Mock Internal Admin
```

可选：

```text
Reverse Proxy
Mailpit
Observability Stack
```

第一阶段不强制引入 Redis。

## 3. Docker Compose 拓扑

推荐拓扑：

```text
host machine
  |
  |-- desktop client
  |-- browser
  |
docker network: bifrost-dev
  |
  |-- postgres
  |-- gateway
  |-- admin-web
  |-- mock-gitlab
  |-- mock-jenkins
  |-- mock-docs
```

网关通过 Docker 内部 DNS 访问上游：

```text
http://mock-gitlab:8080
http://mock-jenkins:8080
http://mock-docs:8080
```

宿主机访问本地 Docker 调试环境：

```text
http://localhost:8080
http://localhost:5173
```

具体端口在实现阶段确定，但必须避免常见冲突并写入 `.env.example`。

## 4. 为什么使用多个上游容器

多容器上游能验证真实问题：

- 不同服务 key 映射到不同上游
- 网关路径改写是否正确
- 不同服务授权是否隔离
- 上游服务不可达时错误是否正确
- 审计是否记录 serviceId

如果只使用一个 mock 服务，很难发现服务目录和转发隔离问题。

## 5. Mock 服务设计

每个 mock 服务建议提供：

```text
GET /
GET /health
GET /whoami
GET /headers
GET /slow
POST /echo
```

用途：

- `/`：显示服务名称
- `/health`：健康检查
- `/whoami`：返回服务标识
- `/headers`：检查网关注入 header
- `/slow`：测试超时
- `/echo`：测试请求体转发

Mock 服务可以使用轻量 HTTP 服务实现，具体代码在实现阶段再写。

## 6. 本地服务目录种子数据

本地开发建议默认创建：

| 服务 key | 名称 | 上游 |
|---|---|---|
| `gitlab` | GitLab | `http://mock-gitlab:8080` |
| `jenkins` | Jenkins | `http://mock-jenkins:8080` |
| `docs` | Docs | `http://mock-docs:8080` |

角色建议：

| 角色 | 可访问服务 |
|---|---|
| `developer` | `gitlab`, `docs` |
| `ops` | `jenkins`, `docs` |
| `admin` | 全部 |

用户建议：

| 用户 | 角色 | 特殊覆盖 |
|---|---|---|
| `alice` | `developer` | 无 |
| `bob` | `ops` | deny `jenkins` |
| `admin` | `admin` | 无 |

默认种子密码：

- `admin` / `alice` / `bob` 初始密码统一为 `ChangeMe123!`

## 7. 联调场景

## 7.1 正常访问 GitLab

1. 使用 `alice` 登录客户端
2. 客户端服务列表包含 `GitLab` 与 `Docs`
3. 点击 `GitLab`
4. 浏览器打开网关入口
5. 网关转发到 `mock-gitlab`
6. 审计记录 `GATEWAY_SERVICE_ACCESS`

## 7.2 未授权访问 Jenkins

1. 使用 `alice` 登录客户端
2. `Jenkins` 不出现在服务列表
3. 强行访问 `/s/jenkins`
4. 网关返回 `POLICY_ACCESS_DENIED`
5. 审计记录拒绝事件

## 7.3 用户级 deny 覆盖角色授权

1. 使用 `bob` 登录客户端
2. `bob` 角色是 `ops`
3. `ops` 默认可访问 `jenkins`
4. 用户级 deny 禁止 `bob` 访问 `jenkins`
5. 访问 `jenkins` 返回 `POLICY_USER_DENIED`

## 7.4 服务禁用

1. 管理员禁用 `docs`
2. 所有用户访问 `/s/docs` 都返回 `SERVICE_DISABLED`
3. 审计记录拒绝事件

## 7.5 上游不可达

1. 停止 `mock-jenkins` 容器
2. 授权用户访问 `/s/jenkins`
3. 网关返回 `SERVICE_UPSTREAM_UNREACHABLE`
4. HTTP 状态码为 `502`

## 8. 测试执行策略

本地开发建议提供三类命令：

```text
dev:infra
dev:apps
test:e2e
```

职责：

- `dev:infra`：启动 PostgreSQL 与 mock 上游
- `dev:apps`：启动 Gateway、Admin、Desktop renderer
- `test:e2e`：运行端到端测试

具体命令在实现阶段写入 `package.json`、`turbo.json`、Makefile 或 Taskfile。

当前默认本地开发闭环使用远端 dev 后端：

```bash
pnpm dev:backend
pnpm --filter @bifrost/admin dev
```

`pnpm dev:backend` 不启动 Docker，也不会拉取镜像；它只检查 `http://142.171.208.80:18080` 的 Gateway、readyz 和几个私有 upstream 代理是否可用。这样本地开发 Admin 或 Desktop 时，接口统一走已经由 `dev` 分支自动部署好的远端后端。

如需本机隔离调试后端或跑完整后端回归，再显式使用本地 Docker 命令：

```bash
pnpm dev:backend:local
pnpm dev:backend:local:down
pnpm dev:backend:down
pnpm test:backend
pnpm test:e2e
pnpm test:e2e:up
pnpm test:e2e:down
```

其中 `dev:backend:local` 用于不依赖客户端 UI 的本地 Docker 联调与回归：

1. `pnpm dev:backend:local`
2. 启动 PostgreSQL、多个 mock 上游和 Gateway。
3. 自动执行 migration 与种子数据初始化。
4. 保留环境，便于你直接调试 Gateway API、策略和代理链路。

```text
Gateway:   http://127.0.0.1:18080
Postgres:  127.0.0.1:15432
```

`pnpm dev:backend:local:down` 和 `pnpm dev:backend:down` 用于回收上述本地 Docker 后端环境。

`pnpm test:backend` 用于一键执行后端闭环验证，会自动：

1. 清理旧容器与卷。
2. 启动后端专用 Docker 环境。
3. 运行基础 infra 校验。
4. 运行 Gateway Go 服务层测试。
5. 运行 Playwright API / 代理链路 E2E。
6. 自动回收环境。

其中 `pnpm test:e2e` 会自动执行：

1. 清理旧测试容器与数据库卷。
2. 启动 PostgreSQL、mock 上游、Gateway 与 Admin Web。
3. 执行 Playwright E2E。
4. 回收测试容器与卷。

适合 CI 与本地全量回归。`pnpm test:e2e:up` / `down` 仍保留给聚焦调试使用。

默认测试端口与普通开发端口隔离：

| 用途 | 默认端口 |
|---|---:|
| PostgreSQL | `15432` |
| Gateway | `18080` |
| Admin Web | `15173` |

如需覆盖端口，可在执行命令前设置：

```bash
BIFROST_DEV_GATEWAY_PORT=28080 BIFROST_DEV_ADMIN_PORT=25173 BIFROST_DEV_POSTGRES_PORT=25432 pnpm test:e2e:up
```

E2E 启动脚本会执行：

1. 启动 PostgreSQL 与多个 mock 上游容器。
2. 等待 Docker healthcheck 成功。
3. 执行数据库 migration 与种子数据初始化。
4. 重新构建并启动 Gateway 与 Admin Web 容器。
5. 等待 `readyz` 与 Admin health 通过。

## 9. 数据初始化

本地环境必须支持一键初始化：

- 运行数据库 migration
- 创建管理员
- 创建测试用户
- 创建角色
- 创建服务目录
- 创建授权关系
- 创建用户级覆盖

初始化脚本必须幂等。

## 10. 环境变量建议

本地开发至少需要：

```text
BIFROST_PUBLIC_BASE_URL
BIFROST_ADMIN_BASE_URL
BIFROST_DATABASE_URL
BIFROST_ACCESS_TOKEN_TTL
BIFROST_REFRESH_TOKEN_TTL
BIFROST_AUDIT_RETENTION_DAYS
```

敏感信息不得提交到仓库。

## 11. 客户端联调注意事项

客户端运行在宿主机，不在 Docker 网络内。

因此客户端访问：

```text
http://localhost:<gateway-port>
```

网关访问上游：

```text
http://mock-gitlab:8080
```

这能模拟真实情况下“用户只看到网关，网关看到内网服务”的访问边界。

## 12. CI 中的多容器测试

后续 CI 可使用 Docker Compose 启动：

- PostgreSQL
- Gateway
- Mock services

然后执行：

- Go 集成测试
- API 契约测试
- Playwright E2E

CI 不一定运行完整 Electron 三端测试，但必须覆盖 Web 管理后台和网关访问核心链路。

## 13. 验收标准

本地多容器环境应证明：

- 网关能区分多个上游服务
- 服务授权隔离有效
- 禁用用户、设备、服务后访问立即失败
- 上游故障有正确错误码
- 审计日志记录完整
- 客户端不需要修改系统代理、DNS 或路由
