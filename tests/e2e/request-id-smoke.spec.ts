import { expect, test } from "@playwright/test";

import { adminLogin } from "./fixtures/client-api";
import { gatewayBaseURL, seedPassword } from "./fixtures/env";

test("admin list endpoint returns x-request-id header and meta.requestId", async ({ request }) => {
  const { session } = await adminLogin(request, "admin", seedPassword);
  const response = await request.get(`${gatewayBaseURL}/api/v1/admin/users`, {
    headers: {
      Authorization: `Bearer ${session.accessToken}`,
    },
  });
  const payload = (await response.json()) as {
    meta: { requestId: string };
    success: boolean;
  };

  expect(response.ok()).toBeTruthy();
  expect(response.headers()["x-request-id"]).toBeTruthy();
  expect(payload.success).toBeTruthy();
  expect(payload.meta.requestId).toBeTruthy();
});
