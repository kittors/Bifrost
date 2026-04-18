# requestId 排障入口

每个 Gateway API 响应都会包含 requestId，响应头也会返回 `X-Request-Id`。

## 用户反馈时需要收集

- 访问时间
- 用户名
- 服务 key
- 浏览器或客户端截图
- 响应中的 `meta.requestId`
- 响应头 `X-Request-Id`

## 后台查询路径

1. 管理员登录 Admin Web。
2. 打开审计页。
3. 按用户、服务或事件类型过滤。
4. 对照用户提供的时间和 requestId。

## 日志查询

在 Gateway 日志系统中搜索：

```text
request_id=<requestId>
```

重点观察：

- `status`
- `path`
- `duration_ms`
- 错误码
- 是否出现上游 `502` 或 `504`

## 常见判断

| 现象 | 判断方向 |
|---|---|
| `AUTH_INVALID_TOKEN` | 用户登录态失效或访问票据过期 |
| `AUTH_SESSION_REVOKED` | 用户、设备或密码变更导致会话吊销 |
| `POLICY_ACCESS_DENIED` | 角色授权或用户级 deny 阻止访问 |
| `SERVICE_DISABLED` | 服务目录被禁用 |
| `SERVICE_NOT_FOUND` | 服务 key 未配置 |
| `GATEWAY_UPSTREAM_UNREACHABLE` | 上游容器或内网服务不可达 |
| `GATEWAY_UPSTREAM_TIMEOUT` | 上游慢响应或超时 |
