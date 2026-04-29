import { randomBytes } from "node:crypto";
import { request as httpRequest } from "node:http";

import { expect, test } from "@playwright/test";

import { bootstrapClientDevice } from "./fixtures/client-api";
import { gatewayBaseURL, seedPassword } from "./fixtures/env";

test("websocket proxy upgrades to the configured upstream service", async ({ request }) => {
  const client = await bootstrapClientDevice(request, "alice", seedPassword);
  const targetURL = new URL("/s/docs/socket", gatewayBaseURL);

  const upgraded = await new Promise<{
    headers: Record<string, string | string[] | undefined>;
    statusCode: number;
  }>((resolve, reject) => {
    const request = httpRequest(targetURL, {
      headers: {
        Authorization: `Bearer ${client.session.accessToken}`,
        Connection: "Upgrade",
        "Sec-WebSocket-Key": randomBytes(16).toString("base64"),
        "Sec-WebSocket-Version": "13",
        Upgrade: "websocket",
      },
      method: "GET",
    });

    request.on("upgrade", (response, socket) => {
      socket.destroy();
      resolve({
        headers: response.headers,
        statusCode: response.statusCode ?? 0,
      });
    });
    request.on("response", (response) => {
      reject(new Error(`expected websocket upgrade, got ${response.statusCode}`));
    });
    request.on("error", reject);
    request.end();
  });

  expect(upgraded.statusCode).toBe(101);
  expect(upgraded.headers["x-mock-service-key"]).toBe("docs");
  expect(upgraded.headers["x-mock-service-path"]).toBe("/socket");
});
