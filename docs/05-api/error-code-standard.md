# 错误码枚举规范

## 1. 目标

本文件定义 Bifrost API 的错误码命名、分类、HTTP 状态码映射和前端处理规则。

错误码的目标不是替代 HTTP 状态码，而是为前端、客户端、后台、日志和审计提供稳定的业务分支依据。

## 2. 基本原则

- 错误码必须稳定，不应随文案调整而变化
- 错误码统一使用大写蛇形命名
- 错误码按业务域分组
- 前端展示用户可读文案时使用 `error.userMessage`
- 开发与排障时使用 `error.code`、`error.message`、`meta.requestId`
- 日志与审计中必须记录错误码

## 3. 命名格式

格式：

```text
DOMAIN_REASON
```

示例：

```text
AUTH_INVALID_TOKEN
DEVICE_NOT_TRUSTED
SERVICE_DISABLED
POLICY_ACCESS_DENIED
```

规则：

- `DOMAIN` 表示错误所属业务域
- `REASON` 表示明确原因
- 不使用数字错误码作为主要业务错误码
- 不使用含糊命名，例如 `FAILED`、`ERROR`、`BAD_REQUEST`

## 4. 错误对象结构

所有失败响应中的 `error` 对象必须符合以下结构：

```json
{
  "code": "AUTH_INVALID_TOKEN",
  "message": "token is invalid or expired",
  "userMessage": "登录状态已失效，请重新登录",
  "details": {}
}
```

字段说明：

- `code`：稳定业务错误码
- `message`：开发者可读英文信息，用于日志和调试
- `userMessage`：用户可读中文信息，用于界面展示
- `details`：可选结构化信息，不应包含敏感数据

## 5. 通用错误码

| 错误码 | HTTP 状态码 | 用户文案建议 | 说明 |
|---|---:|---|---|
| `COMMON_BAD_REQUEST` | `400` | 请求参数不正确 | 请求格式错误或基础参数缺失 |
| `COMMON_VALIDATION_FAILED` | `422` | 提交内容未通过校验 | 字段级业务校验失败 |
| `COMMON_NOT_FOUND` | `404` | 资源不存在 | 目标对象不存在 |
| `COMMON_CONFLICT` | `409` | 当前状态不允许执行该操作 | 状态冲突 |
| `COMMON_RATE_LIMITED` | `429` | 操作过于频繁，请稍后再试 | 触发限流 |
| `COMMON_INTERNAL_ERROR` | `500` | 服务暂时不可用，请稍后再试 | 未分类服务端错误 |

## 6. 认证错误码 AUTH

| 错误码 | HTTP 状态码 | 用户文案建议 | 说明 |
|---|---:|---|---|
| `AUTH_INVALID_CREDENTIALS` | `401` | 账号或密码不正确 | 登录凭证错误 |
| `AUTH_INVALID_TOKEN` | `401` | 登录状态已失效，请重新登录 | 访问令牌无效或过期 |
| `AUTH_REFRESH_TOKEN_INVALID` | `401` | 登录状态已失效，请重新登录 | 刷新令牌无效 |
| `AUTH_SESSION_EXPIRED` | `401` | 登录已过期，请重新登录 | 会话过期 |
| `AUTH_SESSION_REVOKED` | `401` | 当前会话已被管理员终止 | 会话被吊销 |
| `AUTH_PASSWORD_REQUIRED` | `422` | 请输入密码 | 密码字段缺失 |
| `AUTH_PASSWORD_TOO_WEAK` | `422` | 密码强度不足 | 密码策略不满足 |

## 7. 用户错误码 USER

| 错误码 | HTTP 状态码 | 用户文案建议 | 说明 |
|---|---:|---|---|
| `USER_DISABLED` | `403` | 账号已被禁用 | 用户状态不可用 |
| `USER_NOT_FOUND` | `404` | 用户不存在 | 用户 ID 不存在 |
| `USER_ALREADY_EXISTS` | `409` | 用户已存在 | 用户名或邮箱重复 |
| `USER_CANNOT_DISABLE_SELF` | `422` | 不能禁用当前登录账号 | 管理员保护 |
| `USER_CANNOT_DELETE_SELF` | `422` | 不能删除当前登录账号 | 管理员保护 |
| `USER_LAST_ADMIN_REQUIRED` | `422` | 至少需要保留一个管理员账号 | 防止系统失去管理入口 |

## 8. 角色错误码 ROLE

| 错误码 | HTTP 状态码 | 用户文案建议 | 说明 |
|---|---:|---|---|
| `ROLE_NOT_FOUND` | `404` | 角色不存在 | 角色 ID 不存在 |
| `ROLE_ALREADY_EXISTS` | `409` | 角色已存在 | 角色名称重复 |
| `ROLE_IN_USE` | `409` | 角色正在被使用，无法删除 | 角色仍有关联用户 |
| `ROLE_SYSTEM_ROLE_LOCKED` | `403` | 系统角色不允许修改 | 系统内置角色保护 |
| `ROLE_POLICY_INVALID` | `422` | 角色权限配置无效 | 权限配置不合法 |

## 9. 设备错误码 DEVICE

