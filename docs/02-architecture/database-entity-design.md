# 数据库实体与表结构设计

## 1. 目标

本文档描述 Bifrost 第一阶段的核心实体、关系和表结构建议。具体 SQL migration 在实现阶段再编写，本文件用于先统一数据模型。

推荐数据库：

- PostgreSQL `18.x`

## 2. 设计原则

- 使用稳定字符串 ID，便于跨端展示与审计引用
- 核心业务表保留创建时间和更新时间
- 高风险对象采用软删除或归档策略
- 审计日志不可被普通业务删除
- 所有关键状态使用明确枚举

## 3. 核心实体

第一阶段核心实体包括：

- User
- Role
- UserRole
- Device
- Service
- RoleService
- UserServiceOverride
- Session
- DeviceChallenge
- AuditEvent
- SystemSetting

## 4. users

用途：

- 存储平台本地账号

建议字段：

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | text | 用户 ID |
| `username` | text | 登录名，唯一 |
| `display_name` | text | 显示名 |
| `email` | text | 邮箱，可选 |
| `password_hash` | text | 密码散列 |
| `status` | text | `enabled`、`disabled` |
| `created_at` | timestamptz | 创建时间 |
| `updated_at` | timestamptz | 更新时间 |
| `deleted_at` | timestamptz | 软删除时间 |

约束：

- `username` 唯一
- 禁用用户不能登录
- 删除用户不应破坏审计关联

## 5. roles

用途：

- 存储角色

建议字段：

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | text | 角色 ID |
| `name` | text | 角色名称，唯一 |
| `display_name` | text | 显示名称 |
| `description` | text | 描述 |
| `is_system` | boolean | 是否系统角色 |
| `created_at` | timestamptz | 创建时间 |
| `updated_at` | timestamptz | 更新时间 |

## 6. user_roles

用途：

- 用户与角色的多对多关系

建议字段：

| 字段 | 类型 | 说明 |
|---|---|---|
| `user_id` | text | 用户 ID |
| `role_id` | text | 角色 ID |
| `created_at` | timestamptz | 关联时间 |

约束：

- `(user_id, role_id)` 唯一

## 7. devices

用途：

- 存储用户绑定设备

建议字段：

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | text | 设备 ID |
| `user_id` | text | 绑定用户 |
| `name` | text | 设备名称 |
| `os` | text | 操作系统 |
| `client_version` | text | 客户端版本 |
| `public_key` | text | 设备公钥 |
| `public_key_fingerprint` | text | 公钥指纹 |
| `status` | text | `trusted`、`disabled` |
| `last_seen_at` | timestamptz | 最近活跃时间 |
| `created_at` | timestamptz | 创建时间 |
| `updated_at` | timestamptz | 更新时间 |

约束：

- `public_key_fingerprint` 唯一
- 禁用设备不能登录或访问服务

## 8. services

用途：

- 存储私有服务目录

建议字段：

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | text | 服务 ID |
| `key` | text | 服务标识，唯一 |
| `name` | text | 服务显示名 |
| `description` | text | 描述 |
| `group_name` | text | 分组 |
| `protocol` | text | `http` 或 `https` |
| `upstream_url` | text | 内网上游地址 |
| `public_path` | text | 网关路径，例如 `/s/gitlab` |
| `status` | text | `enabled`、`disabled`、`archived` |
| `created_at` | timestamptz | 创建时间 |
| `updated_at` | timestamptz | 更新时间 |

约束：

- `key` 唯一
- `public_path` 唯一
- 上游地址只能由管理员配置

## 9. role_services

用途：

- 角色可访问服务关系

建议字段：

| 字段 | 类型 | 说明 |
|---|---|---|
| `role_id` | text | 角色 ID |
| `service_id` | text | 服务 ID |
| `created_at` | timestamptz | 授权时间 |

约束：

- `(role_id, service_id)` 唯一

## 10. user_service_overrides

用途：

- 用户级访问覆盖

建议字段：

| 字段 | 类型 | 说明 |
|---|---|---|
| `user_id` | text | 用户 ID |
| `service_id` | text | 服务 ID |
| `effect` | text | `allow` 或 `deny` |
| `reason` | text | 原因 |
| `created_by` | text | 操作者 |
| `created_at` | timestamptz | 创建时间 |

约束：

- `(user_id, service_id)` 唯一
- `deny` 优先级高于角色授权

## 11. sessions

用途：

- 存储用户会话与吊销状态

建议字段：

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | text | 会话 ID |
| `user_id` | text | 用户 ID |
| `device_id` | text | 设备 ID |
| `refresh_token_hash` | text | 刷新令牌散列 |
| `status` | text | `active`、`revoked`、`expired` |
| `expires_at` | timestamptz | 过期时间 |
| `created_at` | timestamptz | 创建时间 |
| `revoked_at` | timestamptz | 吊销时间 |

## 12. device_challenges

用途：

- 存储设备签名挑战，支持服务端确认客户端持有设备私钥

建议字段：

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | text | 挑战 ID |
| `device_id` | text | 设备 ID |
| `challenge` | text | base64url 编码的随机挑战 |
| `expires_at` | timestamptz | 过期时间 |
| `verified_at` | timestamptz | 验证成功时间 |
| `created_at` | timestamptz | 创建时间 |

约束：

- `device_id` 必须引用已存在设备
- 挑战必须短期有效，第一阶段默认 120 秒
- 已验证或过期挑战不得重复使用

## 13. audit_events

用途：

- 存储审计事件

建议字段：

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | text | 审计事件 ID |
| `request_id` | text | 请求 ID |
| `type` | text | 事件类型 |
| `actor_user_id` | text | 操作者用户 |
| `actor_device_id` | text | 操作者设备 |
| `target_type` | text | 目标类型 |
| `target_id` | text | 目标 ID |
| `service_id` | text | 服务 ID |
| `result` | text | `success` 或 `failure` |
| `error_code` | text | 错误码 |
| `source_ip` | inet | 来源 IP |
| `user_agent` | text | User-Agent |
| `summary` | text | 摘要 |
| `details` | jsonb | 结构化详情 |
| `occurred_at` | timestamptz | 发生时间 |

## 14. system_settings

用途：

- 存储少量系统级配置

建议字段：

| 字段 | 类型 | 说明 |
|---|---|---|
| `key` | text | 配置键 |
| `value` | jsonb | 配置值 |
| `updated_by` | text | 更新人 |
| `updated_at` | timestamptz | 更新时间 |

## 15. 索引建议

建议索引：

- `users(username)`
- `devices(user_id)`
- `devices(public_key_fingerprint)`
- `device_challenges(device_id)`
- `device_challenges(expires_at)`
- `services(key)`
- `services(public_path)`
- `audit_events(occurred_at)`
- `audit_events(type, occurred_at)`
- `audit_events(actor_user_id, occurred_at)`
- `audit_events(service_id, occurred_at)`
- `sessions(user_id, device_id)`

## 16. 数据保留策略

- 用户、设备、服务删除优先软删除或归档
- 审计日志不随业务对象删除
- 审计保留周期由系统设置控制

## 17. 后续扩展预留

未来可增加：

- identity_providers，用于 OIDC/LDAP
- service_protocols，用于 SSH/TCP
- approval_requests，用于临时授权审批
- risk_events，用于风险检测
