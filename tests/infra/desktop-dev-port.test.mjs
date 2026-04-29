import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import test from "node:test";

const desktopConfig = readFileSync(
  new URL("../../apps/desktop/electron.vite.config.ts", import.meta.url),
  "utf8",
);

test("desktop renderer dev server uses the reserved local development port", () => {
  assert.match(desktopConfig, /defaultDesktopDevPort\s*=\s*22473/);
  assert.match(desktopConfig, /BIFROST_DESKTOP_DEV_PORT/);
  assert.match(desktopConfig, /host:\s*"127\.0\.0\.1"/);
  assert.match(desktopConfig, /port:\s*desktopDevPort/);
  assert.match(desktopConfig, /strictPort:\s*true/);
});
