# 设计 Token 与主题规范

## 1. 规范目标

本文件定义 Bifrost 全部核心设计 token。客户端与后台必须共同使用这一套 token，不允许各自维护平行体系。

建议在 `packages/design-tokens/src/theme.css` 中以 Tailwind CSS v4 的 `@theme` 方式落地。

## 2. 字体体系

## 2.1 字体选择

推荐字体：

- Sans: `Geist`, `Noto Sans SC`, `system-ui`, `sans-serif`
- Mono: `JetBrains Mono`, `SFMono-Regular`, `Consolas`, `monospace`

说明：

- `Geist` 负责主要英文与 UI 观感
- `Noto Sans SC` 提供稳定的中文显示
- `JetBrains Mono` 用于代码、请求 ID、日志字段、设备指纹等

## 2.2 字号等级

系统仅保留 4 个字号级别。

| Token | Font Size | Line Height | Font Weight | 用途 |
|---|---:|---:|---:|---|
| `text-1` | `12px` | `18px` | `500` | 标签、次要说明、表格次级字段 |
| `text-2` | `13px` | `20px` | `500` | 默认正文、列表说明、次级按钮 |
| `text-3` | `14px` | `22px` | `500/600` | 默认交互文本、输入值、主要按钮 |
| `text-4` | `16px` | `24px` | `600` | 页面标题、弹窗标题、重要区块标题 |

补充规则：

- 页面不使用超过 `16px` 的常规 UI 标题层
- 大标题表达由间距与布局层级承担，而不是通过夸张字号实现

## 3. 圆角等级

仅允许 3 个圆角等级：

| Token | 值 | 推荐用途 |
|---|---:|---|
| `radius-sm` | `6px` | 标签、输入框、分段控件、紧凑按钮 |
| `radius-md` | `10px` | 默认按钮、卡片、下拉、表格容器 |
| `radius-lg` | `14px` | 模态框、抽屉、大面板 |

限制：

- 禁止新增第四档及以上常规圆角
- 禁止默认使用胶囊按钮

## 4. 边框与线条

## 4.1 边框粗细

| Token | 值 | 用途 |
|---|---:|---|
| `border-thin` | `1px` | 默认边框、分割线 |
| `ring-focus` | `2px` | 键盘焦点态 |

规则：

- 常规组件边框统一 `1px`
- 焦点态统一 `2px`
- 不使用 `2px` 作为默认输入框边框

## 4.2 阴影

| Token | 值 | 用途 |
|---|---|---|
| `shadow-sm` | 轻微阴影 | Popover、Dropdown |
| `shadow-md` | 中等阴影 | Modal、Drawer |

默认策略：

- 卡片默认无明显阴影
- 优先用背景层和边框区分层级

## 5. 间距体系

推荐使用 `4px` 为基础步进，关键空间如下：

| Token | 值 | 用途 |
|---|---:|---|
| `space-1` | `4px` | 极小间隙 |
| `space-2` | `8px` | 图标与文字、紧凑分组 |
| `space-3` | `12px` | 控件间隙 |
| `space-4` | `16px` | 区块内标准间距 |
| `space-5` | `20px` | 页面内容上边距 |
| `space-6` | `24px` | 页面区块间距 |
| `space-8` | `32px` | 大区块间距 |

规则：

- 页面主区块优先使用 `20px` 或 `24px`
- 表单项之间优先使用 `12px`
- 卡片内边距优先使用 `16px`

## 6. 组件尺寸

## 6.1 按钮

| Token | 高度 | 水平内边距 | 字号 | 用途 |
|---|---:|---:|---:|---|
| `btn-xs` | `28px` | `10px` | `12px` | 表格行内操作、工具按钮 |
| `btn-sm` | `32px` | `12px` | `13px` | 次级按钮、紧凑表单 |
| `btn-md` | `36px` | `14px` | `14px` | 默认按钮 |
| `btn-lg` | `40px` | `16px` | `14px` | 登录主按钮、关键确认 |

补充：

- 图标按钮尺寸与高度对应
- 默认主按钮使用 `btn-md`
- 默认不使用全宽超大按钮，除登录页等必要场景

## 6.2 输入框

| Token | 高度 | 用途 |
|---|---:|---|
| `input-sm` | `32px` | 紧凑筛选器 |
| `input-md` | `36px` | 默认输入框 |
| `input-lg` | `40px` | 登录页、重点输入 |

规则：

- 后台筛选区优先使用 `32px`
- 默认业务表单优先使用 `36px`

## 6.3 表格

