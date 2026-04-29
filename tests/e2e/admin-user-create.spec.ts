import { expect, test } from "@playwright/test";

import { adminLogin, createAdminUser, listAdminUsers } from "./fixtures/client-api";
import { seedPassword } from "./fixtures/env";

test("admin can create a developer user through the gateway API", async ({ request }) => {
  const { session } = await adminLogin(request, "admin", seedPassword);
  const username = `e2e-user-${Date.now()}`;

  const created = await createAdminUser(request, session.accessToken, {
    displayName: "E2E User",
    email: `${username}@example.com`,
    password: seedPassword,
    roleIds: ["role_developer"],
    username,
  });

  expect(created.username).toBe(username);
  expect(created.roles).toEqual(["role_developer"]);

  const users = await listAdminUsers(request, session.accessToken, username);
  expect(users.map((user) => user.id)).toContain(created.id);
});
