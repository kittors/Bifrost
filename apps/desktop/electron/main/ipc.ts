import { app, ipcMain, shell } from "electron";

import { desktopIPC } from "../shared/ipc";
import { attachDeviceID, ensureDeviceIdentity, signDeviceChallenge } from "./device-identity";
import { getDiagnosticsSnapshot } from "./diagnostics";
import {
  clearDeviceIdentity,
  clearSessionSnapshot,
  loadDeviceIdentity,
  loadSessionSnapshot,
  saveSessionSnapshot,
} from "./security-store";

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
  ipcMain.handle(desktopIPC.openExternal, (_event, url: string) =>
    shell.openExternal(assertHTTPURL(url)),
  );
  ipcMain.handle(desktopIPC.diagnosticsSnapshot, () => getDiagnosticsSnapshot());
}