| 错误码 | HTTP 状态码 | 用户文案建议 | 说明 |
|---|---:|---|---|
| `DEVICE_NOT_FOUND` | `404` | 设备不存在 | 设备 ID 不存在 |
| `DEVICE_NOT_TRUSTED` | `403` | 当前设备未被信任 | 设备未注册或未绑定 |
| `DEVICE_DISABLED` | `403` | 当前设备已被禁用 | 设备被管理员禁用 |
| `DEVICE_KEY_INVALID` | `401` | 设备身份校验失败 | 设备签名不合法 |
| `DEVICE_CHALLENGE_EXPIRED` | `401` | 设备验证已过期，请重试 | 设备挑战过期 |
| `DEVICE_ALREADY_BOUND` | `409` | 设备已绑定 | 重复绑定设备 |
| `DEVICE_LIMIT_EXCEEDED` | `422` | 已达到设备绑定数量上限 | 单用户设备数量限制 |

## 10. 服务目录错误码 SERVICE

| 错误码 | HTTP 状态码 | 用户文案建议 | 说明 |
|---|---:|---|---|
| `SERVICE_NOT_FOUND` | `404` | 服务不存在 | 服务 ID 不存在 |
| `SERVICE_DISABLED` | `403` | 服务已停用 | 服务处于禁用状态 |
| `SERVICE_ALREADY_EXISTS` | `409` | 服务已存在 | 服务标识重复 |
| `SERVICE_UPSTREAM_INVALID` | `422` | 服务目标地址无效 | 上游配置不合法 |
| `SERVICE_UPSTREAM_UNREACHABLE` | `502` | 目标服务暂时不可访问 | 网关无法连接上游 |
| `SERVICE_ROUTE_INVALID` | `422` | 服务路由配置无效 | 路由前缀或路径规则错误 |

## 11. 策略错误码 POLICY

| 错误码 | HTTP 状态码 | 用户文案建议 | 说明 |
|---|---:|---|---|
| `POLICY_ACCESS_DENIED` | `403` | 你没有访问该服务的权限 | 常规策略拒绝 |
| `POLICY_USER_DENIED` | `403` | 你的账号已被禁止访问该服务 | 用户级禁止覆盖 |
| `POLICY_ROLE_DENIED` | `403` | 当前角色无权访问该服务 | 角色未授权 |
| `POLICY_DEVICE_DENIED` | `403` | 当前设备无权访问该服务 | 设备维度拒绝 |
| `POLICY_RULE_INVALID` | `422` | 访问策略配置无效 | 策略规则不合法 |

## 12. 审计错误码 AUDIT

| 错误码 | HTTP 状态码 | 用户文案建议 | 说明 |
|---|---:|---|---|
| `AUDIT_EVENT_NOT_FOUND` | `404` | 审计事件不存在 | 审计记录不存在 |
| `AUDIT_QUERY_INVALID` | `422` | 审计查询条件无效 | 时间范围或过滤条件无效 |
| `AUDIT_EXPORT_TOO_LARGE` | `422` | 导出范围过大，请缩小筛选条件 | 导出数量超过限制 |

## 13. 网关错误码 GATEWAY

| 错误码 | HTTP 状态码 | 用户文案建议 | 说明 |
|---|---:|---|---|
| `GATEWAY_BAD_UPSTREAM` | `502` | 目标服务响应异常 | 上游返回异常响应 |
| `GATEWAY_UPSTREAM_TIMEOUT` | `504` | 目标服务响应超时 | 上游超时 |
| `GATEWAY_ROUTE_NOT_FOUND` | `404` | 访问入口不存在 | 服务路由未命中 |
| `GATEWAY_REQUEST_TOO_LARGE` | `413` | 请求内容过大 | 请求体超过限制 |

## 14. 状态码使用规则

HTTP 状态码表达协议和大类结果，错误码表达业务原因。

| HTTP 状态码 | 使用场景 |
|---:|---|
| `200` | 读取成功、普通操作成功 |
| `201` | 创建成功 |
| `204` | 删除成功且没有响应体 |
| `400` | 请求格式错误或基础参数缺失 |
| `401` | 未登录、令牌无效、会话过期、设备签名认证失败 |
| `403` | 已认证但无权限、账号禁用、设备禁用、策略拒绝 |
| `404` | 资源不存在、路由不存在 |
| `409` | 重复创建、状态冲突、资源正在使用 |
| `413` | 请求体过大 |
| `422` | 业务校验失败 |
| `429` | 请求频率过高 |
| `500` | 服务端内部错误 |
| `502` | 上游服务异常 |
| `504` | 上游服务超时 |

## 15. 前端处理策略

前端处理顺序建议：

1. 根据 HTTP 状态码判断大类
2. 根据 `error.code` 进入具体业务分支
3. 使用 `error.userMessage` 展示给用户
4. 在错误详情或排障入口展示 `meta.requestId`

特殊处理：

- `401`：清理会话并引导重新登录
- `403`：展示权限不足或设备不可用状态
- `429`：展示冷却提示
- `502/504`：提示目标服务不可用，而不是提示用户无权限

## 16. 审计记录要求

以下错误必须进入审计：

- 登录失败
- 设备签名失败
- 设备被禁用后仍尝试访问
- 策略拒绝
- 用户级禁止命中
- 服务禁用后仍尝试访问
- 管理员危险操作失败

审计中至少记录：

- `requestId`
- `error.code`
- 用户 ID
- 设备 ID
- 服务 ID
- 发生时间
- 来源 IP

## 17. 变更规则

- 新增错误码必须写入本文档
- 删除或重命名错误码必须经过兼容性评估
- 前端已经依赖的错误码不得随意改名
- 错误码文案可以优化，但 `code` 不应随意变化
