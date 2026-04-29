import { spawn } from "node:child_process";

import { e2eEnvironment } from "./e2e-env.mjs";

function run(command, args) {
  return new Promise((resolve, reject) => {
    const child = spawn(command, args, {
      cwd: process.cwd(),
      env: e2eEnvironment(),
      stdio: "inherit",
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

let exitCode = 0;

try {
  // 全量 E2E 必须自带干净基线，避免聚焦测试遗留数据库状态污染回归。
  await run("pnpm", ["test:e2e:down"]);
  await run("pnpm", ["test:e2e:up"]);
  await run("pnpm", ["exec", "playwright", "test"]);
} catch (error) {
  exitCode = 1;
  console.error(error instanceof Error ? error.message : error);
} finally {
  try {
    await run("pnpm", ["test:e2e:down"]);
  } catch (error) {
    exitCode = 1;
    console.error(error instanceof Error ? error.message : error);
  }
}

process.exitCode = exitCode;
