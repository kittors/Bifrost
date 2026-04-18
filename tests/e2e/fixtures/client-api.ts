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
    method?: "GET" | "POST";
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
