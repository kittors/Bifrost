import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";
import { existsSync, readFileSync } from "node:fs";
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

test("docker compose uses the Postgres 18 compatible data volume target", () => {
  const config = loadComposeConfig();
  const services = config.services ?? {};

  assert.deepEqual(
    services.postgres?.volumes?.map((volume) => volume.target),
    ["/var/lib/postgresql"],
  );
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
  assert.equal(
    packageJson.scripts["build:gateway:image"],
    "docker build -f apps/gateway/Dockerfile -t bifrost/gateway:dev .",
  );
  assert.equal(
    packageJson.scripts["build:admin:image"],
    "docker build -f docker/admin-web/Dockerfile -t bifrost/admin-web:dev .",
  );
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

test("root scripts expose backend-only environment and validation commands", () => {
  const packageJson = JSON.parse(
    readFileSync(new URL("../../package.json", import.meta.url), "utf8"),
  );

  assert.equal(packageJson.scripts["dev:backend"], "node ./scripts/dev/remote-backend.mjs");
  assert.equal(packageJson.scripts["dev:backend:local"], "node ./scripts/testing/backend-up.mjs");
  assert.equal(
    packageJson.scripts["dev:backend:local:down"],
    "node ./scripts/testing/e2e-down.mjs",
  );
  assert.equal(packageJson.scripts["dev:backend:down"], "node ./scripts/testing/e2e-down.mjs");
  assert.equal(packageJson.scripts["test:backend"], "node ./scripts/testing/backend-run.mjs");
  assert.ok(
    existsSync(new URL("../../scripts/testing/backend-up.mjs", import.meta.url)),
    "backend-up script is required",
  );
  assert.ok(
    existsSync(new URL("../../scripts/testing/backend-run.mjs", import.meta.url)),
    "backend-run script is required",
  );
});

test("admin web image builds the real Vite application", () => {
  const dockerfile = readFileSync(
    new URL("../../docker/admin-web/Dockerfile", import.meta.url),
    "utf8",
  );

  assert.ok(
    existsSync(new URL("../../docker/admin-web/nginx.conf", import.meta.url)),
    "admin nginx runtime config is required",
  );
  assert.match(dockerfile, /COPY scripts\/package-manager \.\/scripts\/package-manager/);
  assert.match(dockerfile, /pnpm --filter @bifrost\/admin build/);
  assert.match(dockerfile, /COPY --from=builder .*apps\/admin\/dist/);
});

test("desktop package scripts cover macOS Windows and Linux installers", () => {
  const desktopPackageJson = JSON.parse(
    readFileSync(new URL("../../apps/desktop/package.json", import.meta.url), "utf8"),
  );
  const desktopBuilderConfig = readFileSync(
    new URL("../../apps/desktop/electron-builder.yml", import.meta.url),
    "utf8",
  );

  assert.equal(desktopPackageJson.devDependencies["electron-builder"], "26.8.1");
  assert.ok(desktopPackageJson.main, "desktop main entry is required for electron-builder");
  assert.match(desktopPackageJson.homepage, /^https:\/\/github\.com\/kittors\/Bifrost/);
  assert.equal(desktopPackageJson.author.email, "14817208+kittors@users.noreply.github.com");
  assert.equal(typeof desktopPackageJson.scripts["dist:mac"], "string");
  assert.equal(typeof desktopPackageJson.scripts["dist:win"], "string");
  assert.equal(typeof desktopPackageJson.scripts["dist:linux"], "string");
  assert.equal(typeof desktopPackageJson.scripts["dist:dir"], "string");
  assert.match(
    desktopBuilderConfig,
    /maintainer:\s*"Kittors <14817208\+kittors@users\.noreply\.github\.com>"/,
  );
});

test("github actions workflows exist for CI gate and desktop artifacts", () => {
  const ciWorkflowURL = new URL("../../.github/workflows/ci.yml", import.meta.url);
  const desktopWorkflowURL = new URL(
    "../../.github/workflows/desktop-packages.yml",
    import.meta.url,
  );

  assert.ok(existsSync(ciWorkflowURL), "ci workflow is required");
  assert.ok(existsSync(desktopWorkflowURL), "desktop artifact workflow is required");

  const ciWorkflow = readFileSync(ciWorkflowURL, "utf8");
  const desktopWorkflow = readFileSync(desktopWorkflowURL, "utf8");

  for (const command of [
    "pnpm lint",
    "pnpm check",
    "pnpm test",
    "pnpm test:infra",
    "pnpm test:e2e",
  ]) {
    assert.match(ciWorkflow, new RegExp(command.replaceAll(" ", "\\s+")));
  }
  assert.match(ciWorkflow, /go test \.\/\.\.\./);
  assert.match(ciWorkflow, /FORCE_JAVASCRIPT_ACTIONS_TO_NODE24:\s*"true"/);
  assert.match(ciWorkflow, /cache-dependency-path:\s*apps\/gateway\/go\.sum/);
  assert.match(ciWorkflow, /corepack prepare pnpm@10\.33\.0 --activate/);
  assert.match(ciWorkflow, /actions\/upload-artifact@v7/);
  assert.match(ciWorkflow, /test-results\/playwright/);
  assert.match(ciWorkflow, /retention-days:\s*7/);
  assert.match(desktopWorkflow, /FORCE_JAVASCRIPT_ACTIONS_TO_NODE24:\s*"true"/);
  assert.match(desktopWorkflow, /corepack prepare pnpm@10\.33\.0 --activate/);
  assert.match(desktopWorkflow, /actions\/upload-artifact@v7/);

  for (const osName of ["macos-latest", "windows-latest", "ubuntu-latest"]) {
    assert.match(desktopWorkflow, new RegExp(osName));
  }
  assert.match(desktopWorkflow, /dist:mac/);
  assert.match(desktopWorkflow, /dist:win/);
  assert.match(desktopWorkflow, /dist:linux/);
  assert.match(desktopWorkflow, /apps\/desktop\/release\/\*\*/);
});
