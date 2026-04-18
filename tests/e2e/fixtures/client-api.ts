import { generateKeyPairSync } from "node:crypto";

import { type APIRequestContext, expect } from "@playwright/test";

import { gatewayBaseURL } from "./env";

type LoginPayload = {
  accessToken: string;
  refreshToken: string;
  user: {
    displayName: string;
    id: string;
    roles: string[];
    username: string;
  };
};

function generateDeviceIdentity() {
  const { publicKey } = generateKeyPairSync("ed25519");
  const publicKeyDER = publicKey.export({ format: "der", type: "spki" });
  const publicKeyRaw = Buffer.from(publicKeyDER).subarray(-32);
  const encodedPublicKey = Buffer.from(publicKeyRaw).toString("base64url");

  return {
    fingerprint: `fp_${encodedPublicKey.slice(0, 16)}`,
    publicKey: encodedPublicKey,
  };
}

async function requestJSON<T>(
  request: APIRequestContext,
  input: {
    accessToken?: string;
    body?: unknown;
    method?: "GET" | "PATCH" | "POST" | "PUT";
    path: string;
  },
) {
  const response = await request.fetch(`${gatewayBaseURL}${input.path}`, {
    data: input.body,
    headers: {
      ...(input.accessToken ? { Authorization: `Bearer ${input.accessToken}` } : {}),
    },
    method: input.method ?? "GET",
  });
  const payload = (await response.json()) as {
    data: T;
    error: { code: string; userMessage: string } | null;
    meta: { requestId: string; timestamp: string };
    success: boolean;
  };

  return {
    payload,
    response,
  };
}

export async function adminLogin(request: APIRequestContext, username: string, password: string) {
  const { payload, response } = await requestJSON<LoginPayload>(request, {
    body: { password, username },
    method: "POST",
    path: "/api/v1/admin/auth/login",
  });

  expect(response.ok()).toBeTruthy();
  expect(payload.success).toBeTruthy();

  return {
    meta: payload.meta,
    session: payload.data,
  };
}

export async function createAdminUser(
  request: APIRequestContext,
  accessToken: string,
  input: {
    displayName: string;
    email: string;
    password: string;
    roleIds: string[];
    username: string;
  },
) {
  const { payload, response } = await requestJSON<{
    displayName: string;
    email: string;
    id: string;
    roles: string[];
    status: string;
    username: string;
  }>(request, {
    accessToken,
    body: input,
    method: "POST",
    path: "/api/v1/admin/users",
  });

  expect(response.status()).toBe(201);
  expect(payload.success).toBeTruthy();

  return payload.data;
}

export async function createAdminRole(
  request: APIRequestContext,
  accessToken: string,
  input: {
    description: string;
    displayName: string;
    name: string;
  },
) {
  const { payload, response } = await requestJSON<{
    description: string;
    displayName: string;
    id: string;
    name: string;
  }>(request, {
    accessToken,
    body: input,
    method: "POST",
    path: "/api/v1/admin/roles",
  });

  expect(response.status()).toBe(201);
  expect(payload.success).toBeTruthy();

  return payload.data;
}

export async function replaceRoleServices(
  request: APIRequestContext,
  accessToken: string,
  roleID: string,
  serviceIDs: string[],
) {
  const { payload, response } = await requestJSON<{ roleId: string; serviceIds: string[] }>(
    request,
    {
      accessToken,
      body: { serviceIds: serviceIDs },
      method: "PUT",
      path: `/api/v1/admin/roles/${roleID}/services`,
    },
  );

  expect(response.ok()).toBeTruthy();
  expect(payload.success).toBeTruthy();

  return payload.data;
}

export async function setAdminServiceStatus(
  request: APIRequestContext,
  accessToken: string,
  serviceID: string,
  status: "disabled" | "enabled",
) {
  const { payload, response } = await requestJSON<{
    id: string;
    status: string;
  }>(request, {
    accessToken,
    body: { status },
    method: "POST",
    path: `/api/v1/admin/services/${serviceID}/status`,
  });

  expect(response.ok()).toBeTruthy();
  expect(payload.success).toBeTruthy();

  return payload.data;
}

export async function updateAdminService(
  request: APIRequestContext,
  accessToken: string,
  serviceID: string,
  input: {
    description: string;
    group: string;
    name: string;
    protocol: string;
    publicPath: string;
    upstreamUrl: string;
  },
) {
  const { payload, response } = await requestJSON<{
    id: string;
    key: string;
    upstreamUrl: string;
  }>(request, {
    accessToken,
    body: input,
    method: "PATCH",
    path: `/api/v1/admin/services/${serviceID}`,
  });

  expect(response.ok()).toBeTruthy();
  expect(payload.success).toBeTruthy();

  return payload.data;
}

export async function listAdminUsers(
  request: APIRequestContext,
  accessToken: string,
  keyword = "",
) {
  const response = await request.get(
    `${gatewayBaseURL}/api/v1/admin/users?keyword=${encodeURIComponent(keyword)}`,
    {
      headers: {
        Authorization: `Bearer ${accessToken}`,
      },
    },
  );
  const payload = (await response.json()) as {
    data: {
      items: Array<{
        displayName: string;
        email: string;
        id: string;
        roles: string[];
        status: string;
        username: string;
      }>;
    };
    success: boolean;
  };

  expect(response.ok()).toBeTruthy();
  expect(payload.success).toBeTruthy();

  return payload.data.items;
}

export async function bootstrapClientDevice(
  request: APIRequestContext,
  username: string,
  password: string,
) {
  const identity = generateDeviceIdentity();
  const { payload, response } = await requestJSON<
    LoginPayload & { device: { deviceId: string; status: string } }
  >(request, {
    body: {
      clientVersion: "0.1.0-e2e",
      deviceName: "Playwright Device",
      deviceOs: process.platform,
      password,
      publicKey: identity.publicKey,
      publicKeyFingerprint: identity.fingerprint,
      username,
    },
    method: "POST",
    path: "/api/v1/client/devices/bootstrap",
  });

  expect(response.ok()).toBeTruthy();
  expect(payload.success).toBeTruthy();

  return {
    device: payload.data.device,
    identity,
    meta: payload.meta,
    session: payload.data,
  };
}

export async function listClientServices(request: APIRequestContext, accessToken: string) {
  const { payload, response } = await requestJSON<{
    items: Array<{
      accessSource: string;
      group: string;
      id: string;
      key: string;
      name: string;
      status: string;
    }>;
  }>(request, {
    accessToken,
    method: "GET",
    path: "/api/v1/client/services",
  });

  expect(response.ok()).toBeTruthy();
  expect(payload.success).toBeTruthy();

  return payload.data.items;
}

export async function createServiceAccessURL(
  request: APIRequestContext,
  accessToken: string,
  serviceID: string,
) {
  const { payload, response } = await requestJSON<{ expiresIn: number; url: string }>(request, {
    accessToken,
    method: "POST",
    path: `/api/v1/client/services/${serviceID}/access-url`,
  });

  return { payload, response };
}

export async function proxyServiceRequest(
  request: APIRequestContext,
  accessToken: string,
  serviceKey: string,
  path = "whoami",
) {
  const response = await request.get(`${gatewayBaseURL}/s/${serviceKey}/${path}`, {
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  });
  const payload = (await response.json()) as {
    error?: { code: string; userMessage: string };
    serviceKey?: string;
    serviceName?: string;
    success?: boolean;
  };

  return {
    payload,
    response,
  };
}
