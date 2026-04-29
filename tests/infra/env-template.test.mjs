import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import test from "node:test";

const envExample = readFileSync(new URL("../../.env.example", import.meta.url), "utf8");
const adminViteConfig = readFileSync(
  new URL("../../apps/admin/vite.config.ts", import.meta.url),
  "utf8",
);

function expectActiveEntry(name, value) {
  assert.match(envExample, new RegExp(`^${name}=${value.replaceAll(".", "\\.")}$`, "m"));
}

function expectCommentedEntry(name, valuePattern) {
  assert.match(envExample, new RegExp(`^# ${name}=${valuePattern}$`, "m"));
}

test("root env example documents local development runtime and ports", () => {
  assert.match(envExample, /# 根目录环境变量模板/);
  assert.match(envExample, /# 本地开发环境/);
  assert.match(envExample, /# 本地应用端口/);
  assert.match(envExample, /# 远端 dev 后端/);

  expectActiveEntry("BIFROST_ENV", "development");
  expectActiveEntry("BIFROST_PUBLIC_BASE_URL", "http://142.171.208.80:18080");
  expectActiveEntry("BIFROST_ADMIN_BASE_URL", "http://127.0.0.1:5173");
  expectActiveEntry("VITE_GATEWAY_BASE_URL", "http://142.171.208.80:18080");
  expectActiveEntry("BIFROST_REMOTE_DEV_GATEWAY_URL", "http://142.171.208.80:18080");
  expectActiveEntry("BIFROST_ADMIN_DEV_PORT", "5173");
  expectActiveEntry("BIFROST_DESKTOP_DEV_PORT", "22473");
  expectActiveEntry("PORT", "8080");
  expectActiveEntry("BIFROST_DEV_GATEWAY_PORT", "8080");
  expectActiveEntry("BIFROST_DEV_ADMIN_PORT", "5173");
  expectActiveEntry("BIFROST_DEV_POSTGRES_PORT", "5432");
});

test("root env example documents optional docker test ports without making them default", () => {
  assert.match(envExample, /# 测试环境/);
  expectCommentedEntry("BIFROST_DEV_GATEWAY_PORT", "18080");
  expectCommentedEntry("BIFROST_DEV_ADMIN_PORT", "15173");
  expectCommentedEntry("BIFROST_DEV_POSTGRES_PORT", "15432");
  expectCommentedEntry(
    "BIFROST_DATABASE_TEST_URL",
    "postgres://bifrost:bifrost@127\\.0\\.0\\.1:15432/postgres\\?sslmode=disable",
  );
});

test("root env example keeps production values documented as non-committable placeholders", () => {
  assert.match(envExample, /# 生产环境/);
  expectCommentedEntry("BIFROST_ENV", "production");
  expectCommentedEntry("BIFROST_PUBLIC_BASE_URL", "https://gateway\\.example\\.com");
  expectCommentedEntry("BIFROST_ADMIN_BASE_URL", "https://admin\\.example\\.com");
  expectCommentedEntry(
    "BIFROST_DATABASE_URL",
    "postgres://bifrost:<secret>@postgres:5432/bifrost\\?sslmode=require",
  );
  expectCommentedEntry("BIFROST_TOKEN_SECRET", "<use-a-secret-manager-value-at-least-32-chars>");
  expectCommentedEntry("BIFROST_DEV_DEPLOY_KEY", "<github-actions-secret-only>");
});

test("root env example comments stay readable for Chinese-speaking developers", () => {
  const forbiddenEnglishCommentFragments = [
    "Copy this file",
    "Keep real production secrets",
    "Local development",
    "Remote dev backend",
    "Application ports",
    "Optional local Docker ports",
    "Gateway runtime",
    "Test environment",
    "Remote dev deployment",
    "Production environment",
  ];

  for (const fragment of forbiddenEnglishCommentFragments) {
    assert.doesNotMatch(envExample, new RegExp(`# .*${fragment}`));
  }
});

test("admin development port is driven by the root env file", () => {
  assert.match(adminViteConfig, /envDir:\s*repositoryRoot/);
  assert.match(adminViteConfig, /loadEnv\(mode,\s*repositoryRoot,\s*""\)/);
  assert.match(adminViteConfig, /BIFROST_ADMIN_DEV_PORT/);
  assert.match(adminViteConfig, /port:\s*adminDevPort/);
  assert.match(adminViteConfig, /strictPort:\s*true/);
});
