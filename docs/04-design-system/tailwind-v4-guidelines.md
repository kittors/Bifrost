# Tailwind CSS v4 与 shadcn/ui 规范

## 1. 总目标

本项目要求尽可能少写原生 CSS，样式实现以 `Tailwind CSS v4` 为中心。所有设计 token 通过 Tailwind v4 的 `CSS-first` 方式管理，组件样式以 utility class 为主，shadcn/ui 作为开放源码组件基底。

## 2. 技术原则

### 2.1 Tailwind v4 优先

本项目前端样式系统必须基于 Tailwind CSS v4。

原因：

- v4 支持 `@theme` 方式定义 token
- token 可以自动映射为 CSS 变量
- 更适合将设计系统集中维护在少量样式入口中
- 能减少 `tailwind.config.js` 式的额外心智负担

### 2.2 shadcn/ui 作为开放源码组件基底

shadcn/ui 不是传统“装个库就结束”的黑盒组件库，它会把组件源码复制到项目中。项目必须：

- 以 shadcn/ui 为基础建立自有组件体系
- 统一改造样式与 token
- 不直接在业务页面复制粘贴不同版本组件

## 3. 允许存在的 CSS 文件

CSS 文件必须被严格控制，推荐只有以下几类：

- `packages/design-tokens/src/theme.css`
- `packages/design-tokens/src/base.css`
- `packages/design-tokens/src/app.css`
- 特殊内容渲染所需的极少量附加样式，例如 `prose.css`

### 3.1 theme.css

负责：

- `@import "tailwindcss"`
- `@theme` token 定义
- 主题变量分层

### 3.2 base.css

负责：

- 全局 reset 的必要补充
- 默认文本渲染
- 焦点样式
- 选中文本
- 滚动条策略
- 颜色模式根节点行为

### 3.3 app.css

负责：

- 组合导入 `theme.css` 与 `base.css`
- 应用全局入口

## 4. 禁止项

以下技术和行为在本项目中默认禁止：

- `SCSS`
- `Less`
- `styled-components`
- `emotion`
- 页面级 CSS Module
- 在业务页面写大段原生 CSS
- 在任意目录新增随意命名的 `.css` 文件承担业务样式

## 5. 组件样式编写规则

### 5.1 utility-first

组件样式应优先通过 Tailwind utilities 完成：

- 布局
- 间距
- 颜色
- 边框
- 字号
- 圆角
- 状态变化
- 亮暗主题切换

### 5.2 组件抽象优先于自定义 class

如果同一组 class 在三个以上位置重复出现，应：

- 抽成共享组件
- 或抽成受控的样式工具函数

而不是：

- 临时写一个自定义 class
- 或在多个页面继续重复 class 串

### 5.3 少量原生 CSS 的合理场景

以下场景允许少量原生 CSS：

- `@theme` token 声明
- `@layer base` 中的全局基础样式
- Markdown 或富文本内容的受控样式
- 浏览器无法用 utilities 干净表达的极少数系统行为

这些 CSS 也必须建立在 token 之上。

## 6. 主题管理

推荐采用统一主题根节点控制，例如：

- 根节点设置 `data-theme="light"` 或 `data-theme="dark"`
- 所有颜色都从 token 推导
- 组件内部不自行维护“暗色版颜色”

禁止：

- 每个组件单独写一套暗色判断
- 在页面里硬编码亮暗模式颜色值

## 7. 类名策略

推荐：

- 使用 `cn()` 或等价工具组合 class
- 对变体组件使用受控 variant 方案
- 对复杂状态使用语义化 props，而不是随意拼接 class

例如：

- `intent: default | primary | danger`
- `size: xs | sm | md | lg`
- `density: compact | default`

## 8. shadcn/ui 使用规范

### 8.1 单一来源

基础组件应集中维护在共享 `ui` 包中，而不是客户端和后台各有一套。

### 8.2 改造策略

引入 shadcn/ui 后必须做以下动作：

- 对齐设计 token
- 对齐按钮、输入、表格、弹层的尺寸体系
- 清理不符合本项目视觉方向的默认样式
- 强化亮暗主题一致性

### 8.3 不允许的行为

- 不允许客户端和后台各自随意覆盖同名组件
- 不允许将 shadcn 默认配色直接照搬为产品配色
- 不允许在业务代码中堆大量一次性组件变体

## 9. 工程实践建议

### 9.1 token 先行

先建立 token，再搭组件，再做页面。不要先拼页面，后补规范。

### 9.2 组件清单先行

第一阶段优先沉淀以下组件：

- Button
- Input
- Select
- Checkbox
- Switch
- Badge
- Tabs
- Dialog
- Drawer
- Table
- Empty State
- Alert
- Toast
- Search Filter Bar
- Status Pill

### 9.3 页面样式来源限制

页面只能组合：

- 共享 token
- 共享基础组件
- 受控业务组件

不应成为自由写样式的地方。

## 10. 验收清单

每个新增页面或组件都应检查：

- 是否只使用了既定 token
- 是否仅通过 Tailwind utilities 完成主要样式
- 是否在 Light 与 Dark 下都可读
- 是否符合统一高度、字号、圆角、边框规范
- 是否可被客户端与后台共享或复用
