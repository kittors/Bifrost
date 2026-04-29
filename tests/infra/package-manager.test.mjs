import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { readFileSync } from "node:fs";
import test from "node:test";
import { fileURLToPath } from "node:url";

const packageJson = JSON.parse(
  readFileSync(new URL("../../package.json", import.meta.url), "utf8"),
);
const guardScriptPath = fileURLToPath(
  new URL("../../scripts/package-manager/ensure-pnpm.mjs", import.meta.url),
);

function runGuard(userAgent) {
  return spawnSync(process.execPath, [guardScriptPath], {
    encoding: "utf8",
    env: {
      ...process.env,
      npm_config_user_agent: userAgent,
    },
  });
}

test("root package requires pnpm for dependency installation", () => {
  assert.equal(packageJson.packageManager, "pnpm@10.33.0");
  assert.equal(packageJson.scripts.preinstall, "node ./scripts/package-manager/ensure-pnpm.mjs");
});

test("root package approves native build scripts required by the desktop client", () => {
  assert.deepEqual(packageJson.pnpm?.onlyBuiltDependencies, [
    "electron",
    "electron-winstaller",
    "esbuild",
  ]);
});

test("package manager guard accepts pnpm and rejects other installers", () => {
  const pnpm = runGuard("pnpm/10.33.0 npm/? node/v24.11.1 darwin arm64");
  assert.equal(pnpm.status, 0, pnpm.stderr);

  for (const userAgent of [
    "npm/11.6.2 node/v24.11.1 darwin arm64 workspaces/false",
    "yarn/4.12.0 npm/? node/v24.11.1 darwin arm64",
    "bun/1.3.4 npm/? node/v24.11.1 darwin arm64",
  ]) {
    const result = runGuard(userAgent);
    assert.notEqual(result.status, 0, `${userAgent} should be rejected`);
    assert.match(result.stderr, /Please install dependencies with pnpm/);
  }
});
