import type {
  DesktopAppInfo,
  DesktopDeviceIdentity,
  DesktopDiagnosticsSnapshot,
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
      openExternal: (url: string) => Promise<void>;
      session: {
        clear: () => Promise<void>;
        load: () => Promise<DesktopSessionSnapshot | null>;
        save: (session: DesktopSessionSnapshot) => Promise<void>;
      };
    };
  }
}
