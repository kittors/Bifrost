import { create } from "zustand";

import type {
  DesktopDeviceIdentity,
  DesktopLocalProxyStatus,
  DesktopSessionSnapshot,
} from "../../../../electron/shared/types";
import { normalizeGatewayBaseURL } from "../../shared/config/env";
import { resolveApiErrorMessage } from "../../shared/lib/http";
import { attachLocalDeviceID } from "../device/api";
import { refreshClientSession } from "./api";
import type { DesktopSessionState, DesktopView } from "./types";

type DesktopSessionActions = {
  clearSession: () => Promise<void>;
  hydrateFromSecureStore: () => Promise<void>;
  refreshActiveSession: () => Promise<void>;
  saveSession: (session: DesktopSessionSnapshot) => Promise<void>;
  setDevice: (device: DesktopDeviceIdentity | null) => void;
  setErrorMessage: (value: string | null) => void;
  setGatewayBaseURL: (value: string) => void;
  setTheme: (value: "dark" | "light") => void;
  setView: (view: DesktopView) => void;
  updateDeviceID: (deviceID: string) => Promise<void>;
};

type DesktopStore = DesktopSessionState & DesktopSessionActions;

const stoppedLocalProxyStatus: DesktopLocalProxyStatus = {
  baseURL: "",
  host: "127.0.0.1",
  port: 0,
  running: false,
};

async function startLocalProxy(
  session: DesktopSessionSnapshot,
  set: (value: Partial<DesktopStore>) => void,
) {
  try {
    const status = await window.bifrostDesktop.localProxy.start(session);
    set({ localProxyStatus: status });
  } catch (error) {
    set({
      errorMessage: resolveApiErrorMessage(error, "本地代理启动失败"),
      localProxyStatus: stoppedLocalProxyStatus,
    });
  }
}

function readTheme() {
  return window.localStorage.getItem("bifrost.desktop.theme") === "dark" ? "dark" : "light";
}

export const useDesktopSessionStore = create<DesktopStore>((set) => ({
  device: null,
  errorMessage: null,
  gatewayBaseURL: "http://142.171.208.80:18080",
  isHydrating: true,
  localProxyStatus: stoppedLocalProxyStatus,
  session: null,
  theme: readTheme(),
  view: "services",
  clearSession: async () => {
    await window.bifrostDesktop.localProxy.stop();
    await window.bifrostDesktop.session.clear();
    set({ localProxyStatus: stoppedLocalProxyStatus, session: null, view: "services" });
  },
  hydrateFromSecureStore: async () => {
    const session = await window.bifrostDesktop.session.load();
    if (!session) {
      set({ isHydrating: false, localProxyStatus: stoppedLocalProxyStatus });
      return;
    }

    try {
      const refreshed = await refreshClientSession({
        baseURL: session.gatewayBaseURL,
        deviceId: session.deviceId,
        refreshToken: session.refreshToken,
      });

      const nextSession: DesktopSessionSnapshot = {
        accessToken: refreshed.accessToken,
        deviceId: session.deviceId,
        expiresAt: new Date(Date.now() + refreshed.expiresIn * 1000).toISOString(),
        gatewayBaseURL: session.gatewayBaseURL,
        refreshToken: refreshed.refreshToken,
        user: refreshed.user,
      };
      await window.bifrostDesktop.session.save(nextSession);
      set({ gatewayBaseURL: session.gatewayBaseURL, isHydrating: false, session: nextSession });
      await startLocalProxy(nextSession, set);
    } catch (error) {
      await window.bifrostDesktop.localProxy.stop();
      await window.bifrostDesktop.session.clear();
      set({
        errorMessage: resolveApiErrorMessage(error, "登录状态已失效，请重新登录"),
        isHydrating: false,
        localProxyStatus: stoppedLocalProxyStatus,
        session: null,
      });
    }
  },
  refreshActiveSession: async () => {
    const currentSession = useDesktopSessionStore.getState().session;
    if (!currentSession) {
      return;
    }

    try {
      const refreshed = await refreshClientSession({
        baseURL: currentSession.gatewayBaseURL,
        deviceId: currentSession.deviceId,
        refreshToken: currentSession.refreshToken,
      });

      const nextSession: DesktopSessionSnapshot = {
        accessToken: refreshed.accessToken,
        deviceId: currentSession.deviceId,
        expiresAt: new Date(Date.now() + refreshed.expiresIn * 1000).toISOString(),
        gatewayBaseURL: currentSession.gatewayBaseURL,
        refreshToken: refreshed.refreshToken,
        user: refreshed.user,
      };
      await window.bifrostDesktop.session.save(nextSession);
      set({
        errorMessage: null,
        gatewayBaseURL: nextSession.gatewayBaseURL,
        session: nextSession,
      });
      await startLocalProxy(nextSession, set);
    } catch (error) {
      await window.bifrostDesktop.localProxy.stop();
      await window.bifrostDesktop.session.clear();
      set({
        errorMessage: resolveApiErrorMessage(error, "登录状态已失效，请重新登录"),
        localProxyStatus: stoppedLocalProxyStatus,
        session: null,
      });
    }
  },
  saveSession: async (session) => {
    await window.bifrostDesktop.session.save(session);
    set({ gatewayBaseURL: session.gatewayBaseURL, session });
    await startLocalProxy(session, set);
  },
  setDevice: (device) => set({ device }),
  setErrorMessage: (errorMessage) => set({ errorMessage }),
  setGatewayBaseURL: (gatewayBaseURL) =>
    set({ gatewayBaseURL: normalizeGatewayBaseURL(gatewayBaseURL) }),
  setTheme: (theme) => {
    window.localStorage.setItem("bifrost.desktop.theme", theme);
    set({ theme });
  },
  setView: (view) => set({ view }),
  updateDeviceID: async (deviceID) => {
    const nextDevice = await attachLocalDeviceID(deviceID);
    set({ device: nextDevice });
  },
}));
