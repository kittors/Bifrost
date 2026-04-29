import { describe, expect, it, vi } from "vitest";

import { createServiceAccessURL, listClientServices } from "./api";

describe("desktop services api", () => {
  it("loads accessible services from the gateway", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        json: async () => ({
          success: true,
          data: {
            items: [
              {
                id: "service_gitlab",
                key: "gitlab",
                name: "GitLab",
                description: "",
                group: "development",
                status: "enabled",
                publicPath: "/s/gitlab/",
                accessSource: "role",
              },
            ],
          },
          meta: { requestId: "req_services", timestamp: "2026-04-18T04:10:00Z" },
        }),
        status: 200,
      }),
    );

    const result = await listClientServices({
      accessToken: "access",
      baseURL: "http://127.0.0.1:8080",
    });

    expect(result).toHaveLength(1);
    expect(result[0]?.key).toBe("gitlab");
  });

  it("requests a service access url", async () => {
    vi.stubGlobal(
      "fetch",
      vi.fn().mockResolvedValue({
        json: async () => ({
          success: true,
          data: { publicPath: "/s/gitlab/", expiresIn: 300, accessTicket: "ticket" },
          meta: { requestId: "req_service_access", timestamp: "2026-04-18T04:12:00Z" },
        }),
        status: 200,
      }),
    );

    const result = await createServiceAccessURL({
      accessToken: "access",
      baseURL: "http://127.0.0.1:8080",
      serviceId: "service_gitlab",
    });

    expect(result.accessTicket).toBe("ticket");
  });
});
