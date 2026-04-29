import { fileURLToPath } from "node:url";

import { defineConfig } from "vitest/config";

const setupFile = fileURLToPath(new URL("./setup.ts", import.meta.url));

export const reactVitestConfig = defineConfig({
  test: {
    clearMocks: true,
    css: true,
    environment: "happy-dom",
    globals: true,
    restoreMocks: true,
    setupFiles: [setupFile],
  },
});
