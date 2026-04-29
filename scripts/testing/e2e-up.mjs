import { execFile } from "node:child_process";
import { promisify } from "node:util";

import { e2eEnvironment, e2ePorts } from "./e2e-env.mjs";
import { waitForHTTP } from "./wait-for-http.mjs";

const execFileAsync = promisify(execFile);
const infrastructureServices = [
  "postgres",
  "mock-gitlab",
  "mock-jenkins",
  "mock-docs",
  "mock-internal-admin",
];
const appServices = ["gateway", "admin-web"];

async function runDockerCompose(args) {
  await execFileAsync("docker", ["compose", ...args], {
    cwd: process.cwd(),
    env: e2eEnvironment(),
  });
}

// 分两段启动：先数据库与上游 mock，再迁移 seed，最后启动依赖数据库结构的应用。
await runDockerCompose(["up", "-d", "--wait", "--build", ...infrastructureServices]);
await execFileAsync("node", ["./scripts/testing/e2e-seed.mjs"], {
  cwd: process.cwd(),
  env: e2eEnvironment(),
});
await runDockerCompose(["up", "-d", "--wait", "--build", ...appServices]);

await waitForHTTP(`http://127.0.0.1:${e2ePorts.gateway}/readyz`);
await waitForHTTP(`http://127.0.0.1:${e2ePorts.admin}/health`);
