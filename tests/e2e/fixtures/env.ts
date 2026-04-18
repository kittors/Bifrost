export const gatewayBaseURL =
  process.env.BIFROST_PUBLIC_BASE_URL?.replace(/\/+$/, "") ?? "http://127.0.0.1:18080";

export const adminBaseURL =
  process.env.BIFROST_ADMIN_BASE_URL?.replace(/\/+$/, "") ?? "http://127.0.0.1:15173";

export const seedPassword = "ChangeMe123!";
