import { beforeEach, describe, expect, it, vi } from "vitest";

import type { DesktopSessionSnapshot } from "../../../../electron/shared/types";
import { ApiClientError } from "../../shared/lib/http";

const sessionSnapshot: DesktopSessionSnapshot = {
  accessToken: "access_01",
  deviceId: "device_01",
  expiresAt: "2026-04-18T12:00:00.000Z",
  gatewayBaseURL: "http://127.0.0.1:8080",
  refreshToken: "refresh_01",
  user: {
    displayName: "Alice",
    id: "user_alice",
    roles: ["role_developer"],
    username: "alice",
  },
};

describe("desktop session store", () => {
  beforeEach(() => {
    vi.resetModules();
    vi.unstubAllGlobals();

    const localStorageStub = new Map<string, string>();
    const desktopWindow = {
      app: {
        getInfo: vi.fn(),
      },
      device: {
        attach: vi.fn(),
        clear: vi.fn(),
        ensure: vi.fn(),
        load: vi.fn(),
        signChallenge: vi.fn(),
      },
      diagnostics: {
        snapshot: vi.fn(),
      },
      openExternal: vi.fn(),
      session: {
        clear: vi.fn().mockResolvedValue(undefined),
        load: vi.fn().mockResolvedValue(sessionSnapshot),
        save: vi.fn().mockResolvedValue(undefined),
      },
    };

    vi.stubGlobal("window", {
      bifrostDesktop: desktopWindow,
      localStorage: {
        clear: () => localStorageStub.clear(),
        getItem: (key: string) => localStorageStub.get(key) ?? null,
        removeItem: (key: string) => localStorageStub.delete(key),
        setItem: (key: string, value: string) => localStorageStub.set(key, value),
      },
    });
  });

  it("shows userMessage when refresh fails because device is disabled", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        json: async () => ({
          success: false,
          error: {
            code: "DEVICE_DISABLED",
            message: "device disabled",
            userMessage: "当前设备已被管理员禁用，请联系管理员处理",
          },
          meta: { requestId: "req_refresh_disabled", timestamp: "2026-04-18T12:01:00Z" },
        }),
        status: 403,
      }),
    );

    const { useDesktopSessionStore } = await import("./store");

    await useDesktopSessionStore.getState().hydrateFromSecureStore();

    expect(window.bifrostDesktop.session.clear).toHaveBeenCalledTimes(1);
    expect(useDesktopSessionStore.getState().session).toBeNull();
    expect(useDesktopSessionStore.getState().isHydrating).toBe(false);
    expect(useDesktopSessionStore.getState().errorMessage).toBe(
      "当前设备已被管理员禁用，请联系管理员处理",
    );
  });

  it("preserves ApiClientError metadata for downstream UI mapping", () => {
    const error = new ApiClientError({
      code: "DEVICE_DISABLED",
      message: "device disabled",
      requestId: "req_meta",
      statusCode: 403,
      userMessage: "当前设备已被管理员禁用，请联系管理员处理",
    });

    expect(error.userMessage).toBe("当前设备已被管理员禁用，请联系管理员处理");
  });
});
