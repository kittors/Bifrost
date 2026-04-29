# 第一版内部试用发布说明

## 版本目标

第一版内部试用版本用于验证：

- 管理员可以配置账号、角色、服务和授权策略。
- 客户端用户登录后只能看到已授权服务。
- 网关可以代理多个私有 Web 上游。
- 设备禁用、服务禁用、用户 deny 可以立即生效。
- 审计日志可查询关键访问行为。

## 包含组件

- `bifrost/gateway:dev`
- `bifrost/admin-web:dev`
- macOS Desktop 安装包
- Windows Desktop 安装包
- Linux Desktop 安装包

## 默认测试账号

| 账号 | 角色 | 默认密码 |
|---|---|---|
| `admin` | `admin` | `ChangeMe123!` |
| `alice` | `developer` | `ChangeMe123!` |
| `bob` | `ops` | `ChangeMe123!` |

试用环境上线后必须立即修改默认密码。

## 试用验收路径

1. 管理员登录 Admin Web。
2. 查看 `gitlab`、`jenkins`、`docs` 服务目录。
3. 使用 Desktop 以 `alice` 登录。
4. 确认客户端仅展示 `GitLab` 和 `Docs`。
5. 打开 `GitLab`，确认浏览器访问的是 Gateway `/s/gitlab`。
6. 强行访问 `/s/jenkins`，确认返回 `POLICY_ACCESS_DENIED`。
7. 在后台审计页查询登录和访问记录。

## 发布风险

- 第一阶段不包含自动更新。
- 第一阶段未启用正式代码签名。
- 第一阶段客户端不接管系统代理、DNS 或路由。
- 第一阶段只支持 HTTP/HTTPS Web 服务，不支持全流量 VPN。
