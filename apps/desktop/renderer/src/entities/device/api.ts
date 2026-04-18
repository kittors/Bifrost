import { requestJSON } from "../../shared/lib/http";
import type { DesktopBootstrapPayload } from "./types";

export async function loadLocalDeviceIdentity() {
  return window.bifrostDesktop.device.load();
}

export async function ensureLocalDeviceIdentity() {
  return window.bifrostDesktop.device.ensure();
}

export async function attachLocalDeviceID(deviceID: string) {
  return window.bifrostDesktop.device.attach(deviceID);
}

export async function bootstrapClientDevice(baseURL: string, input: DesktopBootstrapPayload) {
  return requestJSON<{
    accessToken: string;
    refreshToken: string;
    expiresIn: number;
    user: {
      displayName: string;
      id: string;
      roles: string[];
      username: string;
    };
    device: {
      deviceId: string;
      status: string;
    };
  }>({
    baseURL,
    body: input,
    method: "POST",
    path: "/api/v1/client/devices/bootstrap",
  });
}
