# 审计事件字典

## 1. 目标

审计事件用于记录系统关键行为，支持安全追溯、管理员排障和合规检查。

本文件定义第一阶段审计事件类型、触发时机和字段要求。

## 2. 审计事件通用结构

建议结构：

```json
{
  "id": "audit_01",
  "requestId": "req_01",
  "type": "AUTH_LOGIN_SUCCESS",
  "actorUserId": "user_01",
  "actorDeviceId": "dev_01",
  "targetType": "user",
  "targetId": "user_01",
  "serviceId": null,
  "result": "success",
  "errorCode": null,
  "sourceIp": "203.0.113.10",
  "userAgent": "Bifrost Desktop/1.0.0",
  "summary": "用户登录成功",
  "details": {},
  "occurredAt": "2026-04-17T10:30:00Z"
}
```

## 3. 认证事件

| 事件类型 | 结果 | 触发时机 |
|---|---|---|
| `AUTH_LOGIN_SUCCESS` | success | 用户登录成功 |
| `AUTH_LOGIN_FAILURE` | failure | 用户登录失败 |
| `AUTH_LOGOUT` | success | 用户退出登录 |
| `AUTH_SESSION_REFRESH` | success | 刷新会话成功 |
| `AUTH_SESSION_REVOKED` | success | 会话被吊销 |

## 4. 设备事件

| 事件类型 | 结果 | 触发时机 |
|---|---|---|
| `DEVICE_REGISTERED` | success | 设备注册成功 |
| `DEVICE_CHALLENGE_FAILED` | failure | 设备挑战验签失败 |
| `DEVICE_DISABLED` | success | 管理员禁用设备 |
| `DEVICE_ENABLED` | success | 管理员启用设备 |
| `DEVICE_UNBOUND` | success | 管理员解绑设备 |

## 5. 用户事件

| 事件类型 | 结果 | 触发时机 |
|---|---|---|
| `USER_CREATED` | success | 创建用户 |
| `USER_UPDATED` | success | 更新用户 |
| `USER_DISABLED` | success | 禁用用户 |
| `USER_ENABLED` | success | 启用用户 |
| `USER_PASSWORD_RESET` | success | 重置用户密码 |
| `USER_DELETED` | success | 删除或归档用户 |

## 6. 角色事件

| 事件类型 | 结果 | 触发时机 |
|---|---|---|
| `ROLE_CREATED` | success | 创建角色 |
| `ROLE_UPDATED` | success | 更新角色 |
| `ROLE_DELETED` | success | 删除角色 |
| `ROLE_SERVICE_ACCESS_UPDATED` | success | 更新角色服务授权 |

## 7. 服务事件

| 事件类型 | 结果 | 触发时机 |
|---|---|---|
| `SERVICE_CREATED` | success | 创建服务 |
| `SERVICE_UPDATED` | success | 更新服务 |
| `SERVICE_ENABLED` | success | 启用服务 |
| `SERVICE_DISABLED` | success | 禁用服务 |
| `SERVICE_DELETED` | success | 删除或归档服务 |
| `SERVICE_UPSTREAM_UPDATED` | success | 修改服务上游地址 |

## 8. 策略事件

| 事件类型 | 结果 | 触发时机 |
|---|---|---|
| `POLICY_USER_OVERRIDE_UPDATED` | success | 更新用户级服务覆盖 |
| `POLICY_ACCESS_ALLOWED` | success | 服务访问被允许 |
| `POLICY_ACCESS_DENIED` | failure | 服务访问被策略拒绝 |
| `POLICY_USER_DENIED` | failure | 用户级禁止命中 |
| `POLICY_ROLE_DENIED` | failure | 角色未授权 |

## 9. 网关访问事件

| 事件类型 | 结果 | 触发时机 |
|---|---|---|
| `GATEWAY_SERVICE_ACCESS` | success | 网关成功代理服务请求 |
| `GATEWAY_SERVICE_DENIED` | failure | 网关拒绝服务请求 |
| `GATEWAY_UPSTREAM_UNREACHABLE` | failure | 上游不可达 |
| `GATEWAY_UPSTREAM_TIMEOUT` | failure | 上游超时 |

## 10. 系统设置事件

| 事件类型 | 结果 | 触发时机 |
|---|---|---|
| `SETTING_UPDATED` | success | 修改系统设置 |
| `AUDIT_EXPORTED` | success | 导出审计日志 |
| `AUDIT_EXPORT_REJECTED` | failure | 审计导出被拒绝 |

## 11. 详情字段规范

`details` 必须是结构化对象。

示例：

```json
{
  "before": {
    "status": "enabled"
  },
  "after": {
    "status": "disabled"
  }
}
```

要求：

- 不记录密码
- 不记录令牌
- 不记录设备私钥
- 可记录变更前后摘要

## 12. 审计展示建议

后台审计列表至少显示：

- 时间
- 事件类型
- 操作者
- 目标
- 服务
- 结果
- requestId

审计详情显示：

- 完整结构化字段
- 错误码
- 来源 IP
- User-Agent
- 设备 ID

## 13. 保留策略

第一阶段建议：

- 默认保留 180 天
- 可通过系统设置调整
- 不允许普通管理员直接删除单条审计
