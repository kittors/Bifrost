# Bifrost 专业化优化清单

本文记录当前阶段必须落地的工程质量优化项。每一项都要求有代码、测试或文档证据；不能只靠口头判断。

## 当前范围

- 优先处理 `apps/gateway`、`packages/contracts`、`scripts/testing` 和 `tests/infra`。
- 不触碰当前已有大量改动的 Admin / Desktop UI 文件，避免把无关变更混进后端专业化工作。
- 所有行为变更必须先有失败测试，再做实现。

## 优化项

- [x] 生产环境配置必须安全：`BIFROST_ENV=production` 时必须显式配置强 `BIFROST_TOKEN_SECRET`，不得使用开发默认签名密钥。
- [x] 列表 query 参数不得静默兜底：非法 `page` / `pageSize` 必须返回统一 `COMMON_BAD_REQUEST` 响应，避免错误输入被伪装成默认值。
- [x] 后端文件规模必须有门禁：Gateway Go 文件超过 500 行立即触发 infra 测试失败。
- [x] 核心代理授权逻辑必须保持高聚合：客户端服务编排、策略判断、数据读取、代理审计分别在独立文件维护。
- [x] OpenAPI 必须记录已实现路由：不允许 `paths: {}` 回退，也不允许实现路由没有契约记录。
- [x] 后端专用回归必须可重复：默认端口被占用时，覆盖端口运行不能污染静态 infra 检查。
- [x] 危险占位必须被门禁拦截：生产代码中不得出现 `TODO`、`FIXME`、`TBD`、`not implemented`、`dev-only` 默认安全材料。
- [x] 部署文档必须和代码一致：生产配置、安全密钥、后端验证命令都必须有明确说明。

## 验收命令

```bash
pnpm --filter @bifrost/contracts check
pnpm --filter @bifrost/contracts test
pnpm test:infra
BIFROST_DEV_GATEWAY_PORT=19080 BIFROST_DEV_POSTGRES_PORT=15433 BIFROST_DEV_ADMIN_PORT=15174 pnpm test:backend
```
