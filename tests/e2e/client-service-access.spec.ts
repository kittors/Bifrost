import { expect, test } from "@playwright/test";

import {
  bootstrapClientDevice,
  createServiceAccessURL,
  listClientServices,
} from "./fixtures/client-api";
import { seedPassword } from "./fixtures/env";

test("alice can bootstrap a device and only see her authorized services", async ({ request }) => {
  const session = await bootstrapClientDevice(request, "alice", seedPassword);
  const services = await listClientServices(request, session.session.accessToken);

  expect(services.map((service) => service.key)).toEqual(["docs", "gitlab"]);
});

test("alice can request gitlab access but cannot request jenkins access", async ({ request }) => {
  const session = await bootstrapClientDevice(request, "alice", seedPassword);
  const services = await listClientServices(request, session.session.accessToken);
  const gitlab = services.find((service) => service.key === "gitlab");

  expect(gitlab).toBeDefined();

  const allowed = await createServiceAccessURL(
    request,
    session.session.accessToken,
    gitlab?.id ?? "",
  );
  expect(allowed.response.ok()).toBeTruthy();
  expect(allowed.payload.success).toBeTruthy();
  expect(allowed.payload.data.url).toContain("/s/gitlab");

  const gitlabWhoamiURL = new URL(
    "whoami",
    allowed.payload.data.url.endsWith("/")
      ? allowed.payload.data.url
      : `${allowed.payload.data.url}/`,
  );
  const upstreamResponse = await request.get(gitlabWhoamiURL.toString());
  const upstreamPayload = (await upstreamResponse.json()) as {
    serviceKey: string;
    serviceName: string;
  };
  expect(upstreamResponse.ok()).toBeTruthy();
  expect(upstreamPayload.serviceKey).toBe("gitlab");
  expect(upstreamPayload.serviceName).toBe("Mock GitLab");

  const denied = await createServiceAccessURL(
    request,
    session.session.accessToken,
    "service_jenkins",
  );
  expect(denied.response.status()).toBe(403);
  expect(denied.payload.success).toBeFalsy();
  expect(denied.payload.error?.code).toBe("POLICY_ACCESS_DENIED");
});
