import { execFile } from "node:child_process";
import { promisify } from "node:util";

import { e2eEnvironment } from "./e2e-env.mjs";

const execFileAsync = promisify(execFile);

// 测试环境清理统一走 docker compose down，确保数据库卷和残留容器一起移除。
await execFileAsync("docker", ["compose", "down", "-v", "--remove-orphans"], {
  cwd: process.cwd(),
  env: e2eEnvironment(),
});
