# 日志与监控接入说明

第一阶段以结构化日志、审计事件和健康检查作为运维基础。

## 应用日志字段

Gateway 访问日志至少包含：

| 字段 | 说明 |
|---|---|
| `request_id` | 一次请求的唯一 ID |
| `method` | HTTP 方法 |
| `path` | 请求路径 |
| `status` | 响应状态码 |
| `duration_ms` | 请求耗时 |

日志不得输出：

- 密码
- 访问令牌
- 刷新令牌
- Cookie
- 私钥

## 审计事件

审计数据写入 PostgreSQL `audit_events` 表，后台可按类型、结果、用户、服务查询。

关键事件：

- `auth.login.succeeded`
- `auth.login.failed`
- `service.access.succeeded`
- `service.access.denied`
- `admin.user.created`
- `admin.policy.updated`

## 健康检查

Gateway：

```text
GET /healthz
GET /readyz
```

Admin Web：

```text
GET /health
```

## 告警建议

| 指标 | 建议阈值 |
|---|---|
| Gateway `5xx` 比例 | 5 分钟内超过 2% |
| Gateway `504` 次数 | 5 分钟内连续出现 |
| 登录失败次数 | 单用户 5 分钟内超过 5 次 |
| `/readyz` 失败 | 连续 3 次 |
| PostgreSQL 连接失败 | 任意一次生产失败都告警 |

## 日志保留

- 应用日志建议保留 14 天。
- 审计日志建议保留 180 天或按公司合规要求。
- 失败 E2E 工件在 CI 中保留 7 天。
- 桌面安装包 artifact 在 CI 中保留 14 天。
