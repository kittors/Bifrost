import { expect, test } from "@playwright/test";

import {
  adminLogin,
  bootstrapClientDevice,
  listAdminAuditEvents,
  proxyServiceRequest,
} from "./fixtures/client-api";
import { gatewayBaseURL, seedPassword } from "./fixtures/env";

test("slow upstream returns a gateway timeout instead of hanging the client", async ({
  request,
}) => {
  const client = await bootstrapClientDevice(request, "alice", seedPassword);
  const failed = await proxyServiceRequest(
    request,
    client.session.accessToken,
    "gitlab",
    "slow?delayMs=7000",
  );

  expect(failed.response.status()).toBe(504);
  expect(failed.payload.error?.code).toBe("GATEWAY_UPSTREAM_TIMEOUT");
});

test("admin audit queries can find login success, login failure, access success and access denial", async ({
  request,
}) => {
  const admin = await adminLogin(request, "admin", seedPassword);
  const client = await bootstrapClientDevice(request, "alice", seedPassword);

  await request.post(`${gatewayBaseURL}/api/v1/admin/auth/login`, {
    data: {
      password: "WrongPassword!",
      username: "admin",
    },
  });
  await proxyServiceRequest(request, client.session.accessToken, "gitlab");
  await proxyServiceRequest(request, client.session.accessToken, "jenkins");

  const loginSuccessEvents = await listAdminAuditEvents(request, admin.session.accessToken, {
    result: "success",
    type: "auth.login.succeeded",
  });
  const loginFailedEvents = await listAdminAuditEvents(request, admin.session.accessToken, {
    result: "failure",
    type: "auth.login.failed",
  });
  const accessGrantedEvents = await listAdminAuditEvents(request, admin.session.accessToken, {
    result: "success",
    type: "service.access.granted",
  });
  const accessDeniedEvents = await listAdminAuditEvents(request, admin.session.accessToken, {
    result: "failure",
    type: "service.access.denied",
  });

  expect(loginSuccessEvents.length).toBeGreaterThan(0);
  expect(loginFailedEvents.length).toBeGreaterThan(0);
  expect(accessGrantedEvents.length).toBeGreaterThan(0);
  expect(accessDeniedEvents.length).toBeGreaterThan(0);
});
