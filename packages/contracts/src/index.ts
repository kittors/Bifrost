export type ApiMeta = {
  requestId: string;
  timestamp: string;
};

export type ApiError = {
  code: string;
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
