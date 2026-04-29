import type { ApiMeta, ApiResponse } from "@bifrost/contracts";
import { ERROR_CODES, type ErrorCode } from "@bifrost/contracts";

export class ApiClientError extends Error {
  code: ErrorCode;
  requestId: string;
  statusCode: number;
  userMessage: string;

  constructor(input: {
    code: ErrorCode;
    message: string;
    requestId: string;
    statusCode: number;
    userMessage: string;
  }) {
    super(input.message);
    this.name = "ApiClientError";
    this.code = input.code;
    this.requestId = input.requestId;
    this.statusCode = input.statusCode;
    this.userMessage = input.userMessage;
  }
}

export function buildApiURL(baseURL: string, path: string) {
  const normalizedBaseURL = baseURL.replace(/\/+$/, "");
  const normalizedPath = path.startsWith("/") ? path : `/${path}`;
  return `${normalizedBaseURL}${normalizedPath}`;
}

export function buildQueryString(query?: Record<string, string | number | undefined>) {
  if (!query) {
    return "";
  }

  const params = new URLSearchParams();

  for (const [key, value] of Object.entries(query)) {
    if (value === undefined || value === "") {
      continue;
    }

    params.set(key, String(value));
  }

  const encoded = params.toString();
  return encoded ? `?${encoded}` : "";
}

export function parseApiResponse<T>(payload: ApiResponse<T>) {
  if (payload.success) {
    return {
      data: payload.data,
      meta: payload.meta,
    };
  }

  throw new ApiClientError({
    code: payload.error.code,
    message: payload.error.message,
    requestId: payload.meta.requestId,
    statusCode: 400,
    userMessage: payload.error.userMessage,
  });
}

function unknownErrorCode(): ErrorCode {
  return ERROR_CODES.includes("COMMON_INTERNAL_ERROR") ? "COMMON_INTERNAL_ERROR" : ERROR_CODES[0];
}

function unknownMeta(): ApiMeta {
  return {
    requestId: "",
    timestamp: new Date().toISOString(),
  };
}

export async function requestJSON<T>(input: {
  accessToken?: string;
  baseURL: string;
  body?: unknown;
  method?: "GET" | "POST" | "PATCH" | "PUT";
  path: string;
  query?: Record<string, string | number | undefined>;
  signal?: AbortSignal;
}) {
  const response = await fetch(
    buildApiURL(input.baseURL, input.path) + buildQueryString(input.query),
    {
      body: input.body === undefined ? undefined : JSON.stringify(input.body),
      headers: {
        ...(input.body === undefined ? {} : { "Content-Type": "application/json" }),
        ...(input.accessToken ? { Authorization: `Bearer ${input.accessToken}` } : {}),
      },
      method: input.method ?? "GET",
      signal: input.signal,
    },
  );

  let payload: ApiResponse<T> | null = null;

  try {
    payload = (await response.json()) as ApiResponse<T>;
  } catch {
    throw new ApiClientError({
      code: unknownErrorCode(),
      message: "response body must be valid JSON",
      requestId: "",
      statusCode: response.status,
      userMessage: "服务暂时不可用，请稍后再试",
    });
  }

  try {
    return parseApiResponse(payload);
  } catch (error) {
    if (error instanceof ApiClientError) {
      error.statusCode = response.status;
      if (!error.requestId) {
        error.requestId = payload?.meta.requestId ?? "";
      }
    }

    throw error;
  }
}

export function normalizeUnknownError(error: unknown) {
  if (error instanceof ApiClientError) {
    return error;
  }

  return new ApiClientError({
    code: unknownErrorCode(),
    message: error instanceof Error ? error.message : "unknown error",
    requestId: unknownMeta().requestId,
    statusCode: 500,
    userMessage: "服务暂时不可用，请稍后再试",
  });
}
