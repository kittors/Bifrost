import { reactVitestConfig } from "@bifrost/config-vitest/react";
import { defineConfig, mergeConfig } from "vitest/config";

export default mergeConfig(
  reactVitestConfig,
  defineConfig({
    test: {
      include: ["src/**/*.{test,spec}.{ts,tsx}"],
    },
  }),
);
