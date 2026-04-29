import { requestJSON } from "../../shared/lib/http";
import type { ClientService } from "./types";

export async function listClientServices(input: {
  accessToken: string;
  baseURL: string;
  group?: string;
  keyword?: string;
}) {
  const data = await requestJSON<{ items: ClientService[] }>({
    accessToken: input.accessToken,
    baseURL: input.baseURL,
    path: "/api/v1/client/services",
    query: {
      group: input.group,
      keyword: input.keyword,
    },
  });

  return data.items;
}

export async function createServiceAccessURL(input: {
  accessToken: string;
  baseURL: string;
  serviceId: string;
}) {
  return requestJSON<{
    accessTicket: string;
    expiresIn: number;
    url?: string;
    publicPath?: string;
  }>({
    accessToken: input.accessToken,
    baseURL: input.baseURL,
    method: "POST",
    path: `/api/v1/client/services/${input.serviceId}/access-url`,
  });
}
