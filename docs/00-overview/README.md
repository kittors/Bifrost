# Bifrost 文档总览

## 项目定位

Bifrost 是一个面向企业私有服务访问控制的系统。它的目标不是把外网设备“接入整张内网”，而是通过 `桌面客户端 + 统一访问网关 + Web 管理后台` 的方式，让已登录、已授权、已绑定设备的用户安全访问指定的私有 Web 服务。

第一阶段范围已经明确为：

- 只支持 Web 类私有服务访问，例如 `GitLab`、`Jenkins`、内部管理后台、文档系统，以及以 `HTTP/HTTPS` 暴露的 Docker 内服务
- 用户必须安装客户端并登录，且设备处于允许状态后，才能访问被授权的服务
- 授权模型采用“角色为主 + 用户级允许/禁止覆盖”
- 服务端使用 `Go`
- 桌面客户端使用 `Electron + TypeScript + React + shadcn/ui + Tailwind CSS v4 + pnpm`
- Web 管理后台与客户端共用同一套前端设计系统与技术栈
- 明确禁止实现全局 VPN、系统代理接管、系统 DNS 接管、TUN/TAP、透明代理

## 文档地图

### 产品与范围

- [项目背景与需求说明](/Users/kittors/Developer/opensource/Bifrost/docs/01-product/background-and-requirements.md)

### 架构设计

- [系统架构设计](/Users/kittors/Developer/opensource/Bifrost/docs/02-architecture/system-architecture.md)
- [Monorepo 与项目结构设计](/Users/kittors/Developer/opensource/Bifrost/docs/02-architecture/monorepo-structure.md)
- [客户端架构与页面流转](/Users/kittors/Developer/opensource/Bifrost/docs/02-architecture/client-architecture.md)
- [后台架构与页面流转](/Users/kittors/Developer/opensource/Bifrost/docs/02-architecture/admin-architecture.md)
- [数据流与访问链路](/Users/kittors/Developer/opensource/Bifrost/docs/02-architecture/data-flow.md)
- [数据库实体与表结构设计](/Users/kittors/Developer/opensource/Bifrost/docs/02-architecture/database-entity-design.md)

### 安全与访问控制

- [安全模型与信任边界](/Users/kittors/Developer/opensource/Bifrost/docs/03-security/security-model.md)
- [客户端本地存储与设备信任](/Users/kittors/Developer/opensource/Bifrost/docs/03-security/client-local-storage-and-device-trust.md)

### 设计系统

- [设计原则与界面风格](/Users/kittors/Developer/opensource/Bifrost/docs/04-design-system/design-principles.md)
- [设计 Token 与主题规范](/Users/kittors/Developer/opensource/Bifrost/docs/04-design-system/design-tokens.md)
- [Tailwind CSS v4 与 shadcn/ui 规范](/Users/kittors/Developer/opensource/Bifrost/docs/04-design-system/tailwind-v4-guidelines.md)
- [Electron 客户端 UI 规范](/Users/kittors/Developer/opensource/Bifrost/docs/04-design-system/electron-ui-guidelines.md)
- [后台管理端 UI 规范](/Users/kittors/Developer/opensource/Bifrost/docs/04-design-system/admin-ui-guidelines.md)

### API 契约

- [统一 API 响应结构规范](/Users/kittors/Developer/opensource/Bifrost/docs/05-api/api-response-standard.md)
- [错误码枚举规范](/Users/kittors/Developer/opensource/Bifrost/docs/05-api/error-code-standard.md)
- [后台 API 列表设计](/Users/kittors/Developer/opensource/Bifrost/docs/05-api/admin-api-design.md)
- [客户端 API 列表设计](/Users/kittors/Developer/opensource/Bifrost/docs/05-api/client-api-design.md)
- [网关访问 API 与代理规则](/Users/kittors/Developer/opensource/Bifrost/docs/05-api/gateway-access-api.md)
- [审计事件字典](/Users/kittors/Developer/opensource/Bifrost/docs/05-api/audit-event-dictionary.md)

### 工程规范

- [开发规范与协作原则](/Users/kittors/Developer/opensource/Bifrost/docs/06-engineering/development-standards.md)
- [依赖版本基线](/Users/kittors/Developer/opensource/Bifrost/docs/06-engineering/dependency-version-baseline.md)
- [测试框架与质量策略](/Users/kittors/Developer/opensource/Bifrost/docs/06-engineering/testing-strategy.md)
- [本地多容器开发与联调环境](/Users/kittors/Developer/opensource/Bifrost/docs/06-engineering/local-docker-development.md)

### 部署与路线图

- [部署架构与运行环境](/Users/kittors/Developer/opensource/Bifrost/docs/07-deployment/deployment-architecture.md)
- [第一阶段实施路线图](/Users/kittors/Developer/opensource/Bifrost/docs/08-roadmap/phase-1-web-access.md)
- [详细开发清单](/Users/kittors/Developer/opensource/Bifrost/docs/08-roadmap/development-checklist.md)

## 阅读顺序建议

### 给产品负责人或项目发起人

1. 项目背景与需求说明
2. 系统架构设计
3. 安全模型与信任边界

### 给前端开发

1. Monorepo 与项目结构设计
2. 设计原则与界面风格
3. 设计 Token 与主题规范
4. Tailwind CSS v4 与 shadcn/ui 规范
5. 统一 API 响应结构规范

### 给服务端开发

1. 系统架构设计
2. 安全模型与信任边界
3. 统一 API 响应结构规范
4. 开发规范与协作原则
5. 依赖版本基线

## 当前阶段说明

当前仓库阶段为“设计与规划阶段”，这些文档用于确定：

- 项目目标与边界
- 技术栈与版本基线
- 系统架构与模块边界
- 视觉与交互规范
- API 契约规范
- 开发协作与约束

在未明确进入实现阶段前，不应根据这些文档直接生成脚手架或业务代码。