| Token | 值 | 用途 |
|---|---:|---|
| `table-row-compact` | `32px` | 高密度后台场景 |
| `table-row-default` | `36px` | 默认列表 |
| `table-header-height` | `36px` | 表头高度 |

## 7. 布局尺寸

## 7.1 客户端

| Token | 值 |
|---|---:|
| `desktop-window-default-width` | `420px` |
| `desktop-window-default-height` | `560px` |
| `desktop-window-min-width` | `380px` |
| `desktop-window-min-height` | `480px` |
| `desktop-card-padding` | `16px` |
| `desktop-section-gap` | `12px` |

客户端定位为“小卡片式轻量入口”，不采用后台式左侧导航与大工作台窗口。

## 7.2 后台

| Token | 值 |
|---|---:|
| `admin-sidebar-width` | `232px` |
| `admin-topbar-height` | `52px` |
| `admin-page-padding-x` | `24px` |
| `admin-page-padding-y` | `20px` |

## 8. 颜色体系

## 8.1 设计原则

- 基础色使用中性灰
- 品牌主色使用冷静蓝青色
- 语义色清晰但不刺眼
- 使用 `OKLCH` 以便在亮暗主题下保持感知一致性

## 8.2 Light Theme

| Token | 值 |
|---|---|
| `bg` | `oklch(0.985 0.002 247)` |
| `surface` | `oklch(0.998 0.001 247)` |
| `surface-2` | `oklch(0.972 0.004 247)` |
| `text-primary` | `oklch(0.22 0.01 255)` |
| `text-secondary` | `oklch(0.47 0.01 255)` |
| `text-muted` | `oklch(0.62 0.008 255)` |
| `border` | `oklch(0.90 0.006 255)` |
| `border-soft` | `oklch(0.94 0.004 255)` |
| `brand` | `oklch(0.60 0.11 244)` |
| `brand-hover` | `oklch(0.55 0.12 244)` |
| `brand-soft` | `oklch(0.95 0.02 244)` |
| `success` | `oklch(0.63 0.13 152)` |
| `warning` | `oklch(0.73 0.15 78)` |
| `danger` | `oklch(0.60 0.18 25)` |
| `info` | `oklch(0.62 0.11 230)` |

## 8.3 Dark Theme

| Token | 值 |
|---|---|
| `bg` | `oklch(0.18 0.008 255)` |
| `surface` | `oklch(0.22 0.01 255)` |
| `surface-2` | `oklch(0.26 0.012 255)` |
| `text-primary` | `oklch(0.94 0.003 255)` |
| `text-secondary` | `oklch(0.75 0.006 255)` |
| `text-muted` | `oklch(0.60 0.008 255)` |
| `border` | `oklch(0.33 0.01 255)` |
| `border-soft` | `oklch(0.28 0.01 255)` |
| `brand` | `oklch(0.72 0.10 244)` |
| `brand-hover` | `oklch(0.77 0.11 244)` |
| `brand-soft` | `oklch(0.30 0.04 244)` |
| `success` | `oklch(0.74 0.12 152)` |
| `warning` | `oklch(0.80 0.13 78)` |
| `danger` | `oklch(0.72 0.17 25)` |
| `info` | `oklch(0.75 0.10 230)` |

## 9. 语义色使用规则

- 主按钮只能使用 `brand`
- 危险删除只能使用 `danger`
- 成功态只用于结果确认或状态展示
- `warning` 只用于风险提醒
- 页面背景不使用语义色大面积染色

## 10. Token 命名建议

推荐在 Tailwind v4 `@theme` 中使用如下命名：

```css
--font-sans
--font-mono

--text-1
--text-2
--text-3
--text-4

--radius-sm
--radius-md
--radius-lg

--color-bg
--color-surface
--color-surface-2
--color-text-primary
--color-text-secondary
--color-text-muted
--color-border
--color-border-soft
--color-brand
--color-brand-hover
--color-brand-soft
--color-success
--color-warning
--color-danger
--color-info
```

## 11. 组件映射建议

- 页面背景使用 `bg`
- 卡片、表格容器、Modal 内容区使用 `surface`
- 次级背景区使用 `surface-2`
- 主文本使用 `text-primary`
- 说明文本使用 `text-secondary` 或 `text-muted`
- 常规边框使用 `border`
- 分割线优先使用 `border-soft`

## 12. 执行约束

- 所有主题值必须定义为 token，不允许散写十六进制颜色
- 所有组件必须从 token 出发，不允许单页自行定义风格
- 亮暗主题的所有关键组件都必须经过对照验证
