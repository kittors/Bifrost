import { describe, expect, it, vi } from "vitest";

import { bootstrapClientDevice } from "./api";

describe("desktop device api", () => {
  it("posts bootstrap payload to the gateway", async () => {
    const fetchMock = vi.fn().mockResolvedValue({
      json: async () => ({
        success: true,
        data: {
          accessToken: "token",
          refreshToken: "refresh",
          expiresIn: 900,
          user: {
            id: "user_alice",
            username: "alice",
            displayName: "Alice",
            roles: ["role_developer"],
          },
          device: { deviceId: "device_01", status: "trusted" },
        },
        meta: { requestId: "req_01", timestamp: "2026-04-18T04:00:00Z" },
      }),
      status: 200,
    });
    vi.stubGlobal("fetch", fetchMock);

    const result = await bootstrapClientDevice("http://127.0.0.1:8080", {
      clientVersion: "0.1.0",
      deviceName: "Alice MacBook Pro",
      deviceOs: "macOS",
      password: "ChangeMe123!",
      publicKey: "public-key",
      publicKeyFingerprint: "fp_public_key_01",
      username: "alice",
    });

    expect(result.device.deviceId).toBe("device_01");
    const [url, init] = fetchMock.mock.calls[0] ?? [];
    expect(String(url)).toBe("http://127.0.0.1:8080/api/v1/client/devices/bootstrap");
    expect(init).toMatchObject({ method: "POST" });
  });
});
