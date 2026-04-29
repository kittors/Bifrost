import { describe, expect, it } from "vitest";

import { ApiClientError, normalizeUnknownError, resolveApiErrorMessage } from "./http";

describe("desktop http helpers", () => {
  it("normalizes typed API errors while preserving requestId for shared error UI", () => {
    const error = new ApiClientError({
      code: "POLICY_ACCESS_DENIED",
      message: "policy denied",
      requestId: "req_desktop_denied",
      statusCode: 403,
      userMessage: "你暂时没有访问该服务的权限",
    });

    const normalized = normalizeUnknownError(error);

    expect(normalized.userMessage).toBe("你暂时没有访问该服务的权限");
    expect(normalized.requestId).toBe("req_desktop_denied");
    expect(resolveApiErrorMessage(error, "打开服务失败")).toBe("你暂时没有访问该服务的权限");
  });

  it("normalizes unknown failures to the same fallback copy used by admin", () => {
    const normalized = normalizeUnknownError(new Error("socket closed"));

    expect(normalized.userMessage).toBe("服务暂时不可用，请稍后再试");
    expect(normalized.requestId).toBe("");
  });
});
