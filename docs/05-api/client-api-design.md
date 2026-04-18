# 客户端 API 列表设计

## 1. 目标

本文档定义桌面客户端第一阶段需要调用的 API。

统一前缀：

```text
/api/v1/client
```

所有响应遵守统一响应结构。

## 2. 客户端认证 API

第一阶段实现说明：

- 首台设备可通过 `POST /api/v1/client/devices/bootstrap` 完成账号校验、设备公钥绑定与会话签发
- 已绑定设备的日常登录继续使用 `POST /api/v1/client/auth/login`
- 设备注册、挑战申请、挑战验签仍要求客户端已携带有效 `Bearer access token`

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

## 2.2 首台设备 bootstrap

```text
POST /api/v1/client/devices/bootstrap
```

用途：

- 用户首次安装客户端、本机还没有 `deviceId` 时使用
- 服务端先校验账号密码，再绑定客户端生成的 Ed25519 公钥
- 绑定成功后直接签发与该设备关联的会话

请求体：

```json
{
  "username": "alice",
  "password": "********",
  "deviceName": "Alice MacBook Pro",
  "deviceOs": "macOS",
  "clientVersion": "0.1.0",
  "publicKey": "base64url-ed25519-public-key",
  "publicKeyFingerprint": "fp_xxxxxxxx"
}
```

返回：

```json
{
  "accessToken": "access_token",
  "refreshToken": "refresh_token",
  "expiresIn": 900,
  "user": {
    "id": "user_alice",
    "username": "alice",
    "displayName": "Alice",
    "roles": ["role_developer"]
  },
  "device": {
    "deviceId": "device_01",
    "status": "trusted"
  }
}
```

安全约束：

- 该接口只用于账号密码正确后的新设备绑定
- 服务端保存公钥与指纹，不接收也不保存私钥
- 若公钥指纹已绑定，返回 `DEVICE_ALREADY_BOUND`
- 客户端必须把私钥保存在系统安全存储中

## 2.3 获取设备挑战

```text
POST /api/v1/client/devices/challenge
```

请求头：

- `Authorization: Bearer <access_token>`

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

## 2.4 提交设备挑战签名

```text
POST /api/v1/client/devices/challenge/verify
```

请求头：

- `Authorization: Bearer <access_token>`

请求体：

```json
{
  "challengeId": "ch_01",
  "signature": "base64url-encoded-signature"
}
```

## 2.5 刷新会话

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

## 2.6 退出登录

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

请求头：

- `Authorization: Bearer <access_token>`

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
