import { afterEach, describe, expect, it, vi } from "vitest";

import {
  createAdminRole,
  createAdminService,
  listUserServiceOverrides,
  replaceRoleServices,
  replaceUserServiceOverrides,
  resetAdminUserPassword,
  setAdminDeviceStatus,
  updateAdminRole,
  updateAdminService,
} from "./api";

function jsonResponse(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), {
    headers: {
      "Content-Type": "application/json",
    },
    status,
  });
}

describe("admin entity api helpers", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("creates an admin role through the gateway API", async () => {
    const fetchMock = vi.fn(async () =>
      jsonResponse(
        {
          data: {
            description: "研发角色",
            displayName: "研发",
            id: "role_developer",
            name: "developer",
          },
          error: null,
          meta: {
            requestId: "req_role_create_01",
            timestamp: "2026-04-17T15:20:00Z",
          },
          success: true,
        },
        201,
      ),
    );
    vi.stubGlobal("fetch", fetchMock);

    const role = await createAdminRole({
      accessToken: "access_01",
      description: "研发角色",
      displayName: "研发",
      name: "developer",
    });

    expect(role.id).toBe("role_developer");
    expect(fetchMock).toHaveBeenCalledWith(
      "/__bifrost_gateway__/api/v1/admin/roles",
      expect.objectContaining({
        body: JSON.stringify({
          description: "研发角色",
          displayName: "研发",
          name: "developer",
        }),
        headers: expect.objectContaining({
          Authorization: "Bearer access_01",
          "Content-Type": "application/json",
        }),
        method: "POST",
      }),
    );
  });

  it("creates an admin service through the gateway API", async () => {
    const fetchMock = vi.fn(async () =>
      jsonResponse(
        {
          data: {
            description: "研发代码平台",
            group: "development",
            id: "service_gitlab",
            key: "gitlab",
            name: "GitLab",
            protocol: "https",
            publicPath: "/s/gitlab",
            status: "enabled",
            upstreamUrl: "http://mock-gitlab:8080",
          },
          error: null,
          meta: {
            requestId: "req_service_create_01",
            timestamp: "2026-04-17T15:25:00Z",
          },
          success: true,
        },
        201,
      ),
    );
    vi.stubGlobal("fetch", fetchMock);

    const service = await createAdminService({
      accessToken: "access_01",
      description: "研发代码平台",
      enabled: true,
      group: "development",
      key: "gitlab",
      name: "GitLab",
      protocol: "https",
      publicPath: "/s/gitlab",
      upstreamUrl: "http://mock-gitlab:8080",
    });

    expect(service.id).toBe("service_gitlab");
    expect(fetchMock).toHaveBeenCalledWith(
      "/__bifrost_gateway__/api/v1/admin/services",
      expect.objectContaining({
        body: JSON.stringify({
          description: "研发代码平台",
          enabled: true,
          group: "development",
          key: "gitlab",
          name: "GitLab",
          protocol: "https",
          publicPath: "/s/gitlab",
          upstreamUrl: "http://mock-gitlab:8080",
        }),
        headers: expect.objectContaining({
          Authorization: "Bearer access_01",
          "Content-Type": "application/json",
        }),
        method: "POST",
      }),
    );
  });

  it("replaces role services through the gateway API", async () => {
    const fetchMock = vi.fn(async () =>
      jsonResponse({
        data: {
          roleId: "role_developer",
          serviceIds: ["service_gitlab", "service_docs"],
        },
        error: null,
        meta: {
          requestId: "req_role_services_01",
          timestamp: "2026-04-17T15:27:00Z",
        },
        success: true,
      }),
    );
    vi.stubGlobal("fetch", fetchMock);

    await replaceRoleServices({
      accessToken: "access_01",
      roleID: "role_developer",
      serviceIDs: ["service_gitlab", "service_docs"],
    });

    expect(fetchMock).toHaveBeenCalledWith(
      "/__bifrost_gateway__/api/v1/admin/roles/role_developer/services",
      expect.objectContaining({
        body: JSON.stringify({
          serviceIds: ["service_gitlab", "service_docs"],
        }),
        headers: expect.objectContaining({
          Authorization: "Bearer access_01",
          "Content-Type": "application/json",
        }),
        method: "PUT",
      }),
    );
  });

  it("resets an admin user password through the gateway API", async () => {
    const fetchMock = vi.fn(async () =>
      jsonResponse({
        data: {
          reset: true,
        },
        error: null,
        meta: {
          requestId: "req_reset_password_01",
          timestamp: "2026-04-18T00:12:00Z",
        },
        success: true,
      }),
    );
    vi.stubGlobal("fetch", fetchMock);

    await resetAdminUserPassword({
      accessToken: "access_01",
      password: "NewPassword123!",
      userID: "user_alice",
    });

    expect(fetchMock).toHaveBeenCalledWith(
      "/__bifrost_gateway__/api/v1/admin/users/user_alice/reset-password",
      expect.objectContaining({
        body: JSON.stringify({
          password: "NewPassword123!",
        }),
        headers: expect.objectContaining({
          Authorization: "Bearer access_01",
          "Content-Type": "application/json",
        }),
        method: "POST",
      }),
    );
  });

  it("updates an admin service through the gateway API", async () => {
    const fetchMock = vi.fn(async () =>
      jsonResponse({
        data: {
          description: "共享文档",
          group: "shared",
          id: "service_docs",
          key: "docs",
          name: "Docs Portal",
          protocol: "http",
          publicPath: "/s/docs",
          status: "enabled",
          upstreamUrl: "http://mock-docs:8080",
        },
        error: null,
        meta: {
          requestId: "req_service_update_01",
          timestamp: "2026-04-18T00:13:00Z",
        },
        success: true,
      }),
    );
    vi.stubGlobal("fetch", fetchMock);

    const service = await updateAdminService({
      accessToken: "access_01",
      description: "共享文档",
      group: "shared",
      name: "Docs Portal",
      protocol: "http",
      publicPath: "/s/docs",
      serviceID: "service_docs",
      upstreamUrl: "http://mock-docs:8080",
    });

    expect(service.name).toBe("Docs Portal");
    expect(fetchMock).toHaveBeenCalledWith(
      "/__bifrost_gateway__/api/v1/admin/services/service_docs",
      expect.objectContaining({
        body: JSON.stringify({
          description: "共享文档",
          group: "shared",
          name: "Docs Portal",
          protocol: "http",
          publicPath: "/s/docs",
          upstreamUrl: "http://mock-docs:8080",
        }),
        headers: expect.objectContaining({
          Authorization: "Bearer access_01",
          "Content-Type": "application/json",
        }),
        method: "PATCH",
      }),
    );
  });

  it("sets an admin device status through the gateway API", async () => {
    const fetchMock = vi.fn(async () =>
      jsonResponse({
        data: {
          clientVersion: "1.0.0",
          id: "device_01",
          name: "Alice Mac",
          os: "macOS",
          publicKeyFingerprint: "fp_01",
          status: "disabled",
          userId: "user_alice",
          userUsername: "alice",
        },
        error: null,
        meta: {
          requestId: "req_device_status_01",
          timestamp: "2026-04-18T00:14:00Z",
        },
        success: true,
      }),
    );
    vi.stubGlobal("fetch", fetchMock);

    await setAdminDeviceStatus({
      accessToken: "access_01",
      deviceID: "device_01",
      status: "disabled",
    });

    expect(fetchMock).toHaveBeenCalledWith(
      "/__bifrost_gateway__/api/v1/admin/devices/device_01/status",
      expect.objectContaining({
        body: JSON.stringify({
          status: "disabled",
        }),
        headers: expect.objectContaining({
          Authorization: "Bearer access_01",
          "Content-Type": "application/json",
        }),
        method: "POST",
      }),
    );
  });

  it("updates an admin role through the gateway API", async () => {
    const fetchMock = vi.fn(async () =>
      jsonResponse({
        data: {
          description: "运维私有服务访问角色",
          displayName: "运维团队",
          id: "role_ops",
          name: "ops",
        },
        error: null,
        meta: {
          requestId: "req_role_update_01",
          timestamp: "2026-04-18T00:31:00Z",
        },
        success: true,
      }),
    );
    vi.stubGlobal("fetch", fetchMock);

    const role = await updateAdminRole({
      accessToken: "access_01",
      description: "运维私有服务访问角色",
      displayName: "运维团队",
      roleID: "role_ops",
    });

    expect(role.displayName).toBe("运维团队");
    expect(fetchMock).toHaveBeenCalledWith(
      "/__bifrost_gateway__/api/v1/admin/roles/role_ops",
      expect.objectContaining({
        body: JSON.stringify({
          description: "运维私有服务访问角色",
          displayName: "运维团队",
        }),
        method: "PATCH",
      }),
    );
  });

  it("lists user service overrides through the gateway API", async () => {
    const fetchMock = vi.fn(async () =>
      jsonResponse({
        data: {
          items: [
            { effect: "allow", serviceId: "service_docs" },
            { effect: "deny", serviceId: "service_gitlab" },
          ],
        },
        error: null,
        meta: {
          requestId: "req_user_overrides_list_01",
          timestamp: "2026-04-18T00:32:00Z",
        },
        success: true,
      }),
    );
    vi.stubGlobal("fetch", fetchMock);

    const overrides = await listUserServiceOverrides({
      accessToken: "access_01",
      userID: "user_alice",
    });

    expect(overrides).toHaveLength(2);
    expect(fetchMock).toHaveBeenCalledWith(
      "/__bifrost_gateway__/api/v1/admin/users/user_alice/service-overrides",
      expect.objectContaining({
        method: "GET",
      }),
    );
  });

  it("replaces user service overrides through the gateway API", async () => {
    const fetchMock = vi.fn(async () =>
      jsonResponse({
        data: {
          items: [
            { effect: "allow", serviceId: "service_docs" },
            { effect: "deny", serviceId: "service_gitlab" },
          ],
        },
        error: null,
        meta: {
          requestId: "req_user_overrides_replace_01",
          timestamp: "2026-04-18T00:33:00Z",
        },
        success: true,
      }),
    );
    vi.stubGlobal("fetch", fetchMock);

    const overrides = await replaceUserServiceOverrides({
      accessToken: "access_01",
      allowServiceIDs: ["service_docs"],
      denyServiceIDs: ["service_gitlab"],
      userID: "user_alice",
    });

    expect(overrides[0]?.serviceId).toBe("service_docs");
    expect(fetchMock).toHaveBeenCalledWith(
      "/__bifrost_gateway__/api/v1/admin/users/user_alice/service-overrides",
      expect.objectContaining({
        body: JSON.stringify({
          allowServiceIds: ["service_docs"],
          denyServiceIds: ["service_gitlab"],
        }),
        method: "PUT",
      }),
    );
  });
});
