import type { StoredAdminSession } from "../../features/auth/session";
import { getGatewayBaseURL } from "../../shared/config/env";
import { requestJSON } from "../../shared/lib/http";
import type {
  AdminAuditEvent,
  AdminDevice,
  AdminRole,
  AdminService,
  AdminUser,
  PaginatedResult,
} from "./types";

function unwrapPaginated<T>(payload: Awaited<ReturnType<typeof requestJSON<{ items: T[] }>>>) {
  return {
    items: payload.data.items,
    total: payload.meta.pagination?.total ?? payload.data.items.length,
  } satisfies PaginatedResult<T>;
}

export async function listAdminUsers(input: {
  accessToken: string;
  keyword?: string;
  page?: number;
  pageSize?: number;
  roleId?: string;
  status?: string;
}) {
  const payload = await requestJSON<{ items: AdminUser[] }>({
    accessToken: input.accessToken,
    baseURL: getGatewayBaseURL(),
    path: "/api/v1/admin/users",
    query: {
      keyword: input.keyword,
      page: input.page ?? 1,
      pageSize: input.pageSize ?? 20,
      roleId: input.roleId,
      status: input.status,
    },
  });

  return unwrapPaginated(payload);
}

export async function createAdminUser(input: {
  accessToken: string;
  displayName: string;
  email: string;
  password: string;
  roleIds: string[];
  username: string;
}) {
  const payload = await requestJSON<AdminUser>({
    accessToken: input.accessToken,
    baseURL: getGatewayBaseURL(),
    body: {
      displayName: input.displayName,
      email: input.email,
      password: input.password,
      roleIds: input.roleIds,
      username: input.username,
    },
    method: "POST",
    path: "/api/v1/admin/users",
  });

  return payload.data;
}

export async function listAdminRoles(input: {
  accessToken: string;
  keyword?: string;
  page?: number;
  pageSize?: number;
}) {
  const payload = await requestJSON<{ items: AdminRole[] }>({
    accessToken: input.accessToken,
    baseURL: getGatewayBaseURL(),
    path: "/api/v1/admin/roles",
    query: {
      keyword: input.keyword,
      page: input.page ?? 1,
      pageSize: input.pageSize ?? 50,
    },
  });

  return unwrapPaginated(payload);
}

export async function listAdminServices(input: {
  accessToken: string;
  group?: string;
  keyword?: string;
  page?: number;
  pageSize?: number;
  status?: string;
}) {
  const payload = await requestJSON<{ items: AdminService[] }>({
    accessToken: input.accessToken,
    baseURL: getGatewayBaseURL(),
    path: "/api/v1/admin/services",
    query: {
      group: input.group,
      keyword: input.keyword,
      page: input.page ?? 1,
      pageSize: input.pageSize ?? 20,
      status: input.status,
    },
  });

  return unwrapPaginated(payload);
}

export async function listAdminDevices(input: {
  accessToken: string;
  keyword?: string;
  page?: number;
  pageSize?: number;
  status?: string;
  userId?: string;
}) {
  const payload = await requestJSON<{ items: AdminDevice[] }>({
    accessToken: input.accessToken,
    baseURL: getGatewayBaseURL(),
    path: "/api/v1/admin/devices",
    query: {
      keyword: input.keyword,
      page: input.page ?? 1,
      pageSize: input.pageSize ?? 20,
      status: input.status,
      userId: input.userId,
    },
  });

  return unwrapPaginated(payload);
}

export async function listAdminAuditEvents(input: {
  accessToken: string;
  page?: number;
  pageSize?: number;
  result?: string;
  type?: string;
}) {
  const payload = await requestJSON<{ items: AdminAuditEvent[] }>({
    accessToken: input.accessToken,
    baseURL: getGatewayBaseURL(),
    path: "/api/v1/admin/audit-events",
    query: {
      page: input.page ?? 1,
      pageSize: input.pageSize ?? 10,
      result: input.result,
      type: input.type,
    },
  });

  return unwrapPaginated(payload);
}

export function requireAccessToken(session: StoredAdminSession | null) {
  if (!session) {
    throw new Error("admin session is required");
  }

  return session.accessToken;
}
