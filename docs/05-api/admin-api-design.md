# 后台 API 列表设计

## 1. 目标

本文件定义 Bifrost 管理后台第一阶段需要的 API 分组、路径风格、请求语义和响应结构。

后台 API 面向 Web 管理端，用于管理：

- 用户
- 角色
- 设备
- 私有服务目录
- 服务访问策略
- 审计日志
- 管理员会话

所有接口必须遵守统一响应结构，详见 [统一 API 响应结构规范](/Users/kittors/Developer/opensource/Bifrost/docs/05-api/api-response-standard.md)。

## 2. 路径规范

统一前缀：

```text
/api/v1/admin
```

路径设计原则：

- 使用名词复数
- 使用 HTTP Method 表达动作
- 避免在路径中堆砌动词
- 对状态变更类动作允许使用清晰的动作子路径

示例：

```text
GET    /api/v1/admin/users
POST   /api/v1/admin/users
PATCH  /api/v1/admin/users/{userId}
POST   /api/v1/admin/users/{userId}/disable
```

## 3. 通用请求要求

所有后台 API 请求必须携带：

- 管理员访问令牌
- `Content-Type: application/json`，文件下载等特殊接口除外
- 可选 `X-Request-Id`，若未提供由服务端生成

服务端必须返回：

- `meta.requestId`
- `meta.timestamp`

## 4. 管理员认证 API

### 4.1 登录

```text
POST /api/v1/admin/auth/login
```

用途：

- 管理员登录后台

请求体：

```json
{
  "username": "admin",
  "password": "********"
}
```

成功响应 `data`：

```json
{
  "accessToken": "access_token",
  "refreshToken": "refresh_token",
  "expiresIn": 900,
  "user": {
    "id": "user_01",
    "username": "admin",
    "displayName": "Administrator",
    "roles": ["role_admin"]
  }
}
```

### 4.2 刷新会话

```text
POST /api/v1/admin/auth/refresh
```

用途：

- 使用刷新令牌换取新的访问令牌

### 4.3 退出登录

```text
POST /api/v1/admin/auth/logout
```

用途：

- 主动注销当前后台会话

### 4.4 获取当前用户

```text
GET /api/v1/admin/auth/me
```

用途：

- 获取当前登录管理员信息、角色、权限摘要

## 5. 用户管理 API

### 5.1 用户列表

```text
GET /api/v1/admin/users
```

查询参数：

| 参数 | 类型 | 说明 |
|---|---|---|
| `page` | number | 页码，从 1 开始 |
| `pageSize` | number | 每页数量 |
| `keyword` | string | 用户名、显示名、邮箱模糊搜索 |
| `status` | string | `enabled` 或 `disabled` |
| `roleId` | string | 按角色筛选 |

成功响应：

- `data.items` 为用户列表
- `meta.pagination` 为分页信息

### 5.2 创建用户

```text
POST /api/v1/admin/users
```

请求体：

```json
{
  "username": "alice",
  "displayName": "Alice",
  "email": "alice@example.com",
  "password": "********",
  "roleIds": ["role_developer"]
}
```

### 5.3 用户详情

```text
GET /api/v1/admin/users/{userId}
```

返回：

- 基础信息
- 角色列表
- 用户级服务覆盖规则
- 设备摘要
- 最近登录信息

### 5.4 更新用户

```text
PATCH /api/v1/admin/users/{userId}
```

可更新字段：

- `displayName`
- `email`
- `roleIds`

### 5.5 重置密码

```text
POST /api/v1/admin/users/{userId}/reset-password
```

请求体：

```json
{
  "password": "********"
}
```

### 5.6 禁用用户

```text
POST /api/v1/admin/users/{userId}/disable
```

效果：

- 用户无法继续登录
- 可选择吊销其所有会话

### 5.7 启用用户

```text
POST /api/v1/admin/users/{userId}/enable
```

### 5.8 删除用户

```text
DELETE /api/v1/admin/users/{userId}
```

建议：

- 第一阶段可以实现软删除
- 至少保留审计关联信息

## 6. 角色管理 API

### 6.1 角色列表

```text
GET /api/v1/admin/roles
```

查询参数：

