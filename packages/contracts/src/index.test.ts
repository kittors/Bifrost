import { describe, expect, it } from "vitest";

import {
  type ApiError,
  type ApiResponse,
  AUDIT_EVENT_TYPES,
  createApiError,
  createApiSuccess,
  ERROR_CODES,
  isApiFailure,
  isApiSuccess,
} from "./index";

describe("contracts", () => {
  it("creates success envelopes with a unified meta structure", () => {
    const response = createApiSuccess({
      data: { items: [{ id: "svc_gitlab" }] },
      requestId: "req_01JZABCDEF1234567890",
      timestamp: "2026-04-17T10:30:00Z",
    });

    expect(response.success).toBe(true);
    expect(response.error).toBeNull();
    expect(response.meta.requestId).toBe("req_01JZABCDEF1234567890");
    expect(response.data.items).toHaveLength(1);
    expect(isApiSuccess(response)).toBe(true);
    expect(isApiFailure(response)).toBe(false);
  });

  it("creates failure envelopes with stable error metadata", () => {
    const error = createApiError({
      code: "AUTH_INVALID_TOKEN",
      message: "token is invalid or expired",
      userMessage: "登录状态已失效，请重新登录",
    });
    const response: ApiResponse<null> = {
      success: false,
      data: null,
      error,
      meta: {
        requestId: "req_01JZFAIL",
        timestamp: "2026-04-17T10:35:00Z",
      },
    };

    expect(response.success).toBe(false);
    expect(response.data).toBeNull();
    expect(response.error.code).toBe("AUTH_INVALID_TOKEN");
    expect(isApiFailure(response)).toBe(true);
    expect(isApiSuccess(response)).toBe(false);
  });

  it("exports documented error codes and audit event types", () => {
    expect(ERROR_CODES).toContain("POLICY_ACCESS_DENIED");
    expect(ERROR_CODES).toContain("GATEWAY_UPSTREAM_TIMEOUT");
    expect(AUDIT_EVENT_TYPES).toContain("auth.login.succeeded");
    expect(AUDIT_EVENT_TYPES).toContain("service.access.denied");
  });

  it("keeps ApiError details optional but always object shaped", () => {
    const error: ApiError = createApiError({
      code: "COMMON_BAD_REQUEST",
      message: "missing field",
      userMessage: "请求参数不正确",
    });

    expect(error.details).toEqual({});
  });
});
