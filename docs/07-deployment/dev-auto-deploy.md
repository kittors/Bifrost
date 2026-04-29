# Dev 自动部署

`dev` 分支推送后会触发 `.github/workflows/deploy-dev.yml`，通过 SSH 将当前仓库内容上传到 `root@142.171.208.80:48222`，并在服务器 `/opt/bifrost-dev/current` 执行 `deploy/dev/deploy.sh`。

## 服务器运行方式

- Gateway 暴露 `http://142.171.208.80:18080`，用于桌面端和 API 连通性测试。
- Postgres 与 `mock-gitlab`、`mock-jenkins`、`mock-docs`、`mock-internal-admin` 都只在 Docker 私有网络 `bifrost-private` 内可见，不映射宿主机端口。
- 部署脚本会在 `/opt/bifrost-dev/shared/dev.env` 首次生成数据库密码和 Token Secret，后续部署复用该文件。
- 每次部署都会构建 Gateway 镜像，执行数据库 migrate 和 seed，再启动 Gateway 并检查健康状态。

## GitHub Secret

仓库需要配置 `BIFROST_DEV_DEPLOY_KEY`，内容是能登录 `root@142.171.208.80:48222` 的私钥。该私钥只用于部署工作流，不写入仓库。

## 线上测试步骤

1. 功能分支测试通过后合并到 `dev`。
2. 推送 `dev`，等待 `Deploy Dev` GitHub Action 完成。
3. 确认 `http://142.171.208.80:18080/healthz` 返回成功。
4. 使用本地桌面端或 API 通过 Gateway 访问 `/s/gitlab`、`/s/jenkins`、`/s/docs` 等服务。
5. 不直接访问 mock 服务；这些服务没有公网端口，只能通过 Bifrost Gateway 代理访问。
