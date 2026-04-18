import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";
import { readFileSync } from "node:fs";
import test from "node:test";

const requiredServices = [
  "postgres",
  "gateway",
  "admin-web",
  "mock-gitlab",
  "mock-jenkins",
  "mock-docs",
  "mock-internal-admin",
];

function loadComposeConfig() {
  const raw = execFileSync("docker", ["compose", "config", "--format", "json"], {
    cwd: process.cwd(),
    encoding: "utf8",
  });

  return JSON.parse(raw);
}

test("docker compose defines the required local development services", () => {
  const config = loadComposeConfig();
  const services = Object.keys(config.services ?? {});

  for (const serviceName of requiredServices) {
    assert.ok(services.includes(serviceName), `missing service ${serviceName}`);
  }
});

test("docker compose exposes host entrypoints and health checks", () => {
  const config = loadComposeConfig();
  const services = config.services ?? {};

  assert.deepEqual(services.gateway?.ports, [
    { mode: "ingress", target: 8080, published: "8080", protocol: "tcp" },
  ]);
  assert.deepEqual(services["admin-web"]?.ports, [
    { mode: "ingress", target: 5173, published: "5173", protocol: "tcp" },
  ]);

  for (const serviceName of requiredServices) {
    assert.ok(services[serviceName]?.healthcheck, `missing healthcheck for ${serviceName}`);
  }
});

test("root scripts and env example expose local infrastructure commands", () => {
  const packageJson = JSON.parse(
    readFileSync(new URL("../../package.json", import.meta.url), "utf8"),
  );
  const envExample = readFileSync(new URL("../../.env.example", import.meta.url), "utf8");

  assert.equal(
    packageJson.scripts["dev:infra"],
    "docker compose up -d postgres gateway admin-web mock-gitlab mock-jenkins mock-docs mock-internal-admin",
  );
  assert.equal(packageJson.scripts["dev:infra:down"], "docker compose down -v --remove-orphans");
  assert.match(envExample, /^BIFROST_PUBLIC_BASE_URL=/m);
  assert.match(envExample, /^BIFROST_ADMIN_BASE_URL=/m);
  assert.match(envExample, /^BIFROST_DATABASE_URL=/m);
});

test("root scripts expose docker-driven e2e orchestration commands", () => {
  const packageJson = JSON.parse(
    readFileSync(new URL("../../package.json", import.meta.url), "utf8"),
  );

  assert.equal(packageJson.scripts["test:e2e"], "node ./scripts/testing/e2e-run.mjs");
  assert.equal(packageJson.scripts["test:e2e:up"], "node ./scripts/testing/e2e-up.mjs");
  assert.equal(packageJson.scripts["test:e2e:seed"], "node ./scripts/testing/e2e-seed.mjs");
  assert.equal(packageJson.scripts["test:e2e:down"], "node ./scripts/testing/e2e-down.mjs");
});
