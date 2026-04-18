import { requestJSON } from "../../shared/lib/http";

export async function clientLogin(input: {
  baseURL: string;
  clientVersion: string;
  deviceId: string;
  password: string;
  username: string;
}) {
  return requestJSON<{
    accessToken: string;
    refreshToken: string;
    expiresIn: number;
    user: {
      displayName: string;
      id: string;
      roles: string[];
      username: string;
    };
  }>({
    baseURL: input.baseURL,
    body: {
      clientVersion: input.clientVersion,
      deviceId: input.deviceId,
      password: input.password,
      username: input.username,
    },
    method: "POST",
    path: "/api/v1/client/auth/login",
  });
}

export async function refreshClientSession(input: {
  baseURL: string;
  deviceId: string;
  refreshToken: string;
}) {
  return requestJSON<{
    accessToken: string;
    refreshToken: string;
    expiresIn: number;
    user: {
      displayName: string;
      id: string;
      roles: string[];
      username: string;
    };
  }>({
    baseURL: input.baseURL,
    body: {
      deviceId: input.deviceId,
      refreshToken: input.refreshToken,
    },
    method: "POST",
    path: "/api/v1/client/auth/refresh",
  });
}

export async function logoutClientSession(input: { accessToken: string; baseURL: string }) {
  return requestJSON({
    accessToken: input.accessToken,
    baseURL: input.baseURL,
    method: "POST",
    path: "/api/v1/client/auth/logout",
  });
}