| 参数 | 类型 | 说明 |
|---|---|---|
| `page` | number | 页码 |
| `pageSize` | number | 每页数量 |
| `keyword` | string | 角色名搜索 |

### 6.2 创建角色

```text
POST /api/v1/admin/roles
```

请求体：

```json
{
  "name": "developer",
  "displayName": "研发人员",
  "description": "可访问研发相关服务"
}
```

### 6.3 角色详情

```text
GET /api/v1/admin/roles/{roleId}
```

返回：

- 角色基础信息
- 已授权服务
- 关联用户数量

### 6.4 更新角色

```text
PATCH /api/v1/admin/roles/{roleId}
```

### 6.5 删除角色

```text
DELETE /api/v1/admin/roles/{roleId}
```

若角色仍被用户使用，应返回：

- HTTP `409`
- 错误码 `ROLE_IN_USE`

### 6.6 设置角色可访问服务

```text
PUT /api/v1/admin/roles/{roleId}/services
```

请求体：

```json
{
  "serviceIds": ["svc_gitlab", "svc_jenkins"]
}
```

## 7. 用户级访问覆盖 API

用户级访问覆盖用于处理例外授权，优先级高于角色授权。

### 7.1 获取用户服务覆盖

```text
GET /api/v1/admin/users/{userId}/service-overrides
```

### 7.2 设置用户服务覆盖

```text
PUT /api/v1/admin/users/{userId}/service-overrides
```

请求体：

```json
{
  "allowServiceIds": ["svc_gitlab"],
  "denyServiceIds": ["svc_jenkins"]
}
```

规则：

- `denyServiceIds` 优先级最高
- 同一服务不允许同时出现在 allow 与 deny 中
- 若出现冲突，返回 `POLICY_RULE_INVALID`

## 8. 设备管理 API

### 8.1 设备列表

```text
GET /api/v1/admin/devices
```

查询参数：

| 参数 | 类型 | 说明 |
|---|---|---|
| `page` | number | 页码 |
| `pageSize` | number | 每页数量 |
| `keyword` | string | 设备名、用户、指纹搜索 |
| `status` | string | `trusted`、`disabled` |
| `userId` | string | 按用户筛选 |

### 8.2 设备详情

```text
GET /api/v1/admin/devices/{deviceId}
```

返回：

- 设备基础信息
- 绑定用户
- 公钥指纹
- 最近登录时间
- 最近访问服务

### 8.3 禁用设备

```text
POST /api/v1/admin/devices/{deviceId}/disable
```

效果：

- 该设备无法继续登录或访问服务
- 应吊销该设备关联会话

### 8.4 启用设备

```text
POST /api/v1/admin/devices/{deviceId}/enable
```

### 8.5 解绑设备

```text
DELETE /api/v1/admin/devices/{deviceId}
```

建议：

- 删除设备绑定关系
- 保留历史审计记录

## 9. 服务目录 API

### 9.1 服务列表

```text
GET /api/v1/admin/services
```

查询参数：

| 参数 | 类型 | 说明 |
|---|---|---|
| `page` | number | 页码 |
| `pageSize` | number | 每页数量 |
| `keyword` | string | 服务名、标识搜索 |
| `status` | string | `enabled` 或 `disabled` |
| `group` | string | 服务分组 |

### 9.2 创建服务

```text
POST /api/v1/admin/services
```

请求体：

```json
{
  "key": "gitlab",
  "name": "GitLab",
  "description": "研发代码仓库",
  "group": "development",
  "protocol": "https",
  "upstreamUrl": "http://10.0.0.12:8929",
  "publicPath": "/s/gitlab",
  "enabled": true
}
```

约束：

- `key` 必须唯一
- `upstreamUrl` 只能由管理员配置
- 客户端不能传入或覆盖上游地址

### 9.3 服务详情

```text
GET /api/v1/admin/services/{serviceId}
```

### 9.4 更新服务

```text
PATCH /api/v1/admin/services/{serviceId}
```

可更新：

- 显示名
- 描述
- 分组
- 上游地址
- 公共路径
- 启用状态

### 9.5 启用服务

```text
POST /api/v1/admin/services/{serviceId}/enable
```

### 9.6 禁用服务

```text
POST /api/v1/admin/services/{serviceId}/disable
```

### 9.7 删除服务

