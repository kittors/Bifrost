import { defineConfig } from "@playwright/test";

export default defineConfig({
  fullyParallel: false,
  outputDir: "test-results/playwright",
  projects: [
    {
      name: "api-chromium",
      use: {
        browserName: "chromium",
      },
    },
  ],
  reporter: process.env.CI ? [["github"], ["html", { open: "never" }]] : "list",
  testDir: "./tests/e2e",
  timeout: 45_000,
  workers: 1,
  use: {
    baseURL: process.env.BIFROST_ADMIN_BASE_URL ?? "http://127.0.0.1:15173",
    trace: "retain-on-failure",
  },
});
