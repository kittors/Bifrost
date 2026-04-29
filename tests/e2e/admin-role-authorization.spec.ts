import { expect, test } from "@playwright/test";

import {
  adminLogin,
  bootstrapClientDevice,
  createAdminRole,
  createAdminUser,
  listClientServices,
  replaceRoleServices,
} from "./fixtures/client-api";
import { seedPassword } from "./fixtures/env";

test("admin role authorization changes are immediately visible to a newly created user", async ({
  request,
}) => {
  const { session } = await adminLogin(request, "admin", seedPassword);
  const roleName = `role-e2e-${Date.now()}`;
  const username = `role-user-${Date.now()}`;

  const role = await createAdminRole(request, session.accessToken, {
    description: "E2E role",
    displayName: "E2E Role",
    name: roleName,
  });

  await replaceRoleServices(request, session.accessToken, role.id, [
    "service_jenkins",
    "service_docs",
  ]);

  await createAdminUser(request, session.accessToken, {
    displayName: "Role User",
    email: `${username}@example.com`,
    password: seedPassword,
    roleIds: [role.id],
    username,
  });

  const client = await bootstrapClientDevice(request, username, seedPassword);
  const services = await listClientServices(request, client.session.accessToken);

  expect(services.map((service) => service.key)).toEqual(["docs", "jenkins"]);
});
