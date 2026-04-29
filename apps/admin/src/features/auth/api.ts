import { getGatewayBaseURL } from "../../shared/config/env";
import { requestJSON } from "../../shared/lib/http";
import type { StoredAdminSession } from "./session";

type LoginResponse = {
  accessToken: string;
  expiresIn: number;
  refreshToken: string;
  user: StoredAdminSession["user"];
};

export async function adminLogin(input: { password: string; username: string }) {
  const payload = await requestJSON<LoginResponse>({
    baseURL: getGatewayBaseURL(),
    body: input,
    method: "POST",
    path: "/api/v1/admin/auth/login",
  });

  const expiresAt = new Date(Date.now() + payload.data.expiresIn * 1000).toISOString();

  return {
    accessToken: payload.data.accessToken,
    expiresAt,
    refreshToken: payload.data.refreshToken,
    user: payload.data.user,
  } satisfies StoredAdminSession;
}

export async function adminLogout(accessToken: string) {
  return requestJSON({
    accessToken,
    baseURL: getGatewayBaseURL(),
    method: "POST",
    path: "/api/v1/admin/auth/logout",
  });
}
