import { afterEach, describe, expect, it, vi } from "vitest";

import { getGatewayBaseURL } from "./env";

describe("admin env helpers", () => {
  afterEach(() => {
    vi.unstubAllEnvs();
  });

  it("uses the same-origin Vite proxy in local dev even when a remote gateway URL is configured", () => {
    vi.stubEnv("VITE_GATEWAY_BASE_URL", "http://142.171.208.80:18080");

    expect(getGatewayBaseURL()).toBe("/__bifrost_gateway__");
  });
});
