import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { defineConfig, loadEnv } from "vite";

const repositoryRoot = resolve(dirname(fileURLToPath(import.meta.url)), "../..");
const defaultAdminDevPort = 5173;
const defaultGatewayDevProxyBaseURL = "/__bifrost_gateway__";
const fallbackGatewayProxyTarget = "http://142.171.208.80:18080";

function parsePort(value: string | undefined, fallback: number) {
  const parsed = Number.parseInt(value ?? "", 10);

  if (Number.isInteger(parsed) && parsed > 0 && parsed <= 65535) {
    return parsed;
  }

  return fallback;
}

function escapeRegExp(value: string) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

function normalizeBaseURL(value: string) {
  return value.replace(/\/+$/, "");
}

function normalizeProxyBaseURL(value: string | undefined) {
  const trimmed = value?.trim();
  if (!trimmed) {
    return defaultGatewayDevProxyBaseURL;
  }

  return normalizeBaseURL(trimmed.startsWith("/") ? trimmed : `/${trimmed}`);
}

function isHTTPURL(value: string | undefined) {
  return /^https?:\/\//.test(value?.trim() ?? "");
}

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, repositoryRoot, "");
  const adminDevPort = parsePort(
    env.BIFROST_ADMIN_DEV_PORT ?? env.BIFROST_DEV_ADMIN_PORT,
    defaultAdminDevPort,
  );
  const gatewayDevProxyBaseURL = normalizeProxyBaseURL(env.VITE_GATEWAY_DEV_PROXY_BASE_URL);
  const gatewayProxyTarget = normalizeBaseURL(
    env.BIFROST_REMOTE_DEV_GATEWAY_URL?.trim() ||
      (isHTTPURL(env.VITE_GATEWAY_BASE_URL) ? env.VITE_GATEWAY_BASE_URL.trim() : "") ||
      fallbackGatewayProxyTarget,
  );
  const gatewayProxyBasePattern = new RegExp(`^${escapeRegExp(gatewayDevProxyBaseURL)}`);

  return {
    envDir: repositoryRoot,
    plugins: [tailwindcss(), react()],
    server: {
      host: true,
      port: adminDevPort,
      proxy: {
        [gatewayDevProxyBaseURL]: {
          changeOrigin: true,
          rewrite: (path) => path.replace(gatewayProxyBasePattern, ""),
          target: gatewayProxyTarget,
        },
      },
      strictPort: true,
    },
  };
});
