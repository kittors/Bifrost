import { type ApiResponse, ERROR_CODES, type ErrorCode } from "@bifrost/contracts";

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
    this.code = input.code;
    this.name = "ApiClientError";
    this.requestId = input.requestId;
    this.statusCode = input.statusCode;
    this.userMessage = input.userMessage;
  }
}

// 统一把接口错误转换为用户可读文案，避免 UI 直接暴露底层 message。
export function resolveApiErrorMessage(error: unknown, fallbackMessage: string) {
  if (error instanceof ApiClientError) {
    return error.userMessage;
  }

  if (error instanceof Error) {
    return error.message;
  }

  return fallbackMessage;
}

function unknownErrorCode(): ErrorCode {
  return ERROR_CODES.includes("COMMON_INTERNAL_ERROR") ? "COMMON_INTERNAL_ERROR" : ERROR_CODES[0];
}

export async function requestJSON<T>(input: {
  accessToken?: string;
  baseURL: string;
  body?: unknown;
  method?: "GET" | "POST";
  path: string;
  query?: Record<string, string | number | undefined>;
}) {
  const url = new URL(input.path, `${input.baseURL.replace(/\/+$/, "")}/`);
  for (const [key, value] of Object.entries(input.query ?? {})) {
    if (value !== undefined && value !== "") {
      url.searchParams.set(key, String(value));
    }
  }

  const response = await fetch(url, {
    body: input.body === undefined ? undefined : JSON.stringify(input.body),
    headers: {
      ...(input.accessToken ? { Authorization: `Bearer ${input.accessToken}` } : {}),
      ...(input.body === undefined ? {} : { "Content-Type": "application/json" }),
    },
    method: input.method ?? "GET",
  });

  const payload = (await response.json()) as ApiResponse<T>;
  if (payload.success) {
    return payload.data;
  }

  throw new ApiClientError({
    code: payload.error.code ?? unknownErrorCode(),
    message: payload.error.message,
    requestId: payload.meta.requestId,
    statusCode: response.status,
    userMessage: payload.error.userMessage,
  });
}
