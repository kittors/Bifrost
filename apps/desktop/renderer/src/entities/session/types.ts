import type {
  DesktopDeviceIdentity,
  DesktopSessionSnapshot,
} from "../../../../electron/shared/types";

export type DesktopView = "account" | "diagnostics" | "services" | "settings";

export type DesktopSessionState = {
  device: DesktopDeviceIdentity | null;
  errorMessage: string | null;
  gatewayBaseURL: string;
  isHydrating: boolean;
  session: DesktopSessionSnapshot | null;
  theme: "dark" | "light";
  view: DesktopView;
};
