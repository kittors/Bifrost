import { expect, test } from "@playwright/test";

import { adminLogin } from "./fixtures/client-api";
import { seedPassword } from "./fixtures/env";

test("admin can login through the gateway API and receives requestId", async ({ request }) => {
  const { meta, session } = await adminLogin(request, "admin", seedPassword);

  expect(session.user.username).toBe("admin");
  expect(session.user.roles).toContain("role_admin");
  expect(meta.requestId).not.toHaveLength(0);
});
