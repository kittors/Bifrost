import { app, BrowserWindow } from "electron";

import { registerDesktopIPC } from "./ipc";
import { createMainWindow } from "./window";

// 主进程只做系统边界内的工作：窗口、IPC、安全存储和外部浏览器打开。
registerDesktopIPC();

app.whenReady().then(() => {
  void createMainWindow();

  app.on("activate", () => {
    if (BrowserWindow.getAllWindows().length === 0) {
      void createMainWindow();
    }
  });
});

app.on("window-all-closed", () => {
  if (process.platform !== "darwin") {
    app.quit();
  }
});
