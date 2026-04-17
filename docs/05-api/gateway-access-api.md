# 网关访问 API 与代理规则

## 1. 目标

本文档定义私有 Web 服务通过网关访问时的入口规则、鉴权规则、代理规则和错误处理。

第一阶段网关只代理 `HTTP/HTTPS` Web 服务。

## 2. 网关入口

推荐访问路径：

```text
https://gateway.company.com/s/{serviceKey}/...
```

示例：

```text
https://gateway.company.com/s/gitlab/
```

规则：

- `{serviceKey}` 必须来自服务目录
- 客户端不能指定任意上游地址
- 网关根据 `serviceKey` 查找服务配置

## 3. 鉴权顺序

每次访问网关入口时必须校验：

1. 会话是否有效
2. 用户是否启用
3. 设备是否可信且启用
4. 服务是否存在
5. 服务是否启用
6. 用户是否命中禁止覆盖
7. 用户是否命中允许覆盖
8. 用户角色是否允许访问服务

任何一步失败都必须拒绝访问。

## 4. Cookie 与令牌策略

浏览器访问网关入口时需要可验证的访问凭证。

第一阶段推荐：

- 客户端通过 API 获取短期访问 URL
- 服务端可设置短期访问 Cookie 或一次性访问票据
- 访问票据与用户、设备、服务绑定

要求：

- 票据短期有效
- 票据不可跨服务复用
- 票据不可由客户端伪造

## 5. 代理规则

网关转发时：

- 根据服务配置确定 `upstreamUrl`
- 保留必要请求路径
- 正确处理 query string
- 设置合理超时
- 限制请求体大小
- 记录访问审计

网关不允许：

- 将客户端传入的 URL 当作上游
- 允许访问服务目录外的地址
- 跳过授权直接转发

## 6. Header 处理

网关应设置或追加：

- `X-Bifrost-Request-Id`
- `X-Bifrost-User-Id`
- `X-Bifrost-Device-Id`
- `X-Bifrost-Service-Id`

注意：

- 是否传递用户身份给上游需要按服务配置决定
- 不应把内部敏感令牌传给上游
- 对原始 `X-Forwarded-*` 头要有明确覆盖或追加策略

## 7. WebSocket 支持

GitLab 等服务可能需要 WebSocket。

第一阶段建议预留支持：

- HTTP Upgrade
- WebSocket 代理

但必须仍然经过同样的访问策略判断。

## 8. 路径改写

若服务配置 `publicPath=/s/gitlab`，上游为 `http://10.0.0.12:8929`，则：

```text
/s/gitlab/users/sign_in
```

转发为：

```text
http://10.0.0.12:8929/users/sign_in
```

路径改写必须是服务端配置驱动，不由客户端控制。

## 9. 超时建议

建议默认值：

- 连接上游超时：`10s`
- 响应头超时：`30s`
- 空闲连接超时：`90s`
- 最大请求体：按服务配置，默认 `50MB`

GitLab 上传、制品库等场景可能需要服务级调整。

## 10. 错误映射

| 场景 | HTTP 状态码 | 错误码 |
|---|---:|---|
| 未登录 | `401` | `AUTH_INVALID_TOKEN` |
| 会话过期 | `401` | `AUTH_SESSION_EXPIRED` |
| 设备禁用 | `403` | `DEVICE_DISABLED` |
| 策略拒绝 | `403` | `POLICY_ACCESS_DENIED` |
| 服务不存在 | `404` | `SERVICE_NOT_FOUND` |
| 服务禁用 | `403` | `SERVICE_DISABLED` |
| 上游不可达 | `502` | `SERVICE_UPSTREAM_UNREACHABLE` |
| 上游超时 | `504` | `GATEWAY_UPSTREAM_TIMEOUT` |

## 11. 审计要求

网关至少记录：

- 服务访问成功
- 服务访问拒绝
- 上游不可达
- 上游超时

审计字段：

- requestId
- userId
- deviceId
- serviceId
- path
- method
- result
- errorCode
- sourceIp
- userAgent

## 12. 第一阶段验收标准

- 访问未配置服务时拒绝
- 访问未授权服务时拒绝
- 服务禁用后立即拒绝
- 上游不可达时返回明确错误
- 成功访问写入审计
