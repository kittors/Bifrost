import { spawn } from "node:child_process";

import { e2eEnvironment, e2ePorts } from "./e2e-env.mjs";
import { waitForHTTP } from "./wait-for-http.mjs";

const infrastructureServices = [
  "postgres",
  "mock-gitlab",
  "mock-jenkins",
  "mock-docs",
  "mock-internal-admin",
];

function run(command, args, options = {}) {
  return new Promise((resolve, reject) => {
    const child = spawn(command, args, {
      cwd: process.cwd(),
      env: e2eEnvironment(options.env),
      stdio: "inherit",
      ...options,
    });

    child.on("error", reject);
    child.on("exit", (code, signal) => {
      if (code === 0) {
        resolve();
        return;
      }

      reject(new Error(`${command} ${args.join(" ")} failed with ${signal ?? code}`));
    });
  });
}

// 后端专用环境不启动 Admin Web 和 Desktop，只保留数据库、Gateway 与私有服务 mock。
await run("docker", ["compose", "up", "-d", "--wait", "--build", ...infrastructureServices]);
await run("node", ["./scripts/testing/e2e-seed.mjs"]);
await run("docker", ["compose", "up", "-d", "--wait", "--build", "gateway"]);

await waitForHTTP(`http://127.0.0.1:${e2ePorts.gateway}/readyz`);

console.log(`Bifrost backend test environment is ready: http://127.0.0.1:${e2ePorts.gateway}`);
