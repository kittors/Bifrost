import assert from "node:assert/strict";
import { existsSync, readFileSync } from "node:fs";
import { join } from "node:path";
import test from "node:test";

const repositoryRoot = new URL("../..", import.meta.url).pathname;

test("AGENTS records the remote dev server deployment contract", () => {
  const agents = readFileSync(join(repositoryRoot, "AGENTS.md"), "utf8");

  assert.match(agents, /ssh -p 48222 root@142\.171\.208\.80/);
  assert.match(agents, /\/opt\/bifrost-dev\/current/);
  assert.match(agents, /\/opt\/bifrost-dev\/shared\/dev\.env/);
  assert.match(agents, /http:\/\/142\.171\.208\.80:18080/);
  assert.match(agents, /\.github\/workflows\/deploy-dev\.yml/);
  assert.match(agents, /推送到 `dev` 分支/);
  assert.match(agents, /GitHub Action/);
  assert.match(agents, /接口检查/);
});

test("repository exposes a repeatable remote dev API smoke command", () => {
  const packageJSON = JSON.parse(readFileSync(join(repositoryRoot, "package.json"), "utf8"));
  const scriptPath = join(repositoryRoot, "scripts/dev/remote-api-smoke.mjs");

  assert.equal(packageJSON.scripts["test:remote-api"], "node ./scripts/dev/remote-api-smoke.mjs");
  assert.ok(existsSync(scriptPath), "scripts/dev/remote-api-smoke.mjs is required");

  const script = readFileSync(scriptPath, "utf8");
  assert.match(script, /BIFROST_REMOTE_DEV_GATEWAY_URL/);
  assert.match(script, /http:\/\/142\.171\.208\.80:18080/);
  assert.match(script, /\/healthz/);
  assert.match(script, /\/readyz/);
  assert.match(script, /\/api\/v1\/admin\/auth\/login/);
  assert.match(script, /\/api\/v1\/admin\/users/);
  assert.match(script, /\/api\/v1\/client\/devices\/bootstrap/);
  assert.match(script, /\/api\/v1\/client\/services/);
  assert.match(script, /\/debug\/upstreams\/gitlab/);
});
