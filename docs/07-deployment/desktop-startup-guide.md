# Desktop Client 启动说明

Desktop Client 是一个小卡片式访问入口，只负责登录、设备信任、展示可访问服务和打开浏览器访问网关。

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

客户端仅通过 Gateway API 登录并获取服务访问 URL，实际访问由系统默认浏览器打开。

## 签名与公证

第一阶段 CI 先关闭自动签名发现：

```text
CSC_IDENTITY_AUTO_DISCOVERY=false
```

进入正式发布前需要补齐：

- macOS Developer ID 签名与 notarization。
- Windows 代码签名证书。
- Linux 包校验哈希与发布渠道说明。
