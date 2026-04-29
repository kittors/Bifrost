import { describe, expect, it, vi } from "vitest";

import { clientLogin, refreshClientSession } from "./api";

describe("desktop session api", () => {
  it("logs in with an existing trusted device", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        json: async () => ({
          success: true,
          data: {
            accessToken: "access",
            refreshToken: "refresh",
            expiresIn: 900,
            user: {
              id: "user_alice",
              username: "alice",
              displayName: "Alice",
              roles: ["role_developer"],
            },
          },
          meta: { requestId: "req_login", timestamp: "2026-04-18T04:20:00Z" },
        }),
        status: 200,
      }),
    );

    const result = await clientLogin({
      baseURL: "http://127.0.0.1:8080",
      clientVersion: "0.1.0",
      deviceId: "device_01",
      password: "ChangeMe123!",
      username: "alice",
    });

    expect(result.accessToken).toBe("access");
  });

  it("refreshes a session with refresh token and device id", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        json: async () => ({
          success: true,
          data: {
            accessToken: "access_2",
            refreshToken: "refresh_2",
            expiresIn: 900,
            user: {
              id: "user_alice",
              username: "alice",
              displayName: "Alice",
              roles: ["role_developer"],
            },
          },
          meta: { requestId: "req_refresh", timestamp: "2026-04-18T04:22:00Z" },
        }),
        status: 200,
      }),
    );

    const result = await refreshClientSession({
      baseURL: "http://127.0.0.1:8080",
      deviceId: "device_01",
      refreshToken: "refresh",
    });

    expect(result.refreshToken).toBe("refresh_2");
  });
});
