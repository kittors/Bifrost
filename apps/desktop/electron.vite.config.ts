import { resolve } from "node:path";
import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { defineConfig, externalizeDepsPlugin } from "electron-vite";

const defaultDesktopDevPort = 22473;
const desktopDevPort = Number.parseInt(
  process.env.BIFROST_DESKTOP_DEV_PORT ?? `${defaultDesktopDevPort}`,
  10,
);

export default defineConfig({
  main: {
    plugins: [externalizeDepsPlugin()],
    build: {
      lib: {
        entry: resolve("electron/main/index.ts"),
      },
    },
  },
  preload: {
    plugins: [externalizeDepsPlugin()],
    build: {
      lib: {
        entry: resolve("electron/preload/index.ts"),
      },
    },
  },
  renderer: {
    root: resolve("renderer"),
    server: {
      host: "127.0.0.1",
      port: desktopDevPort,
      strictPort: true,
    },
    build: {
      outDir: resolve("out/renderer"),
      rollupOptions: {
        input: {
          index: resolve("renderer/index.html"),
        },
      },
    },
    plugins: [tailwindcss(), react()],
  },
});
