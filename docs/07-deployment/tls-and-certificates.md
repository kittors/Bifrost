# TLS 与证书配置说明

Bifrost 面向公网入口，生产环境必须全程使用 HTTPS。

## 证书边界

推荐证书终止在统一公网入口层：

```text
Client Browser / Desktop
  |
  v
TLS Entry / Load Balancer
  |
  v
Gateway / Admin Web
```

Gateway 和 Admin Web 可以在内网使用 HTTP，但入口层到应用层之间必须位于可信私有网络。

## 推荐域名

| 组件 | 示例 |
|---|---|
| Gateway | `gateway.company.com` |
| Admin Web | `admin.gateway.company.com` |

## 必要 Header

入口层必须传递：

```text
X-Forwarded-Proto: https
X-Forwarded-For: <client-ip>
Host: <original-host>
```

Gateway 会根据 `X-Forwarded-Proto` 判断服务访问 Cookie 是否需要设置 `Secure`。

## 证书轮换流程

1. 新证书签发完成。
2. 在预发入口层加载新证书。
3. 验证 Admin 登录、客户端登录、服务访问 URL 和 Cookie Secure 属性。
4. 在生产入口层更新证书。
5. 观察 30 分钟错误率和 TLS 握手失败率。
6. 记录证书过期时间和下一次轮换提醒。

## 禁止事项

- 禁止明文公网登录。
- 禁止在客户端关闭 TLS 校验。
- 禁止把私钥提交到 Git。
- 禁止把测试证书直接用于生产。
