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
      localProxy: {
        openService: vi.fn(),
        start: vi.fn().mockResolvedValue({
          baseURL: "http://127.0.0.1:18080",
          host: "127.0.0.1",
          port: 18080,
          running: true,
        }),
        status: vi.fn().mockResolvedValue({
          baseURL: "http://127.0.0.1:18080",
          host: "127.0.0.1",
          port: 18080,
          running: true,
        }),
        stop: vi.fn().mockResolvedValue(undefined),
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

  it("starts the local proxy after saving a desktop session", async () => {
    const { useDesktopSessionStore } = await import("./store");

    await useDesktopSessionStore.getState().saveSession(sessionSnapshot);

    expect(window.bifrostDesktop.session.save).toHaveBeenCalledWith(sessionSnapshot);
    expect(window.bifrostDesktop.localProxy.start).toHaveBeenCalledWith(sessionSnapshot);
    expect(useDesktopSessionStore.getState().localProxyStatus).toEqual({
      baseURL: "http://127.0.0.1:18080",
      host: "127.0.0.1",
      port: 18080,
      running: true,
    });
  });

  it("starts the local proxy after hydrating and refreshing a stored session", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        json: async () => ({
          success: true,
          data: {
            accessToken: "access_refreshed",
            expiresIn: 900,
            refreshToken: "refresh_refreshed",
            user: sessionSnapshot.user,
          },
          meta: { requestId: "req_refresh_ok", timestamp: "2026-04-18T12:01:00Z" },
        }),
        status: 200,
      }),
    );

    const { useDesktopSessionStore } = await import("./store");

    await useDesktopSessionStore.getState().hydrateFromSecureStore();

    expect(window.bifrostDesktop.localProxy.start).toHaveBeenCalledWith(
      expect.objectContaining({
        accessToken: "access_refreshed",
        refreshToken: "refresh_refreshed",
      }),
    );
    expect(useDesktopSessionStore.getState().localProxyStatus?.running).toBe(true);
  });

  it("stops the local proxy when clearing the desktop session", async () => {
    const { useDesktopSessionStore } = await import("./store");

    await useDesktopSessionStore.getState().saveSession(sessionSnapshot);
    await useDesktopSessionStore.getState().clearSession();

    expect(window.bifrostDesktop.localProxy.stop).toHaveBeenCalledTimes(1);
    expect(window.bifrostDesktop.session.clear).toHaveBeenCalledTimes(1);
    expect(useDesktopSessionStore.getState().session).toBeNull();
    expect(useDesktopSessionStore.getState().localProxyStatus?.running).toBe(false);
  });

  it("refreshes an active session and rotates the local proxy token snapshot", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        json: async () => ({
          success: true,
          data: {
            accessToken: "access_rotated",
            expiresIn: 900,
            refreshToken: "refresh_rotated",
            user: sessionSnapshot.user,
          },
          meta: { requestId: "req_refresh_active", timestamp: "2026-04-18T12:01:00Z" },
        }),
        status: 200,
      }),
    );
    const { useDesktopSessionStore } = await import("./store");

    await useDesktopSessionStore.getState().saveSession(sessionSnapshot);
    await useDesktopSessionStore.getState().refreshActiveSession();

    expect(window.bifrostDesktop.session.save).toHaveBeenLastCalledWith(
      expect.objectContaining({
        accessToken: "access_rotated",
        refreshToken: "refresh_rotated",
      }),
    );
    expect(window.bifrostDesktop.localProxy.start).toHaveBeenLastCalledWith(
      expect.objectContaining({
        accessToken: "access_rotated",
        refreshToken: "refresh_rotated",
      }),
    );
  });

  it("keeps the session but exposes an error when local proxy startup fails", async () => {
    vi.mocked(window.bifrostDesktop.localProxy.start).mockRejectedValueOnce(
      new Error("no available local proxy port"),
    );
    const { useDesktopSessionStore } = await import("./store");

    await useDesktopSessionStore.getState().saveSession(sessionSnapshot);

    expect(useDesktopSessionStore.getState().session).toEqual(sessionSnapshot);
    expect(useDesktopSessionStore.getState().errorMessage).toBe("no available local proxy port");
    expect(useDesktopSessionStore.getState().localProxyStatus?.running).toBe(false);
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
