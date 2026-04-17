# 部署架构与运行环境

## 1. 目标

本文档描述 Bifrost 第一阶段的部署形态、网络边界、运行环境和证书要求。

## 2. 部署组件

第一阶段包含：

- Gateway Server
- Admin Web
- PostgreSQL
- 私有 Web 服务上游
- Desktop Client

## 3. 推荐部署拓扑

```text
Internet
  |
  v
Public TLS Entry
  |
  v
Gateway Server
  |-----------------> PostgreSQL
  |
  |-----------------> GitLab / Jenkins / Internal Web Services
```

公网只暴露：

- 网关入口
- 后台入口

私有服务不直接暴露给公网。

## 4. 域名建议

建议：

- `gateway.company.com`
- `admin.gateway.company.com`

也可以共用同一域名不同路径，但独立子域名更清晰。

## 5. TLS 要求

生产环境必须启用 HTTPS。

要求：

- 使用可信证书
- 禁止明文登录
- 禁止明文管理后台
- 证书到期前必须有告警或更新流程

## 6. 数据库

推荐：

- PostgreSQL `18.x`

数据库职责：

- 用户
- 角色
- 设备
- 服务目录
- 会话
- 审计
- 系统设置

第一阶段不默认引入 Redis。

## 7. 配置方式

服务端配置建议来源：

- 环境变量
- 配置文件
- 启动参数

配置项包括：

- 监听地址
- 公网基础 URL
- 数据库连接
- TLS 设置
- 会话有效期
- 审计保留策略
- 上游超时默认值

## 8. 网络边界

Gateway Server 必须能访问：

- PostgreSQL
- 内部私有服务

外部用户只能访问：

- Gateway Server 公开入口
- Admin Web 公开入口

外部用户不应直接访问：

- PostgreSQL
- GitLab 内网端口
- Jenkins 内网端口
- Docker 内部服务端口

## 9. 反向代理前置层

可以使用企业已有反向代理或负载均衡作为 TLS 入口，例如：

- Nginx
- Caddy
- Traefik
- 云厂商负载均衡

要求：

- 正确传递来源 IP
- 设置请求体大小限制
- 设置合理超时
- 不绕过 Bifrost 鉴权

## 10. 日志与观测

第一阶段至少需要：

- 应用结构化日志
- 请求日志
- 审计日志
- 错误日志

日志必须包含：

- requestId
- errorCode
- userId，若存在
- deviceId，若存在
- serviceId，若存在

## 11. 备份

必须备份：

- PostgreSQL 数据库
- 服务端配置

审计日志属于重要安全数据，应纳入备份策略。

## 12. 发布策略

建议：

- 服务端先支持数据库迁移
- 后台和网关版本同步发布
- 客户端发布需考虑三端安装包

客户端自动更新第一阶段可先预留，后续实现。

## 13. 健康检查

服务端提供：

```text
GET /healthz
GET /readyz
```

用途：

- `healthz`：进程存活
- `readyz`：数据库和关键依赖可用

## 14. 第一阶段验收标准

- 网关和后台通过 HTTPS 访问
- 私有服务公网不可直连
- 网关可访问上游服务
- 审计日志可写入数据库
- 健康检查可用于部署探针
