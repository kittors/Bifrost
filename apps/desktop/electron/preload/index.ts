import { contextBridge, ipcRenderer } from "electron";

import { desktopIPC } from "../shared/ipc";
import type { DesktopSessionSnapshot } from "../shared/types";

// Preload 只暴露受控 API，不把 ipcRenderer 或任意 channel 暴露给 Renderer。
contextBridge.exposeInMainWorld("bifrostDesktop", {
  app: {
    getInfo: () => ipcRenderer.invoke(desktopIPC.appInfo),
  },
  device: {
    attach: (deviceID: string) => ipcRenderer.invoke(desktopIPC.deviceAttach, deviceID),
    clear: () => ipcRenderer.invoke(desktopIPC.deviceClear),
    ensure: () => ipcRenderer.invoke(desktopIPC.deviceEnsure),
    load: () => ipcRenderer.invoke(desktopIPC.deviceLoad),
    signChallenge: (challenge: string) =>
      ipcRenderer.invoke(desktopIPC.deviceSignChallenge, challenge),
  },
  diagnostics: {
    snapshot: () => ipcRenderer.invoke(desktopIPC.diagnosticsSnapshot),
  },
  openExternal: (url: string) => ipcRenderer.invoke(desktopIPC.openExternal, url),
  session: {
    clear: () => ipcRenderer.invoke(desktopIPC.sessionClear),
    load: () => ipcRenderer.invoke(desktopIPC.sessionLoad),
    save: (session: DesktopSessionSnapshot) => ipcRenderer.invoke(desktopIPC.sessionSave, session),
  },
});
