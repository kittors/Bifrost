# 服务端运行参数说明

本文说明 Gateway Server 在第一阶段需要的运行参数、默认值和生产建议。

## 必填参数

| 参数 | 示例 | 说明 |
|---|---|---|
| `BIFROST_PUBLIC_BASE_URL` | `https://gateway.company.com` | 客户端与浏览器访问网关的公网基础地址 |
| `BIFROST_ADMIN_BASE_URL` | `https://admin.gateway.company.com` | 管理后台公网基础地址 |
| `BIFROST_DATABASE_URL` | `postgres://bifrost:***@postgres:5432/bifrost?sslmode=require` | PostgreSQL 连接串 |
| `BIFROST_TOKEN_SECRET` | 32 个字符以上随机字符串 | 访问令牌和访问票据签名密钥 |

## 可选参数

| 参数 | 默认值 | 说明 |
|---|---:|---|
| `BIFROST_ENV` | `development` | 运行环境；设置为 `production` 时会强制校验签名密钥 |
| `PORT` | `8080` | Gateway HTTP 监听端口 |
| `BIFROST_ACCESS_TOKEN_TTL` | `15m` | 访问令牌有效期 |
| `BIFROST_REFRESH_TOKEN_TTL` | `720h` | 刷新令牌有效期 |
| `BIFROST_AUDIT_RETENTION_DAYS` | `180` | 审计日志保留天数建议值 |
| `BIFROST_UPSTREAM_GITLAB` | `http://mock-gitlab:8080` | GitLab 默认上游地址 |
| `BIFROST_UPSTREAM_JENKINS` | `http://mock-jenkins:8080` | Jenkins 默认上游地址 |
| `BIFROST_UPSTREAM_DOCS` | `http://mock-docs:8080` | Docs 默认上游地址 |
| `BIFROST_UPSTREAM_INTERNAL_ADMIN` | `http://mock-internal-admin:8080` | 内部后台默认上游地址 |

## 镜像构建

```bash
pnpm build:gateway:image
```

默认镜像名：

```text
bifrost/gateway:dev
```

## 启动示例

```bash
docker run --rm \
  -p 8080:8080 \
  -e BIFROST_PUBLIC_BASE_URL=https://gateway.company.com \
  -e BIFROST_ADMIN_BASE_URL=https://admin.gateway.company.com \
  -e BIFROST_ENV=production \
  -e BIFROST_DATABASE_URL='postgres://bifrost:secret@postgres:5432/bifrost?sslmode=require' \
  -e BIFROST_TOKEN_SECRET='use-a-random-secret-with-at-least-32-characters' \
  bifrost/gateway:dev
```

## 生产建议

- 生产环境必须设置 `BIFROST_ENV=production`；此时 Gateway 会拒绝默认开发签名密钥和少于 32 个字符的 `BIFROST_TOKEN_SECRET`。
- `BIFROST_TOKEN_SECRET` 必须来自密钥管理系统，不得写入镜像或 Git。
- 数据库连接在生产环境必须开启 TLS 或使用可信内网专线。
- Gateway 前面应放置企业已有 TLS 入口，例如 Nginx、Caddy、Traefik 或云负载均衡。
- Gateway 容器只暴露给公网入口层和内网服务，不直接暴露数据库。
- 上游服务地址必须由后台服务目录和环境配置控制，客户端不能提交任意 upstream URL。
