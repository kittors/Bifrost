# AGENTS.md

- 所有通过 `spawn_agent` 调用的子代理必须显式指定 `model: "gpt-5.3-codex"`；不要使用 `gpt-5.1-codex-mini`。轻量子任务可通过降低 `reasoning_effort`（例如 `low`）控制成本。
- 前端样式优先使用 HeroUI 组件、Tailwind CSS v4 utility class 和 `@bifrost/design-tokens` 主题变量。只有全局基础样式、第三方集成或 Tailwind 难以表达的场景才添加少量原生 CSS。
- 客户端与管理后台的基础交互组件统一使用固定版本 `@heroui/react@3.0.3`。不要重新引入 `packages/ui` 或 `@bifrost/ui`。
- 图标按钮和状态入口优先使用 `lucide-react` 图标；按钮内有图标时保持紧凑尺寸，避免页面层手写 SVG。
- 允许在 app 内保留很薄的 HeroUI 适配层来承接 Modal、Drawer、错误态等局部组合，但底层必须消费 HeroUI 组件，不要创建新的跨包 UI 组件库。

## 语言与注释规范

- 所有面向用户的回答默认使用中文；除非用户明确要求其他语言，或必须保留命令、错误原文、API 名称、协议字段等不可翻译内容。
- 新增或修改注释时默认使用中文，包括代码注释、配置文件注释、脚本输出说明和文档里的说明性文字。
- 环境变量名、包名、命令名、URL、HTTP 字段、错误码、第三方产品名等技术标识保持原文，不要为了中文化而改坏可执行内容。

## 分支与发布流程

- 开发新的功能必须新建分支，不允许直接在 `dev` 或 `main` 上开发功能代码。
- 功能分支完成后必须跑完对应测试；测试没有任何问题后合并到 dev 分支中。
- 需要线上测试时，先合并到 `dev` 分支并推送远端，等待 GitHub Action 执行完成。
- GitHub Action 执行完成后，先确认部署环境接口已通，再开始线上功能测试。
- `main` 只接收已经在 `dev` 验证过的内容；不要跳过 `dev` 直接合并到 `main`。

## 应用职责

### `apps/desktop`

- Bifrost 桌面客户端，技术栈为 `Electron + React + Vite`。
- 负责 Electron main/preload、渲染层客户端界面、本地安全存储、设备身份、会话刷新、本地代理启动与打开授权服务。
- UI 定位是“小卡片式私有服务访问入口”，不是管理后台；不使用常驻 Sidebar，不做大工作台布局。
- 渲染层继续按 `app / features / entities / shared` 分层：`app` 负责 Provider 和整体壳，`features` 负责登录、服务、账号、设置、诊断等用户动作，`entities` 负责设备/会话/服务 API 与模型，`shared` 只放无业务耦合的工具。

### `apps/admin`

- Bifrost Web 管理后台，技术栈为 `React + Vite + TanStack Router + React Query`。
- 负责用户、角色、设备、服务目录、审计等管理页面。
- UI 定位是安全控制台：高密度、可扫描、可追溯；优先使用表格、筛选条、详情 Drawer 和确认 Dialog。
- 错误态必须通过共享 `ErrorState` 呈现用户可读说明和 requestId。

### `apps/gateway`

- Go 服务端工程，负责 API、鉴权、设备信任、服务访问策略、反向代理和审计。
- 服务端按业务能力拆分目录，避免把授权、代理和审计逻辑混在同一层。

## 包职责

### `packages/design-tokens`

- Bifrost 设计系统的样式事实源，负责 Tailwind v4 主题 token、light/dark 主题变量和全局基础样式入口。
- `src/theme.css` 定义 Tailwind 可消费的语义 token；`src/base.css` 定义全局基础行为；`src/app.css` 是应用统一样式入口。
- 新增颜色、字号、圆角、控件高度、表格密度、桌面窗口尺寸或后台布局尺寸时，优先沉淀到这里，再由 app 或 HeroUI 适配层消费。

### `packages/contracts`

- API 契约事实源，负责 OpenAPI、生成类型、错误码、审计事件名和统一响应结构。
- 客户端、管理后台和服务端测试都应优先消费这里的类型与常量。

### `packages/shared-types`

- 存放跨应用通用类型；不得放带具体业务流程的实现逻辑。

### `packages/shared-utils`

- 存放跨应用可复用的纯函数工具；不得依赖浏览器、Electron 或服务端运行时。

### `packages/config-typescript`

- 统一 TypeScript 配置，包括 base、React 和 Electron 配置。

### `packages/config-vitest`

- 统一 Vitest 配置和测试 setup，供 React 包和应用复用。

## 前端实现约束

- 客户端和后台都必须使用 `@bifrost/design-tokens/app.css` 作为样式入口。
- 页面层可以使用 Tailwind 布局 utility；当某个视觉值重复出现或代表设计系统规格时，必须优先复用 HeroUI variant 或沉淀到 `packages/design-tokens`。
- 避免在 app 层手写重复的基础按钮、输入框和卡片壳；优先组合 HeroUI。
- 接口错误进入 UI 前必须转换为用户可读文案；有 requestId 时必须展示。
- 桌面端默认保持 `420px` 左右卡片密度，后台默认保持 `232px` Sidebar、`52px` Top Bar 和紧凑表格密度。
