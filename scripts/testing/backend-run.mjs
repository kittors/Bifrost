import { spawn } from "node:child_process";

import { e2eEnvironment, e2ePorts, withoutE2EPortOverrides } from "./e2e-env.mjs";

function run(command, args, options = {}) {
  const env =
    options.useE2EEnv === false
      ? withoutE2EPortOverrides({ ...process.env, ...options.env })
      : e2eEnvironment(options.env);
  return new Promise((resolve, reject) => {
    const child = spawn(command, args, {
      cwd: options.cwd ?? process.cwd(),
      env,
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
  // 使用干净容器和干净数据库跑后端闭环，避免本地调试状态污染验收结果。
  await run("pnpm", ["test:e2e:down"]);
  await run("pnpm", ["dev:backend"]);
  // infra 测试验证仓库静态配置，应使用未覆写端口的默认环境。
  await run("pnpm", ["test:infra"], { useE2EEnv: false });
  await run("go", ["test", "./..."], {
    cwd: "apps/gateway",
    env: {
      BIFROST_DATABASE_TEST_URL: `postgres://bifrost:bifrost@127.0.0.1:${e2ePorts.postgres}/postgres?sslmode=disable`,
    },
    useE2EEnv: false,
  });
  await run("pnpm", ["exec", "playwright", "test"], {
    env: {
      BIFROST_PUBLIC_BASE_URL: `http://127.0.0.1:${e2ePorts.gateway}`,
    },
    useE2EEnv: false,
  });
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
