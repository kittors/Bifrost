export type { AuditEventType, ErrorCode } from "./generated";
export { AUDIT_EVENT_TYPES, ERROR_CODES } from "./generated";

export type PaginationMeta = {
  page: number;
  pageSize: number;
  total: number;
  totalPages: number;
};

export type ApiMeta = {
  requestId: string;
  timestamp: string;
  pagination?: PaginationMeta;
};

export type ApiError = {
  code: import("./generated").ErrorCode;
  message: string;
  userMessage: string;
  details: Record<string, unknown>;
};

export type ApiSuccess<T> = {
  success: true;
  data: T;
  meta: ApiMeta;
  error: null;
};

export type ApiFailure = {
  success: false;
  data: null;
  meta: ApiMeta;
  error: ApiError;
};

export type ApiResponse<T> = ApiSuccess<T> | ApiFailure;

export type PaginatedItems<T> = {
  items: T[];
};

type ApiSuccessInput<T> = {
  data: T;
  requestId: string;
  timestamp: string;
  pagination?: PaginationMeta;
};

type ApiErrorInput = {
  code: import("./generated").ErrorCode;
  message: string;
  userMessage: string;
  details?: Record<string, unknown>;
};

export function createApiSuccess<T>({
  data,
  pagination,
  requestId,
  timestamp,
}: ApiSuccessInput<T>): ApiSuccess<T> {
  return {
    success: true,
    data,
    meta: {
      requestId,
      timestamp,
      ...(pagination ? { pagination } : {}),
    },
    error: null,
  };
}

export function createApiError({
  code,
  details = {},
  message,
  userMessage,
}: ApiErrorInput): ApiError {
  return {
    code,
    message,
    userMessage,
    details,
  };
}

export function createApiFailure({
  error,
  requestId,
  timestamp,
}: {
  error: ApiError;
  requestId: string;
  timestamp: string;
}): ApiFailure {
  return {
    success: false,
    data: null,
    meta: {
      requestId,
      timestamp,
    },
    error,
  };
}

export function isApiSuccess<T>(response: ApiResponse<T>): response is ApiSuccess<T> {
  return response.success;
}

export function isApiFailure<T>(response: ApiResponse<T>): response is ApiFailure {
  return !response.success;
}
