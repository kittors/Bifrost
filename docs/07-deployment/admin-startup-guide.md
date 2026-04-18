# Admin Web 启动说明

Admin Web 是独立的 React 管理后台，用于配置账号、角色、设备、服务目录、策略和审计查询。

## 镜像构建

```bash
pnpm build:admin:image
```

默认镜像名：

```text
bifrost/admin-web:dev
```

## 运行方式

```bash
docker run --rm -p 5173:5173 bifrost/admin-web:dev
```

健康检查：

```bash
curl -fsS http://127.0.0.1:5173/health
```

## 路由回退

Admin Web 由 Nginx 托管静态产物：

- `/health` 返回 `200 ok`
- `/assets/*` 按静态文件精确匹配
- 其他路径回退到 `/index.html`

这样用户刷新 `/users`、`/roles`、`/services` 等 SPA 路由时不会得到 404。

## 生产反向代理示例

```nginx
server {
  listen 443 ssl http2;
  server_name admin.gateway.company.com;

  ssl_certificate /etc/ssl/company/fullchain.pem;
  ssl_certificate_key /etc/ssl/company/privkey.pem;

  location / {
    proxy_pass http://admin-web:5173;
    proxy_set_header Host $host;
    proxy_set_header X-Forwarded-Proto https;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
  }
}
```

## 安全要求

- Admin Web 必须通过 HTTPS 访问。
- Admin Web 的 API 调用必须只访问受控 Gateway API。
- 生产环境不得在公网暴露 PostgreSQL、GitLab、Jenkins 等内部端口。
