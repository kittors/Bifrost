import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { defineConfig, externalizeDepsPlugin, loadEnv } from "electron-vite";

const appRoot = dirname(fileURLToPath(import.meta.url));
const repositoryRoot = resolve(appRoot, "../..");
const defaultDesktopDevPort = 22473;

function parsePort(value: string | undefined, fallback: number) {
  const parsed = Number.parseInt(value ?? "", 10);

  if (Number.isInteger(parsed) && parsed > 0 && parsed <= 65535) {
    return parsed;
  }

  return fallback;
}

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, repositoryRoot, "");
  const desktopDevPort = parsePort(env.BIFROST_DESKTOP_DEV_PORT, defaultDesktopDevPort);

  return {
    main: {
      plugins: [externalizeDepsPlugin()],
      build: {
        lib: {
          entry: resolve(appRoot, "electron/main/index.ts"),
        },
      },
    },
    preload: {
      plugins: [externalizeDepsPlugin()],
      build: {
        lib: {
          entry: resolve(appRoot, "electron/preload/index.ts"),
        },
      },
    },
    renderer: {
      envDir: repositoryRoot,
      root: resolve(appRoot, "renderer"),
      server: {
        host: "127.0.0.1",
        port: desktopDevPort,
        strictPort: true,
      },
      build: {
        outDir: resolve(appRoot, "out/renderer"),
        rollupOptions: {
          input: {
            index: resolve(appRoot, "renderer/index.html"),
          },
        },
      },
      plugins: [tailwindcss(), react()],
    },
  };
});
