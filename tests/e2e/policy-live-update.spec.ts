import { expect, test } from "@playwright/test";

import {
  adminLogin,
  bootstrapClientDevice,
  listAdminDevices,
  listClientServices,
  proxyServiceRequest,
  replaceUserServiceOverrides,
  setAdminDeviceStatus,
} from "./fixtures/client-api";
import { seedPassword } from "./fixtures/env";

test("user-level deny takes effect immediately for service visibility and proxy access", async ({
  request,
}) => {
  const admin = await adminLogin(request, "admin", seedPassword);
  const client = await bootstrapClientDevice(request, "alice", seedPassword);

  try {
    await replaceUserServiceOverrides(request, admin.session.accessToken, client.session.user.id, {
      allowServiceIds: [],
      denyServiceIds: ["service_gitlab"],
    });

    const services = await listClientServices(request, client.session.accessToken);
    expect(services.map((service) => service.key)).toEqual(["docs"]);

    const denied = await proxyServiceRequest(request, client.session.accessToken, "gitlab");
    expect(denied.response.status()).toBe(403);
    expect(denied.payload.error?.code).toBe("POLICY_ACCESS_DENIED");
  } finally {
    await replaceUserServiceOverrides(request, admin.session.accessToken, client.session.user.id, {
      allowServiceIds: [],
      denyServiceIds: [],
    });
  }
});

test("device disable takes effect immediately for the active client session", async ({
  request,
}) => {
  const admin = await adminLogin(request, "admin", seedPassword);
  const client = await bootstrapClientDevice(request, "alice", seedPassword);
  const devices = await listAdminDevices(
    request,
    admin.session.accessToken,
    client.session.user.id,
  );
  const createdDevice = devices.find((device) => device.id === client.device.deviceId);

  expect(createdDevice).toBeDefined();

  try {
    await setAdminDeviceStatus(
      request,
      admin.session.accessToken,
      createdDevice?.id ?? client.device.deviceId,
      "disabled",
    );

    const denied = await proxyServiceRequest(request, client.session.accessToken, "gitlab");
    expect(denied.response.status()).toBe(401);
    expect(denied.payload.error?.code).toBe("AUTH_SESSION_REVOKED");
  } finally {
    await setAdminDeviceStatus(
      request,
      admin.session.accessToken,
      createdDevice?.id ?? client.device.deviceId,
      "trusted",
    );
  }
});
