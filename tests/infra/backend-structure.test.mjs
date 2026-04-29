import assert from "node:assert/strict";
import { readdirSync, readFileSync, statSync } from "node:fs";
import { join, relative } from "node:path";
import test from "node:test";

const repositoryRoot = new URL("../..", import.meta.url).pathname;
const gatewayRoot = join(repositoryRoot, "apps/gateway");
const guardedSourceRoots = [
  join(repositoryRoot, "apps/gateway"),
  join(repositoryRoot, "packages/contracts"),
  join(repositoryRoot, "scripts"),
];

function listFiles(directory, predicate) {
  const entries = readdirSync(directory, { withFileTypes: true });
  const files = [];

  for (const entry of entries) {
    const path = join(directory, entry.name);
    if (entry.isDirectory()) {
      files.push(...listFiles(path, predicate));
      continue;
    }
    if (predicate(path)) {
      files.push(path);
    }
  }

  return files;
}

function lineCount(path) {
  return readFileSync(path, "utf8").split("\n").length;
}

test("gateway source files stay below the backend file-size guardrail", () => {
  const oversizedFiles = listFiles(gatewayRoot, (path) => path.endsWith(".go"))
    .map((path) => ({ path, lines: lineCount(path) }))
    .filter((file) => file.lines > 500);

  assert.deepEqual(
    oversizedFiles.map((file) => `${relative(repositoryRoot, file.path)}:${file.lines}`),
    [],
  );
});

test("client proxy auth service is split into focused files", () => {
  const proxyCoordinator = join(gatewayRoot, "internal/auth/service_client_proxy.go");

  assert.ok(
    lineCount(proxyCoordinator) <= 260,
    "service_client_proxy.go should stay as a coordinator; move policy/repository helpers into focused files",
  );
  assert.ok(
    statSync(join(gatewayRoot, "internal/auth/service_client_policy.go")).isFile(),
    "client service policy resolution belongs in service_client_policy.go",
  );
  assert.ok(
    statSync(join(gatewayRoot, "internal/auth/service_client_repository.go")).isFile(),
    "client service loading belongs in service_client_repository.go",
  );
});

test("OpenAPI contract documents implemented gateway routes", () => {
  const openapi = readFileSync(
    join(repositoryRoot, "packages/contracts/openapi/bifrost.v1.yaml"),
    "utf8",
  );
  const requiredRoutes = [
    "/healthz:",
    "/readyz:",
    "/api/v1/admin/auth/login:",
    "/api/v1/admin/auth/refresh:",
    "/api/v1/admin/auth/logout:",
    "/api/v1/admin/auth/me:",
    "/api/v1/admin/users:",
    "/api/v1/admin/users/{userId}:",
    "/api/v1/admin/users/{userId}/service-overrides:",
    "/api/v1/admin/users/{userId}/reset-password:",
    "/api/v1/admin/users/{userId}/status:",
    "/api/v1/admin/roles:",
    "/api/v1/admin/roles/{roleId}:",
    "/api/v1/admin/roles/{roleId}/services:",
    "/api/v1/admin/services:",
    "/api/v1/admin/services/{serviceId}:",
    "/api/v1/admin/services/{serviceId}/status:",
    "/api/v1/admin/devices:",
    "/api/v1/admin/devices/{deviceId}:",
    "/api/v1/admin/devices/{deviceId}/status:",
    "/api/v1/admin/audit-events:",
    "/api/v1/client/devices/bootstrap:",
    "/api/v1/client/auth/login:",
    "/api/v1/client/auth/refresh:",
    "/api/v1/client/auth/logout:",
    "/api/v1/client/me:",
    "/api/v1/client/devices/register:",
    "/api/v1/client/devices/challenge:",
    "/api/v1/client/devices/challenge/verify:",
    "/api/v1/client/services:",
    "/api/v1/client/services/{serviceId}:",
    "/api/v1/client/services/{serviceId}/access-url:",
    "/s/{serviceKey}/{proxyPath}:",
  ];
  const lines = openapi.split("\n");

  assert.doesNotMatch(openapi, /^paths:\s*\{\}\s*$/m);
  for (const route of requiredRoutes) {
    assert.ok(lines.includes(`  ${route}`), `missing OpenAPI route ${route}`);
  }
});

test("backend validation can run infra checks without inherited e2e port overrides", async () => {
  const { withoutE2EPortOverrides } = await import("../../scripts/testing/e2e-env.mjs");
  const cleaned = withoutE2EPortOverrides({
    BIFROST_DEV_ADMIN_PORT: "15174",
    BIFROST_DEV_GATEWAY_PORT: "19080",
    BIFROST_DEV_POSTGRES_PORT: "15433",
    KEEP_ME: "yes",
  });

  assert.deepEqual(cleaned, { KEEP_ME: "yes" });

  const backendRun = readFileSync(join(repositoryRoot, "scripts/testing/backend-run.mjs"), "utf8");
  assert.match(backendRun, /withoutE2EPortOverrides/);
  assert.match(backendRun, /useE2EEnv === false/);
});

test("production source does not contain dangerous placeholders", () => {
  const forbiddenPattern = /\b(TODO|FIXME|TBD)\b|not implemented|dev-only|change-me/i;
  const sourceFiles = guardedSourceRoots.flatMap((root) =>
    listFiles(root, (path) => {
      if (path.includes("node_modules") || path.includes(".turbo")) {
        return false;
      }
      if (path.endsWith("_test.go") || path.endsWith(".test.ts") || path.endsWith(".test.mjs")) {
        return false;
      }
      return [".go", ".ts", ".mjs", ".yaml", ".yml"].some((extension) => path.endsWith(extension));
    }),
  );

  const offenders = sourceFiles.flatMap((path) => {
    const lines = readFileSync(path, "utf8").split("\n");
    return lines
      .map((line, index) => ({ index, line }))
      .filter(({ line }) => forbiddenPattern.test(line))
      .map(({ index, line }) => `${relative(repositoryRoot, path)}:${index + 1}:${line.trim()}`);
  });

  assert.deepEqual(offenders, []);
});

test("runtime documentation exposes production safety controls", () => {
  const envExample = readFileSync(join(repositoryRoot, ".env.example"), "utf8");
  const runtimeParameters = readFileSync(
    join(repositoryRoot, "docs/07-deployment/service-runtime-parameters.md"),
    "utf8",
  );

  assert.match(envExample, /^BIFROST_ENV=development$/m);
  assert.match(envExample, /^BIFROST_TOKEN_SECRET=/m);
  assert.match(runtimeParameters, /BIFROST_ENV/);
  assert.match(runtimeParameters, /production/);
  assert.match(runtimeParameters, /at least 32 characters|32 个字符以上/);
});
