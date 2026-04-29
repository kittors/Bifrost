import { expect, test } from "@playwright/test";

import { createLocalProxyController } from "../../apps/desktop/electron/main/local-proxy";
import { bootstrapClientDevice, proxyServiceRequest } from "./fixtures/client-api";
import { gatewayBaseURL, seedPassword } from "./fixtures/env";

function buildDesktopSessionSnapshot(input: Awaited<ReturnType<typeof bootstrapClientDevice>>) {
  return {
    accessToken: input.session.accessToken,
    deviceId: input.device.deviceId,
    expiresAt: new Date(Date.now() + 15 * 60 * 1000).toISOString(),
    gatewayBaseURL,
    refreshToken: input.session.refreshToken,
    user: input.session.user,
  };
}

test("desktop local proxy restores access to an allowed private web service", async ({
  request,
}) => {
  const client = await bootstrapClientDevice(request, "alice", seedPassword);
  const controller = createLocalProxyController({
    maxPort: 18129,
    preferredPort: 18120,
  });

  try {
    const status = await controller.start(buildDesktopSessionSnapshot(client));
    const response = await fetch(`${status.baseURL}/s/gitlab/whoami`);
    const payload = (await response.json()) as {
      serviceKey: string;
      serviceName: string;
    };

    expect(response.ok).toBeTruthy();
    expect(payload.serviceKey).toBe("gitlab");
    expect(payload.serviceName).toBe("Mock GitLab");
  } finally {
    await controller.stop();
  }
});

test("desktop local proxy still respects gateway policy and denies unauthorized services", async ({
  request,
}) => {
  const client = await bootstrapClientDevice(request, "bob", seedPassword);
  const controller = createLocalProxyController({
    maxPort: 18139,
    preferredPort: 18130,
  });

  try {
    const status = await controller.start(buildDesktopSessionSnapshot(client));
    const response = await fetch(`${status.baseURL}/s/gitlab/whoami`);
    const payload = (await response.json()) as {
      error?: {
        code?: string;
      };
      success?: boolean;
    };

    expect(response.status).toBe(403);
    expect(payload.success).toBeFalsy();
    expect(payload.error?.code).toBe("POLICY_ACCESS_DENIED");
  } finally {
    await controller.stop();
  }
});

test("server-side direct proxy behavior stays consistent with the desktop loopback path", async ({
  request,
}) => {
  const client = await bootstrapClientDevice(request, "alice", seedPassword);
  const proxied = await proxyServiceRequest(request, client.session.accessToken, "gitlab");

  expect(proxied.response.ok()).toBeTruthy();
  expect(proxied.payload.serviceKey).toBe("gitlab");
});
