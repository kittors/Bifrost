import type {
  DesktopAppInfo,
  DesktopDeviceIdentity,
  DesktopDiagnosticsSnapshot,
  DesktopLocalProxyStatus,
  DesktopSessionSnapshot,
} from "../../electron/shared/types";

declare global {
  interface Window {
    bifrostDesktop: {
      app: {
        getInfo: () => Promise<DesktopAppInfo>;
      };
      device: {
        attach: (deviceID: string) => Promise<DesktopDeviceIdentity>;
        clear: () => Promise<void>;
        ensure: () => Promise<DesktopDeviceIdentity>;
        load: () => Promise<DesktopDeviceIdentity | null>;
        signChallenge: (challenge: string) => Promise<string>;
      };
      diagnostics: {
        snapshot: () => Promise<DesktopDiagnosticsSnapshot>;
      };
      localProxy: {
        openService: (publicPath: string) => Promise<string>;
        start: (session: DesktopSessionSnapshot) => Promise<DesktopLocalProxyStatus>;
        status: () => Promise<DesktopLocalProxyStatus>;
        stop: () => Promise<void>;
      };
      openExternal: (url: string) => Promise<void>;
      session: {
        clear: () => Promise<void>;
        load: () => Promise<DesktopSessionSnapshot | null>;
        save: (session: DesktopSessionSnapshot) => Promise<void>;
      };
    };
  }
}