```text
DELETE /api/v1/admin/services/{serviceId}
```

若服务已经有审计记录，建议软删除或标记归档。

## 10. 服务授权 API

### 10.1 获取服务授权摘要

```text
GET /api/v1/admin/services/{serviceId}/access
```

返回：

- 哪些角色可访问
- 哪些用户被额外允许
- 哪些用户被明确禁止

### 10.2 设置服务角色授权

```text
PUT /api/v1/admin/services/{serviceId}/roles
```

请求体：

```json
{
  "roleIds": ["role_developer", "role_ops"]
}
```

## 11. 审计日志 API

### 11.1 审计列表

```text
GET /api/v1/admin/audit-events
```

查询参数：

| 参数 | 类型 | 说明 |
|---|---|---|
| `page` | number | 页码 |
| `pageSize` | number | 每页数量 |
| `type` | string | 事件类型 |
| `actorUserId` | string | 操作者用户 ID |
| `targetType` | string | 目标类型 |
| `targetId` | string | 目标 ID |
| `serviceId` | string | 服务 ID |
| `result` | string | `success` 或 `failure` |
| `from` | string | 起始时间 |
| `to` | string | 结束时间 |

### 11.2 审计详情

```text
GET /api/v1/admin/audit-events/{eventId}
```

返回：

- 事件基础信息
- 请求 ID
- 操作者
- 目标对象
- 设备信息
- 来源 IP
- 错误码
- 结构化详情

### 11.3 审计导出

```text
POST /api/v1/admin/audit-events/export
```

请求体：

```json
{
  "from": "2026-04-01T00:00:00Z",
  "to": "2026-04-17T23:59:59Z",
  "type": "SERVICE_ACCESS",
  "format": "csv"
}
```

约束：

- 导出范围必须限制
- 范围过大返回 `AUDIT_EXPORT_TOO_LARGE`

## 12. 仪表盘 API

### 12.1 管理后台概览

```text
GET /api/v1/admin/dashboard/summary
```

返回建议：

```json
{
  "users": {
    "total": 42,
    "disabled": 3
  },
  "devices": {
    "total": 78,
    "disabled": 5
  },
  "services": {
    "total": 12,
    "disabled": 1
  },
  "access": {
    "todayTotal": 1360,
    "todayDenied": 28
  }
}
```

### 12.2 最近活动

```text
GET /api/v1/admin/dashboard/recent-events
```

用途：

- 后台首页展示最近登录、服务访问、策略拒绝、配置变更

## 13. 系统配置 API

第一阶段只做必要配置，不引入复杂配置中心。

### 13.1 获取系统设置

```text
GET /api/v1/admin/settings
```

### 13.2 更新系统设置

```text
PATCH /api/v1/admin/settings
```

可配置项建议：

- 会话过期时间
- 设备绑定数量上限
- 审计保留策略
- 服务访问默认策略

## 14. 健康检查 API

健康检查可以不挂在 admin 前缀下。

```text
GET /healthz
GET /readyz
```

用途：

- `healthz`：进程是否存活
- `readyz`：数据库、关键依赖是否可用

## 15. 权限要求

后台 API 本身也需要权限控制。

第一阶段建议管理权限先分为：

- `admin:read`
- `admin:write`
- `users:read`
- `users:write`
- `roles:read`
- `roles:write`
- `devices:read`
- `devices:write`
- `services:read`
- `services:write`
- `audit:read`
- `settings:write`

后续可根据角色模型细化。

## 16. 审计要求

以下后台 API 操作必须写审计：

- 创建用户
- 禁用用户
- 重置密码
- 创建角色
- 修改角色服务授权
- 禁用设备
- 解绑设备
- 创建服务
- 修改服务上游地址
- 启用或禁用服务
- 设置用户级访问覆盖
- 修改系统设置
- 导出审计日志

## 17. 第一阶段 API 优先级

第一阶段优先实现：

1. 管理员登录与当前用户
2. 用户管理
3. 角色管理
4. 服务目录管理
5. 角色服务授权
6. 用户级访问覆盖
7. 设备管理
8. 审计列表与详情
9. 仪表盘摘要

暂不优先实现：

- 复杂系统设置
- 审计异步导出任务
- 多租户管理
- 细粒度审批流
