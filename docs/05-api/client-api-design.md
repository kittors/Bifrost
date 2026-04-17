# 客户端 API 列表设计

## 1. 目标

本文档定义桌面客户端第一阶段需要调用的 API。

统一前缀：

```text
/api/v1/client
```

所有响应遵守统一响应结构。

## 2. 客户端认证 API

## 2.1 登录

```text
POST /api/v1/client/auth/login
```

请求体：

```json
{
  "username": "alice",
  "password": "********",
  "deviceId": "dev_01",
  "clientVersion": "1.0.0"
}
```

返回：

- 访问令牌
- 刷新令牌
- 当前用户
- 是否需要设备挑战

## 2.2 获取设备挑战

```text
POST /api/v1/client/devices/challenge
```

请求体：

```json
{
  "deviceId": "dev_01"
}
```

返回：

```json
{
  "challengeId": "ch_01",
  "challenge": "base64url-encoded-challenge",
  "expiresIn": 120
}
```

## 2.3 提交设备挑战签名

```text
POST /api/v1/client/devices/challenge/verify
```

请求体：

```json
{
  "challengeId": "ch_01",
  "signature": "base64url-encoded-signature"
}
```

## 2.4 刷新会话

```text
POST /api/v1/client/auth/refresh
```

请求体：

```json
{
  "refreshToken": "refresh_token",
  "deviceId": "dev_01"
}
```

## 2.5 退出登录

```text
POST /api/v1/client/auth/logout
```

效果：

- 吊销当前会话
- 客户端清理本地会话材料

## 3. 设备 API

## 3.1 注册设备

```text
POST /api/v1/client/devices/register
```

请求体：

```json
{
  "name": "Alice MacBook Pro",
  "os": "macOS",
  "clientVersion": "1.0.0",
  "publicKey": "base64url-encoded-public-key",
  "publicKeyFingerprint": "fingerprint"
}
```

返回：

```json
{
  "deviceId": "dev_01",
  "status": "trusted"
}
```

## 3.2 获取当前设备状态

```text
GET /api/v1/client/devices/current
```

返回：

- 设备 ID
- 设备名称
- 设备状态
- 最近活跃时间

## 4. 当前用户 API

## 4.1 获取当前用户

```text
GET /api/v1/client/me
```

返回：

- 用户基础信息
- 角色摘要
- 设备状态

## 5. 服务访问 API

## 5.1 获取可访问服务列表

```text
GET /api/v1/client/services
```

查询参数：

| 参数 | 类型 | 说明 |
|---|---|---|
| `keyword` | string | 搜索关键词 |
| `group` | string | 服务分组 |

返回：

```json
{
  "items": [
    {
      "id": "svc_gitlab",
      "key": "gitlab",
      "name": "GitLab",
      "description": "研发代码仓库",
      "group": "development",
      "status": "enabled",
      "accessSource": "role",
      "lastAccessedAt": "2026-04-17T10:30:00Z"
    }
  ]
}
```

## 5.2 获取服务详情

```text
GET /api/v1/client/services/{serviceId}
```

返回：

- 服务说明
- 当前访问状态
- 权限来源
- 最近访问摘要

## 5.3 创建服务访问入口

```text
POST /api/v1/client/services/{serviceId}/access-url
```

用途：

- 客户端点击服务时，获取受控访问 URL

返回：

```json
{
  "url": "https://gateway.company.com/s/gitlab/",
  "expiresIn": 300
}
```

服务端在返回前必须重新校验：

- 用户状态
- 设备状态
- 会话状态
- 服务状态
- 授权策略

## 6. 诊断 API

## 6.1 获取连接状态

```text
GET /api/v1/client/diagnostics/status
```

返回：

- 服务端时间
- 当前会话状态
- 当前设备状态
- 当前客户端版本兼容性

## 7. 错误处理

客户端需要重点处理：

- `AUTH_INVALID_TOKEN`
- `AUTH_SESSION_EXPIRED`
- `DEVICE_DISABLED`
- `DEVICE_NOT_TRUSTED`
- `POLICY_ACCESS_DENIED`
- `SERVICE_DISABLED`
- `SERVICE_UPSTREAM_UNREACHABLE`

## 8. 审计要求

客户端 API 触发的以下行为需要审计：

- 登录成功
- 登录失败
- 设备注册
- 设备挑战失败
- 服务访问入口创建
- 服务访问拒绝
- 退出登录
