import { join } from "node:path";
import { BrowserWindow, shell } from "electron";

// 桌面窗口保持“小卡片”尺寸，并显式锁定安全 webPreferences。
export async function createMainWindow() {
  const window = new BrowserWindow({
    height: 560,
    minHeight: 480,
    minWidth: 380,
    show: false,
    title: "Bifrost",
    webPreferences: {
      contextIsolation: true,
      nodeIntegration: false,
      preload: join(__dirname, "../preload/index.js"),
      sandbox: true,
    },
    width: 420,
  });

  window.once("ready-to-show", () => {
    window.show();
  });

  window.webContents.setWindowOpenHandler(({ url }) => {
    if (url.startsWith("https://") || url.startsWith("http://")) {
      void shell.openExternal(url);
    }
    return { action: "deny" };
  });

  window.webContents.on("will-navigate", (event, url) => {
    if (!url.startsWith("file://") && !url.startsWith("http://localhost")) {
      event.preventDefault();
    }
  });

  if (process.env.ELECTRON_RENDERER_URL) {
    await window.loadURL(process.env.ELECTRON_RENDERER_URL);
    return window;
  }

  await window.loadFile(join(__dirname, "../renderer/index.html"));
  return window;
}
