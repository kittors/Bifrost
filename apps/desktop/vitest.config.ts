import { reactVitestConfig } from "@bifrost/config-vitest/react";
import { defineConfig, mergeConfig } from "vitest/config";

export default mergeConfig(
  reactVitestConfig,
  defineConfig({
    test: {
      include: ["renderer/src/**/*.{test,spec}.{ts,tsx}"],
    },
  }),
);
