import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";
import { existsSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";

const repositoryRoot = new URL("../..", import.meta.url).pathname;
const devDeployComposePath = join(repositoryRoot, "deploy/dev/docker-compose.yml");

function loadDevDeployComposeConfig() {
  const envDirectory = mkdtempSync(join(tmpdir(), "bifrost-dev-deploy-"));
  const envFile = join(envDirectory, "dev.env");
  writeFileSync(
    envFile,
    [
      "BIFROST_DEV_POSTGRES_PASSWORD=test-postgres-password",
      "BIFROST_DEV_TOKEN_SECRET=0123456789abcdef0123456789abcdef0123456789abcdef",
      "BIFROST_DEV_GATEWAY_BIND=0.0.0.0",
      "BIFROST_DEV_GATEWAY_PORT=18080",
    ].join("\n"),
  );

  try {
    const raw = execFileSync(
      "docker",
      ["compose", "-f", devDeployComposePath, "--env-file", envFile, "config", "--format", "json"],
      {
        cwd: repositoryRoot,
        encoding: "utf8",
      },
    );

    return JSON.parse(raw);
  } finally {
    rmSync(envDirectory, { force: true, recursive: true });
  }
}

test("dev deploy workflow pushes dev branch builds to the configured SSH host", () => {
  const workflowPath = join(repositoryRoot, ".github/workflows/deploy-dev.yml");

  assert.ok(existsSync(workflowPath), "dev deployment workflow is required");

  const workflow = readFileSync(workflowPath, "utf8");
  assert.match(workflow, /branches:\s*\n\s+- dev/);
  assert.match(workflow, /142\.171\.208\.80/);
  assert.match(workflow, /48222/);
  assert.match(workflow, /BIFROST_DEV_DEPLOY_KEY/);
  assert.match(workflow, /BIFROST_DEV_GATEWAY_PORT:\s*"18080"/);
  assert.match(workflow, /deploy\/dev\/deploy\.sh/);
  assert.match(workflow, /\/opt\/bifrost-dev/);
});

test("dev deploy compose keeps private upstream HTTP services off public ports", () => {
  const config = loadDevDeployComposeConfig();
  const services = config.services ?? {};

  for (const serviceName of [
    "postgres",
    "mock-gitlab",
    "mock-jenkins",
    "mock-docs",
    "mock-internal-admin",
  ]) {
    assert.deepEqual(services[serviceName]?.ports ?? [], [], `${serviceName} must stay private`);
    assert.ok(
      services[serviceName]?.healthcheck,
      `${serviceName} must expose an internal healthcheck`,
    );
  }

  assert.deepEqual(services.gateway?.ports, [
    { mode: "ingress", host_ip: "0.0.0.0", target: 8080, published: "18080", protocol: "tcp" },
  ]);
  assert.deepEqual(
    services.postgres?.volumes?.map((volume) => volume.target),
    ["/var/lib/postgresql"],
  );
  assert.equal(services.gateway?.environment?.BIFROST_ENV, "production");
  assert.match(
    services.gateway?.environment?.BIFROST_UPSTREAM_GITLAB,
    /^http:\/\/mock-gitlab:8080$/,
  );
  assert.equal(config.networks?.["bifrost-private"]?.internal, true);
  assert.equal(config.networks?.["bifrost-public"]?.internal ?? false, false);
  assert.deepEqual(Object.keys(services.gateway?.networks ?? {}).sort(), [
    "bifrost-private",
    "bifrost-public",
  ]);

  for (const serviceName of [
    "postgres",
    "mock-gitlab",
    "mock-jenkins",
    "mock-docs",
    "mock-internal-admin",
  ]) {
    assert.deepEqual(Object.keys(services[serviceName]?.networks ?? {}), ["bifrost-private"]);
  }
});

test("dev deploy compose avoids parallel build image tag conflicts", () => {
  const config = loadDevDeployComposeConfig();
  const services = config.services ?? {};
  const buildServiceNames = [
    "gateway",
    "mock-gitlab",
    "mock-jenkins",
    "mock-docs",
    "mock-internal-admin",
  ];
  const buildImages = buildServiceNames.map((serviceName) => services[serviceName]?.image);

  assert.equal(services.migrate?.build, undefined, "migrate must reuse the gateway image");
  assert.equal(services.migrate?.image, services.gateway?.image);
  assert.equal(new Set(buildImages).size, buildImages.length);
});

test("dev deploy script runs migrations seeds and private exposure checks", () => {
  const deployScriptPath = join(repositoryRoot, "deploy/dev/deploy.sh");
  const dockerfile = readFileSync(join(repositoryRoot, "apps/gateway/Dockerfile"), "utf8");

  assert.ok(existsSync(deployScriptPath), "dev deploy script is required");

  const deployScript = readFileSync(deployScriptPath, "utf8");
  assert.match(dockerfile, /bifrost-migrate/);
  assert.match(deployScript, /BIFROST_DEPLOY_STATE_DIR/);
  assert.match(deployScript, /docker compose/);
  assert.doesNotMatch(deployScript, /compose build gateway migrate/);
  assert.match(deployScript, /upsert_env_key "BIFROST_DEV_GATEWAY_PORT"/);
  assert.match(deployScript, /run --rm migrate up/);
  assert.match(deployScript, /run --rm migrate seed/);
  assert.match(deployScript, /assert_private_service/);
  assert.match(deployScript, /docker inspect/);
  assert.match(deployScript, /HostPort/);
  assert.doesNotMatch(deployScript, /compose port/);
});

test("AGENTS documents branch first dev deployment workflow", () => {
  const agents = readFileSync(join(repositoryRoot, "AGENTS.md"), "utf8");

  assert.match(agents, /开发新的功能必须新建分支/);
  assert.match(agents, /测试没有任何问题后合并到 dev 分支/);
  assert.match(agents, /线上测试/);
  assert.match(agents, /GitHub Action 执行完成/);
});
