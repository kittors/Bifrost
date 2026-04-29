import assert from "node:assert/strict";
import { existsSync, readFileSync } from "node:fs";
import test from "node:test";

const remoteDevGatewayURL = "http://142.171.208.80:18080";

test("local backend development defaults to the remote dev gateway", () => {
  const packageJson = JSON.parse(
    readFileSync(new URL("../../package.json", import.meta.url), "utf8"),
  );
  const envExample = readFileSync(new URL("../../.env.example", import.meta.url), "utf8");

  assert.equal(packageJson.scripts["dev:backend"], "node ./scripts/dev/remote-backend.mjs");
  assert.equal(packageJson.scripts["dev:backend:local"], "node ./scripts/testing/backend-up.mjs");
  assert.equal(
    packageJson.scripts["dev:backend:local:down"],
    "node ./scripts/testing/e2e-down.mjs",
  );
  assert.match(envExample, /^BIFROST_PUBLIC_BASE_URL=http:\/\/142\.171\.208\.80:18080$/m);
});

test("backend validation still uses the explicit local docker backend", () => {
  const backendRun = readFileSync(
    new URL("../../scripts/testing/backend-run.mjs", import.meta.url),
    "utf8",
  );

  assert.match(backendRun, /\["dev:backend:local"\]/);
  assert.doesNotMatch(backendRun, /\["dev:backend"\]/);
});

test("admin and desktop fall back to the remote dev gateway", () => {
  const adminEnv = readFileSync(
    new URL("../../apps/admin/src/shared/config/env.ts", import.meta.url),
    "utf8",
  );
  const desktopEnv = readFileSync(
    new URL("../../apps/desktop/renderer/src/shared/config/env.ts", import.meta.url),
    "utf8",
  );
  const desktopStore = readFileSync(
    new URL("../../apps/desktop/renderer/src/entities/session/store.ts", import.meta.url),
    "utf8",
  );
  const loginCard = readFileSync(
    new URL("../../apps/desktop/renderer/src/features/auth/login-card.tsx", import.meta.url),
    "utf8",
  );

  assert.match(adminEnv, new RegExp(remoteDevGatewayURL.replaceAll(".", "\\.")));
  assert.match(desktopEnv, new RegExp(remoteDevGatewayURL.replaceAll(".", "\\.")));
  assert.match(desktopStore, new RegExp(remoteDevGatewayURL.replaceAll(".", "\\.")));
  assert.match(
    loginCard,
    new RegExp(`placeholder="${remoteDevGatewayURL.replaceAll(".", "\\.")}"`),
  );
});

test("remote backend helper checks gateway and private upstream routing", () => {
  const helperURL = new URL("../../scripts/dev/remote-backend.mjs", import.meta.url);

  assert.ok(existsSync(helperURL), "remote backend helper is required");

  const helper = readFileSync(helperURL, "utf8");
  assert.match(helper, new RegExp(remoteDevGatewayURL.replaceAll(".", "\\.")));
  assert.match(helper, /\/healthz/);
  assert.match(helper, /\/readyz/);
  assert.match(helper, /\/debug\/upstreams\/gitlab/);
  assert.match(helper, /VITE_GATEWAY_BASE_URL/);
});
