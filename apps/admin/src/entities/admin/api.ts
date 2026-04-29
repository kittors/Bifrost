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
  UserServiceOverride,
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

export async function getAdminUser(input: { accessToken: string; userID: string }) {
  const payload = await requestJSON<AdminUser>({
    accessToken: input.accessToken,
    baseURL: getGatewayBaseURL(),
    path: `/api/v1/admin/users/${input.userID}`,
  });

  return payload.data;
}

export async function resetAdminUserPassword(input: {
  accessToken: string;
  password: string;
  userID: string;
}) {
  const payload = await requestJSON<{ reset: boolean }>({
    accessToken: input.accessToken,
    baseURL: getGatewayBaseURL(),
    body: {
      password: input.password,
    },
    method: "POST",
    path: `/api/v1/admin/users/${input.userID}/reset-password`,
  });

  return payload.data;
}

export async function setAdminUserStatus(input: {
  accessToken: string;
  status: string;
  userID: string;
}) {
  const payload = await requestJSON<AdminUser>({
    accessToken: input.accessToken,
    baseURL: getGatewayBaseURL(),
    body: {
      status: input.status,
    },
    method: "POST",
    path: `/api/v1/admin/users/${input.userID}/status`,
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

export async function createAdminRole(input: {
  accessToken: string;
  description: string;
  displayName: string;
  name: string;
}) {
  const payload = await requestJSON<AdminRole>({
    accessToken: input.accessToken,
    baseURL: getGatewayBaseURL(),
    body: {
      description: input.description,
      displayName: input.displayName,
      name: input.name,
    },
    method: "POST",
    path: "/api/v1/admin/roles",
  });

  return payload.data;
}

export async function updateAdminRole(input: {
  accessToken: string;
  description: string;
  displayName: string;
  roleID: string;
}) {
  const payload = await requestJSON<AdminRole>({
    accessToken: input.accessToken,
    baseURL: getGatewayBaseURL(),
    body: {
      description: input.description,
      displayName: input.displayName,
    },
    method: "PATCH",
    path: `/api/v1/admin/roles/${input.roleID}`,
  });

  return payload.data;
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

export async function createAdminService(input: {
  accessToken: string;
  description: string;
  enabled: boolean;
  group: string;
  key: string;
  name: string;
  protocol: string;
  publicPath: string;
  upstreamUrl: string;
}) {
  const payload = await requestJSON<AdminService>({
    accessToken: input.accessToken,
    baseURL: getGatewayBaseURL(),
    body: {
      description: input.description,
      enabled: input.enabled,
      group: input.group,
      key: input.key,
      name: input.name,
      protocol: input.protocol,
      publicPath: input.publicPath,
      upstreamUrl: input.upstreamUrl,
    },
    method: "POST",
    path: "/api/v1/admin/services",
  });

  return payload.data;
}

export async function getAdminService(input: { accessToken: string; serviceID: string }) {
  const payload = await requestJSON<AdminService>({
    accessToken: input.accessToken,
    baseURL: getGatewayBaseURL(),
    path: `/api/v1/admin/services/${input.serviceID}`,
  });

  return payload.data;
}

export async function updateAdminService(input: {
  accessToken: string;
  description: string;
  group: string;
  name: string;
  protocol: string;
  publicPath: string;
  serviceID: string;
  upstreamUrl: string;
}) {
  const payload = await requestJSON<AdminService>({
    accessToken: input.accessToken,
    baseURL: getGatewayBaseURL(),
    body: {
      description: input.description,
      group: input.group,
      name: input.name,
      protocol: input.protocol,
      publicPath: input.publicPath,
      upstreamUrl: input.upstreamUrl,
    },
    method: "PATCH",
    path: `/api/v1/admin/services/${input.serviceID}`,
  });

  return payload.data;
}

export async function setAdminServiceStatus(input: {
  accessToken: string;
  serviceID: string;
  status: string;
}) {
  const payload = await requestJSON<AdminService>({
    accessToken: input.accessToken,
    baseURL: getGatewayBaseURL(),
    body: {
      status: input.status,
    },
    method: "POST",
    path: `/api/v1/admin/services/${input.serviceID}/status`,
  });

  return payload.data;
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

export async function getAdminDevice(input: { accessToken: string; deviceID: string }) {
  const payload = await requestJSON<AdminDevice>({
    accessToken: input.accessToken,
    baseURL: getGatewayBaseURL(),
    path: `/api/v1/admin/devices/${input.deviceID}`,
  });

  return payload.data;
}

export async function setAdminDeviceStatus(input: {
  accessToken: string;
  deviceID: string;
  status: string;
}) {
  const payload = await requestJSON<AdminDevice>({
    accessToken: input.accessToken,
    baseURL: getGatewayBaseURL(),
    body: {
      status: input.status,
    },
    method: "POST",
    path: `/api/v1/admin/devices/${input.deviceID}/status`,
  });

  return payload.data;
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

export async function replaceRoleServices(input: {
  accessToken: string;
  roleID: string;
  serviceIDs: string[];
}) {
  const payload = await requestJSON<{ roleId: string; serviceIds: string[] }>({
    accessToken: input.accessToken,
    baseURL: getGatewayBaseURL(),
    body: {
      serviceIds: input.serviceIDs,
    },
    method: "PUT",
    path: `/api/v1/admin/roles/${input.roleID}/services`,
  });

  return payload.data;
}

export async function listUserServiceOverrides(input: { accessToken: string; userID: string }) {
  const payload = await requestJSON<{ items: UserServiceOverride[] }>({
    accessToken: input.accessToken,
    baseURL: getGatewayBaseURL(),
    path: `/api/v1/admin/users/${input.userID}/service-overrides`,
  });

  return payload.data.items;
}

export async function replaceUserServiceOverrides(input: {
  accessToken: string;
  allowServiceIDs: string[];
  denyServiceIDs: string[];
  userID: string;
}) {
  const payload = await requestJSON<{ items: UserServiceOverride[] }>({
    accessToken: input.accessToken,
    baseURL: getGatewayBaseURL(),
    body: {
      allowServiceIds: input.allowServiceIDs,
      denyServiceIds: input.denyServiceIDs,
    },
    method: "PUT",
    path: `/api/v1/admin/users/${input.userID}/service-overrides`,
  });

  return payload.data.items;
}

export function requireAccessToken(session: StoredAdminSession | null) {
  if (!session) {
    throw new Error("admin session is required");
  }

  return session.accessToken;
}
