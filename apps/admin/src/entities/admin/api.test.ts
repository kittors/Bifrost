import { afterEach, describe, expect, it, vi } from "vitest";

import { createAdminRole, createAdminService, replaceRoleServices } from "./api";

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
      "http://127.0.0.1:8080/api/v1/admin/roles",
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
      "http://127.0.0.1:8080/api/v1/admin/services",
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
      "http://127.0.0.1:8080/api/v1/admin/roles/role_developer/services",
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
});
