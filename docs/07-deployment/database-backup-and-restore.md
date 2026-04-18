# 数据库备份与恢复说明

PostgreSQL 是 Bifrost 第一阶段唯一持久化依赖，保存账号、角色、设备、会话、服务目录、策略和审计日志。

## 备份范围

必须备份：

- `users`
- `roles`
- `user_roles`
- `devices`
- `services`
- `role_services`
- `user_service_overrides`
- `sessions`
- `audit_events`
- `system_settings`

## 备份命令

```bash
pg_dump \
  --format=custom \
  --file=bifrost-$(date +%Y%m%d-%H%M%S).dump \
  "$BIFROST_DATABASE_URL"
```

## 恢复命令

```bash
createdb bifrost_restore
pg_restore \
  --dbname=postgres://bifrost:secret@postgres:5432/bifrost_restore?sslmode=require \
  --clean \
  --if-exists \
  bifrost-YYYYMMDD-HHMMSS.dump
```

恢复后先在隔离环境验证：

- 管理员可以登录。
- 客户端可以登录。
- 服务目录和授权关系正确。
- 最近审计事件可查询。

## 备份频率

| 环境 | 频率 | 保留建议 |
|---|---|---|
| 生产 | 每日全量，关键变更前手动备份 | 至少 30 天 |
| 预发 | 每周或按需 | 至少 7 天 |
| 本地 | 不强制 | 按开发需要 |

## 演练要求

生产上线后每月至少做一次恢复演练，并记录：

- 备份文件名
- 恢复耗时
- 校验账号
- 校验服务
- 问题和改进项
