import { describe, expect, it } from "vitest";

import { ApiClientError, buildApiURL, parseApiResponse } from "./http";

describe("admin http helpers", () => {
  it("builds API URLs from a normalized base URL", () => {
    expect(buildApiURL("http://127.0.0.1:8080/", "/api/v1/admin/users")).toBe(
      "http://127.0.0.1:8080/api/v1/admin/users",
    );
    expect(buildApiURL("http://127.0.0.1:8080", "api/v1/admin/users")).toBe(
      "http://127.0.0.1:8080/api/v1/admin/users",
    );
  });

  it("returns success payload data and meta", () => {
    const payload = parseApiResponse<{ accessToken: string }>({
      data: { accessToken: "token_01" },
      error: null,
      meta: {
        requestId: "req_success_01",
        timestamp: "2026-04-17T15:00:00Z",
      },
      success: true,
    });

    expect(payload.data.accessToken).toBe("token_01");
    expect(payload.meta.requestId).toBe("req_success_01");
  });

  it("throws a typed API error for failure responses", () => {
    expect(() =>
      parseApiResponse({
        data: null,
        error: {
          code: "AUTH_INVALID_CREDENTIALS",
          details: {},
          message: "invalid credentials",
          userMessage: "用户名或密码错误",
        },
        meta: {
          requestId: "req_failure_01",
          timestamp: "2026-04-17T15:05:00Z",
        },
        success: false,
      }),
    ).toThrowError(ApiClientError);

    try {
      parseApiResponse({
        data: null,
        error: {
          code: "AUTH_INVALID_CREDENTIALS",
          details: {},
          message: "invalid credentials",
          userMessage: "用户名或密码错误",
        },
        meta: {
          requestId: "req_failure_01",
          timestamp: "2026-04-17T15:05:00Z",
        },
        success: false,
      });
    } catch (error) {
      expect(error).toBeInstanceOf(ApiClientError);
      expect((error as ApiClientError).requestId).toBe("req_failure_01");
      expect((error as ApiClientError).code).toBe("AUTH_INVALID_CREDENTIALS");
    }
  });
});
