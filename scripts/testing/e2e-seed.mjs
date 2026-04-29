import { execFile } from "node:child_process";
import { promisify } from "node:util";

import { e2eEnvironment, e2ePorts } from "./e2e-env.mjs";

const execFileAsync = promisify(execFile);

function databaseURL() {
  return process.env.BIFROST_DATABASE_URL?.trim()
    ? process.env.BIFROST_DATABASE_URL.trim()
    : `postgres://bifrost:bifrost@127.0.0.1:${e2ePorts.postgres}/bifrost?sslmode=disable`;
}

async function run(command, args) {
  await execFileAsync(command, args, {
    cwd: process.cwd(),
    env: e2eEnvironment({
      BIFROST_DATABASE_URL: databaseURL(),
    }),
  });
}

// 每次启动 E2E 前先迁移并填充固定种子，保证断言可重复。
await run("pnpm", ["db:migrate"]);
await run("pnpm", ["db:seed"]);
