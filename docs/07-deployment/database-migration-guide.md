# 数据库 Migration 执行说明

Bifrost 使用 Go migration 命令管理数据库结构和第一阶段种子数据。

## 环境变量

```bash
export BIFROST_DATABASE_URL='postgres://bifrost:secret@postgres:5432/bifrost?sslmode=require'
```

## 初始化或升级

```bash
pnpm db:migrate
```

该命令会进入 `apps/gateway` 并执行：

```bash
go run ./cmd/bifrost-migrate up
```

## 写入种子数据

```bash
pnpm db:seed
```

种子数据包含：

- 管理员账号
- `developer`、`ops`、`admin` 角色
- `gitlab`、`jenkins`、`docs` 服务目录
- `alice`、`bob` 测试用户
- 角色服务授权与用户级 deny 示例

## 回滚

```bash
pnpm db:reset
```

该命令用于本地和测试环境重置数据库，不建议直接用于生产。生产回滚应优先恢复备份或执行经过审阅的 down migration。

## 发布顺序

1. 备份 PostgreSQL。
2. 停止写入流量或进入维护窗口。
3. 执行 `pnpm db:migrate`。
4. 启动新版本 Gateway。
5. 验证 `/readyz`。
6. 验证 Admin Web 登录和服务列表。
7. 验证客户端登录和一个授权服务访问。

## 失败处理

- migration 失败时，保留日志和数据库快照。
- 不要在不理解失败原因时反复运行 reset。
- 如果已经写入部分结构，先确认 goose migration 状态，再决定恢复备份或补修 migration。
