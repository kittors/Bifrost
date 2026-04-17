# 统一 API 响应结构规范

## 1. 目标

客户端、后台、服务端必须共享一套 API 响应结构，避免：

- 不同接口返回完全不同的外形
- 前端不得不为每个接口单独适配
- 错误处理风格不统一
- 分页结构、错误码、时间字段散乱

本规范适用于：

- 管理后台 API
- 客户端登录与服务列表 API
- 网关辅助 API

## 2. 基本原则

- 所有 JSON 响应都应使用统一包裹结构
- 所有响应都必须携带 `requestId`
- 所有响应都必须携带服务端生成的 `timestamp`
- 成功时 `error` 必须为 `null`
- 失败时 `data` 必须为 `null`

## 3. 标准响应结构

## 3.1 成功响应

```json
{
  "success": true,
  "data": {},
  "meta": {
    "requestId": "req_01JZABCDEF1234567890",
    "timestamp": "2026-04-17T10:30:00Z"
  },
  "error": null
}
```

## 3.2 失败响应

```json
{
  "success": false,
  "data": null,
  "meta": {
    "requestId": "req_01JZABCDEF1234567890",
    "timestamp": "2026-04-17T10:30:00Z"
  },
  "error": {
    "code": "AUTH_INVALID_TOKEN",
    "message": "token is invalid or expired",
    "userMessage": "登录状态已失效，请重新登录",
    "details": {}
  }
}
```

## 3.3 分页响应

```json
{
  "success": true,
  "data": {
    "items": []
  },
  "meta": {
    "requestId": "req_01JZABCDEF1234567890",
    "timestamp": "2026-04-17T10:30:00Z",
    "pagination": {
      "page": 1,
      "pageSize": 20,
      "total": 135,
      "totalPages": 7
    }
  },
  "error": null
}
```

## 4. 字段说明

### 4.1 success

- 类型：`boolean`
- 表示本次业务请求是否成功

### 4.2 data

- 成功时承载业务数据
- 失败时固定为 `null`
- 列表结果统一使用 `items`

### 4.3 meta

必须包含：

- `requestId`
- `timestamp`

可选包含：

- `pagination`

### 4.4 error

失败时必须返回对象，包含：

- `code`
- `message`
- `userMessage`
- `details`

成功时固定为 `null`。

## 5. requestId 规范

要求：

- 服务端为每个请求生成唯一 `requestId`
- `requestId` 写入日志与审计记录
- 客户端和后台在错误提示中可显示 `requestId`

用途：

- 排障
- 审计串联
- 问题定位

## 6. 错误码规范

错误码统一大写蛇形命名，按领域前缀划分：

- `AUTH_*`
- `DEVICE_*`
- `SERVICE_*`
- `POLICY_*`
- `USER_*`
- `ROLE_*`
- `AUDIT_*`
- `INTERNAL_*`

示例：

- `AUTH_INVALID_CREDENTIALS`
- `AUTH_INVALID_TOKEN`
- `DEVICE_NOT_TRUSTED`
- `SERVICE_DISABLED`
- `POLICY_ACCESS_DENIED`
- `USER_DISABLED`

## 7. HTTP 状态码建议

业务错误不应全部塞进 `200`。建议：

- `200 OK`：读取成功、操作成功
- `201 Created`：创建成功
- `204 No Content`：删除成功且无返回体时可选
- `400 Bad Request`：参数错误
- `401 Unauthorized`：未登录、令牌无效
- `403 Forbidden`：已登录但无权限
- `404 Not Found`：对象不存在
- `409 Conflict`：状态冲突
- `422 Unprocessable Entity`：业务校验失败
- `429 Too Many Requests`：频率限制
- `500 Internal Server Error`：服务端内部错误

前端应以 `error.code` 做业务分支，以 HTTP 状态码做大类判断。

## 8. 列表与分页规范

统一要求：

- 列表项放在 `data.items`
- 分页信息放在 `meta.pagination`
- 不允许某接口返回 `rows`，另一个返回 `list`
- 不允许返回裸数组

## 9. 空结果规范

查询成功但没有结果时：

```json
{
  "success": true,
  "data": {
    "items": []
  },
  "meta": {
    "requestId": "req_01JZABCDEF1234567890",
    "timestamp": "2026-04-17T10:30:00Z",
    "pagination": {
      "page": 1,
      "pageSize": 20,
      "total": 0,
      "totalPages": 0
    }
  },
  "error": null
}
```

## 10. 批量操作响应规范

批量操作建议在 `data` 中显式返回结果摘要：

```json
{
  "success": true,
  "data": {
    "requested": 10,
    "succeeded": 8,
    "failed": 2,
    "items": [
      {
        "id": "device_01",
        "success": true
      },
      {
        "id": "device_02",
        "success": false,
        "errorCode": "DEVICE_NOT_FOUND"
      }
    ]
  },
  "meta": {
    "requestId": "req_01JZABCDEF1234567890",
    "timestamp": "2026-04-17T10:30:00Z"
  },
  "error": null
}
```

## 11. 审计事件响应建议

审计接口中的事件记录建议统一包含：

- `id`
- `type`
- `actor`
- `target`
- `result`
- `requestId`
- `occurredAt`
- `summary`

## 12. 前后端协作要求

- 本规范必须落入共享契约包
- 前端禁止手写与契约不一致的响应类型
- 服务端新增接口必须遵守此结构
- 任何例外都必须有文档说明并经过评审
