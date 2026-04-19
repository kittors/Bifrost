# Desktop Client 启动说明

Desktop Client 是一个小卡片式访问入口，负责登录、设备信任、展示可访问服务，并在登录后启动 Bifrost 专用本地回环代理。

## 本地开发

```bash
pnpm --filter @bifrost/desktop dev
```

## 构建

```bash
pnpm --filter @bifrost/desktop build
```

## 本机目录包验证

```bash
CSC_IDENTITY_AUTO_DISCOVERY=false pnpm --filter @bifrost/desktop exec electron-builder --config electron-builder.yml --dir
```

产物目录：

```text
apps/desktop/release
```

## 三端安装包脚本

| 平台 | 命令 | 产物 |
|---|---|---|
| macOS | `pnpm --filter @bifrost/desktop dist:mac` | `dmg`、`zip` |
| Windows | `pnpm --filter @bifrost/desktop dist:win` | `nsis` |
| Linux | `pnpm --filter @bifrost/desktop dist:linux` | `AppImage`、`deb`、`tar.gz` |

三端正式产物由 `.github/workflows/desktop-packages.yml` 在对应 runner 上构建并上传 artifact。

## 客户端网络边界

客户端第一阶段明确不做以下行为：

- 不修改系统代理。
- 不修改系统 DNS。
- 不修改系统路由。
- 不安装 VPN 驱动。

客户端仅通过 Gateway API 登录并启动本机回环入口，实际访问由系统默认浏览器或 API 工具访问 `127.0.0.1` 完成。

当前访问模型为：

```text
浏览器 / API 工具
-> http://127.0.0.1:18080/s/{serviceKey}/...
-> Desktop Client Main Process 本地代理
-> Gateway /s/{serviceKey}
-> 私有服务 upstream
```

约束：

- 本地代理只监听 `127.0.0.1`
- 默认端口从 `18080` 起自动避让到 `18099`
- 支持 HTTP 请求与 WebSocket `Upgrade`，用于 GitLab 等需要长连接的 Web 服务
- 客户端退出登录后，本地代理立即停止
- 未启动客户端或未登录时，本地入口不可用
- 不开放任何系统级代理设置

## 签名与公证

第一阶段 CI 先关闭自动签名发现：

```text
CSC_IDENTITY_AUTO_DISCOVERY=false
```

进入正式发布前需要补齐：

- macOS Developer ID 签名与 notarization。
- Windows 代码签名证书。
- Linux 包校验哈希与发布渠道说明。
