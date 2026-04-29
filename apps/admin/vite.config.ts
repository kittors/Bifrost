import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { defineConfig, loadEnv } from "vite";

const repositoryRoot = resolve(dirname(fileURLToPath(import.meta.url)), "../..");
const defaultAdminDevPort = 5173;

function parsePort(value: string | undefined, fallback: number) {
  const parsed = Number.parseInt(value ?? "", 10);

  if (Number.isInteger(parsed) && parsed > 0 && parsed <= 65535) {
    return parsed;
  }

  return fallback;
}

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, repositoryRoot, "");
  const adminDevPort = parsePort(
    env.BIFROST_ADMIN_DEV_PORT ?? env.BIFROST_DEV_ADMIN_PORT,
    defaultAdminDevPort,
  );

  return {
    envDir: repositoryRoot,
    plugins: [tailwindcss(), react()],
    server: {
      host: true,
      port: adminDevPort,
      strictPort: true,
    },
  };
});
