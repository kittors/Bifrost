import { app, ipcMain, shell } from "electron";

import { desktopIPC } from "../shared/ipc";
import type { DesktopSessionSnapshot } from "../shared/types";
import { attachDeviceID, ensureDeviceIdentity, signDeviceChallenge } from "./device-identity";
import { getDiagnosticsSnapshot } from "./diagnostics";
import { createLocalProxyController } from "./local-proxy";
import {
  clearDeviceIdentity,
  clearSessionSnapshot,
  loadDeviceIdentity,
  loadSessionSnapshot,
  saveSessionSnapshot,
} from "./security-store";

const localProxyController = createLocalProxyController();

function assertHTTPURL(url: string) {
  const parsed = new URL(url);
  if (parsed.protocol !== "https:" && parsed.protocol !== "http:") {
    throw new Error("only http and https URLs can be opened");
  }
  return parsed.toString();
}

// IPC 必须逐个白名单注册，不暴露任意 channel 调用能力。
export function registerDesktopIPC() {
  ipcMain.handle(desktopIPC.appInfo, () => ({
    name: "Bifrost Desktop",
    platform: process.platform,
    version: app.getVersion(),
  }));
  ipcMain.handle(desktopIPC.sessionLoad, () => loadSessionSnapshot());
  ipcMain.handle(desktopIPC.sessionSave, (_event, session) => saveSessionSnapshot(session));
  ipcMain.handle(desktopIPC.sessionClear, () => clearSessionSnapshot());
  ipcMain.handle(desktopIPC.deviceAttach, (_event, deviceID: string) => attachDeviceID(deviceID));
  ipcMain.handle(desktopIPC.deviceLoad, () => loadDeviceIdentity());
  ipcMain.handle(desktopIPC.deviceEnsure, () => ensureDeviceIdentity());
  ipcMain.handle(desktopIPC.deviceClear, () => clearDeviceIdentity());
  ipcMain.handle(desktopIPC.deviceSignChallenge, (_event, challenge: string) =>
    signDeviceChallenge(challenge),
  );
  ipcMain.handle(desktopIPC.localProxyStart, (_event, session: DesktopSessionSnapshot) =>
    localProxyController.start(session),
  );
  ipcMain.handle(desktopIPC.localProxyStatus, () => localProxyController.status());
  ipcMain.handle(desktopIPC.localProxyStop, () => localProxyController.stop());
  ipcMain.handle(desktopIPC.localProxyOpenService, async (_event, publicPath: string) => {
    const localURL = await localProxyController.openService(publicPath);
    await shell.openExternal(localURL);
    return localURL;
  });
  ipcMain.handle(desktopIPC.openExternal, (_event, url: string) =>
    shell.openExternal(assertHTTPURL(url)),
  );
  ipcMain.handle(desktopIPC.diagnosticsSnapshot, () => getDiagnosticsSnapshot());
}
