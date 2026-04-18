import { expect, test } from "@playwright/test";

import {
  adminLogin,
  bootstrapClientDevice,
  proxyServiceRequest,
  setAdminServiceStatus,
  updateAdminService,
} from "./fixtures/client-api";
import { seedPassword } from "./fixtures/env";

test("disabled service is rejected immediately by proxy authorization", async ({ request }) => {
  const admin = await adminLogin(request, "admin", seedPassword);
  await setAdminServiceStatus(request, admin.session.accessToken, "service_gitlab", "disabled");

  try {
    const client = await bootstrapClientDevice(request, "alice", seedPassword);
    const denied = await proxyServiceRequest(request, client.session.accessToken, "gitlab");

    expect(denied.response.status()).toBe(403);
    expect(denied.payload.success).toBeFalsy();
    expect(denied.payload.error?.code).toBe("SERVICE_DISABLED");
  } finally {
    await setAdminServiceStatus(request, admin.session.accessToken, "service_gitlab", "enabled");
  }
});

test("unreachable upstream returns a gateway error instead of policy denial", async ({
  request,
}) => {
  const admin = await adminLogin(request, "admin", seedPassword);
  await updateAdminService(request, admin.session.accessToken, "service_gitlab", {
    description: "Mock GitLab upstream for local development",
    group: "engineering",
    name: "GitLab",
    protocol: "http",
    publicPath: "/s/gitlab",
    upstreamUrl: "http://127.0.0.1:9",
  });

  const client = await bootstrapClientDevice(request, "alice", seedPassword);

  try {
    const failed = await proxyServiceRequest(request, client.session.accessToken, "gitlab");

    expect(failed.response.status()).toBe(502);
    expect(failed.payload.success).toBeFalsy();
    expect(failed.payload.error?.code).toBe("GATEWAY_BAD_UPSTREAM");
  } finally {
    await updateAdminService(request, admin.session.accessToken, "service_gitlab", {
      description: "Mock GitLab upstream for local development",
      group: "engineering",
      name: "GitLab",
      protocol: "http",
      publicPath: "/s/gitlab",
      upstreamUrl: "http://mock-gitlab:8080",
    });
  }
});
