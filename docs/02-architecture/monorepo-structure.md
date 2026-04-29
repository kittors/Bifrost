# Monorepo 与项目结构设计

## 1. 设计目标

Monorepo 结构必须服务于以下目标：

- 客户端、后台、服务端文档与共享包统一管理
- 设计 token、第三方 UI 依赖版本、API 契约只有一份事实源
- 能清晰区分业务应用与共享能力
- 不让前端和后端各自形成孤立工程体系

## 2. 顶层目录设计

推荐目录如下：

```text
Bifrost/
  apps/
    desktop/
    admin/
    gateway/

  packages/
    design-tokens/
    contracts/
    config-typescript/
    shared-types/
    shared-utils/

  docs/
  scripts/
  .changeset/
  package.json
  pnpm-workspace.yaml
  turbo.json
  .oxlintrc.json
  .oxfmtrc.json
  README.md
```

## 3. apps 目录职责

## 3.1 apps/desktop

桌面客户端应用，基于 `Electron + React`。

职责包括：

- Electron 主进程
- Preload 桥接
- 渲染层页面与业务逻辑
- 客户端本地会话与设备状态管理

推荐结构：

```text
apps/desktop/
  electron/
    main/
    preload/
  renderer/
    src/
      app/
      pages/
      widgets/
      features/
      entities/
      shared/
  resources/
  build/
  package.json
```

## 3.2 apps/admin

Web 管理后台，基于 `React + Vite`。

职责包括：

- 管理后台页面
- 用户、角色、设备、服务、审计配置界面
- 管理后台专用业务逻辑

推荐结构：

```text
apps/admin/
  src/
    app/
    pages/
    widgets/
    features/
    entities/
    shared/
  public/
  index.html
  package.json
```

## 3.3 apps/gateway

Go 服务端工程。

职责包括：

- 对外 API
- 鉴权
- 设备信任
- 服务访问策略
- 反向代理
- 审计

推荐结构：

```text
apps/gateway/
  cmd/
    bifrost-gateway/
  internal/
  migrations/
  api/
  configs/
  go.mod
```

## 4. packages 目录职责

## 4.1 packages/design-tokens

统一设计 token 与全局样式入口。

职责：

- `theme.css`：主题 token 定义
- `base.css`：全局基础行为
- `app.css`：应用级入口样式

这是唯一允许承载全局样式语义的包。

## 4.2 packages/contracts

统一契约包。

职责：

- OpenAPI 文档
- TS 生成类型
- 错误码常量
- 分页结构
- 审计事件名

客户端与后台都从这里消费类型与协议定义。

## 4.3 packages/config-typescript

统一 TypeScript 配置。

用途：

- `base.json`
- `react.json`
- `electron.json`

避免各应用自行散落 tsconfig 规则。

## 4.4 packages/shared-types / shared-utils

仅放跨应用通用的类型与纯函数工具。

禁止放业务耦合逻辑。

## 5. 前端源码分层规范

客户端与后台渲染层统一采用以下分层：

```text
src/
  app/
  pages/
  widgets/
  features/
  entities/
  shared/
```

职责说明：

- `app`：应用初始化、Provider、路由、主题、全局布局
- `pages`：页面组装层，不承载复杂业务逻辑
- `widgets`：页面复合区块
- `features`：具体业务动作
- `entities`：业务实体模型与展示
- `shared`：无业务语义的通用能力

## 6. Go 服务端目录规范

服务端建议按业务能力拆目录，而不是机械地按 controller/service/repository 三层堆叠。

推荐目录：

```text
internal/
  auth/
  user/
  role/
  device/
  servicecatalog/
  policy/
  session/
  reverseproxy/
  audit/
  crypto/
  config/
```

设计原则：

- 每个目录只负责一个明确能力边界
- `policy` 不直接做转发
- `reverseproxy` 不直接决定是否允许访问
- `audit` 负责记录，不做业务授权判断

## 7. 命名规范

### 前端

- 包名、目录名使用小写短横线
- 业务 Feature 命名以动作开头，例如 `auth-login`、`device-register`
- 页面命名使用业务域，例如 `services`、`audit`、`devices`

### Go

- 包名使用短小、单数、语义明确的英文小写
- 不使用 `common`、`utils`、`misc` 这类模糊目录作为主要承载点

## 8. 共享原则

以下内容必须沉淀为共享能力：

- 设计 token
- 基础 UI 组件
- 响应结构与错误码
- 业务枚举与公共类型
- 日期、格式化、日志展示工具

以下内容不得过早共享：

- 后台专属业务页面
- 客户端专属窗口逻辑
- 仅服务端使用的领域模型

## 9. 约束清单

- 不允许客户端和后台各自维护一套颜色变量
- 不允许多个位置重复定义错误码
- 不允许 API 响应结构在各端自由变化
- 不允许大而全的 `shared` 文件夹吞掉所有业务边界
- 不允许单个文件持续膨胀成“万能文件”

## 10. 演进策略

第一阶段以最小但清晰的结构上线：

- 先建好应用与共享包边界
- 先沉淀 token、UI、契约
- 再分别扩展客户端、后台、服务端功能

只有当某个能力被至少两个应用使用，才应当上升为共享包。
