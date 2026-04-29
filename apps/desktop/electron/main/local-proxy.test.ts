// @vitest-environment node

import { once } from "node:events";
import { createServer, type IncomingMessage, type Server, type ServerResponse } from "node:http";
import { createConnection, type Socket } from "node:net";
import { afterEach, describe, expect, it } from "vitest";

import type { DesktopSessionSnapshot } from "../shared/types";
import { createLocalProxyController } from "./local-proxy";

type RequestSnapshot = {
  authorization: string;
  method: string;
  url: string;
};

async function startHTTPServer(
  handler: (request: IncomingMessage, response: ServerResponse) => void,
) {
  const server = createServer(handler);
  server.listen(0, "127.0.0.1");
  await once(server, "listening");
  return server;
}

function serverOrigin(server: Server) {
  const address = server.address();
  if (!address || typeof address === "string") {
    throw new Error("server address is not available");
  }
  return `http://${address.address}:${address.port}`;
}

function createSession(baseURL: string): DesktopSessionSnapshot {
  return {
    accessToken: "access_local_proxy",
    deviceId: "device_local_proxy",
    expiresAt: new Date(Date.now() + 15 * 60 * 1000).toISOString(),
    gatewayBaseURL: baseURL,
    refreshToken: "refresh_local_proxy",
    user: {
      displayName: "Alice",
      id: "user_alice",
      roles: ["role_developer"],
      username: "alice",
    },
  };
}

async function readSocketData(socket: Socket) {
  const [chunk] = await Promise.race([
    once(socket, "data"),
    new Promise<never>((_resolve, reject) => {
      setTimeout(() => reject(new Error("timed out waiting for socket response")), 500);
    }),
  ]);
  return Buffer.isBuffer(chunk) ? chunk.toString("utf8") : String(chunk);
}

describe("createLocalProxyController", () => {
  const cleanup: Array<() => Promise<void>> = [];

  afterEach(async () => {
    while (cleanup.length > 0) {
      const dispose = cleanup.pop();
      if (dispose) {
        await dispose();
      }
    }
  });

  it("forwards /s routes to gateway and injects bearer token", async () => {
    let requestSnapshot: RequestSnapshot | null = null;
    const gateway = await startHTTPServer((request, response) => {
      requestSnapshot = {
        authorization: String(request.headers.authorization ?? ""),
        method: String(request.method ?? ""),
        url: String(request.url ?? ""),
      };
      response.writeHead(200, { "Content-Type": "application/json" });
      response.end(JSON.stringify({ ok: true, upstream: "gitlab" }));
    });
    cleanup.push(
      async () =>
        await new Promise<void>((resolve, reject) => {
          gateway.close((error) => (error ? reject(error) : resolve()));
        }),
    );

    const controller = createLocalProxyController({
      maxPort: 18189,
      preferredPort: 18180,
    });
    cleanup.push(async () => {
      await controller.stop();
    });

    const status = await controller.start(createSession(serverOrigin(gateway)));

    expect(status.running).toBe(true);
    expect(status.host).toBe("127.0.0.1");
    expect(status.port).toBeGreaterThanOrEqual(18180);
    expect(status.port).toBeLessThanOrEqual(18189);

    const response = await fetch(`${status.baseURL}/s/gitlab/whoami?from=browser`);
    expect(response.status).toBe(200);
    await expect(response.json()).resolves.toEqual({ ok: true, upstream: "gitlab" });
    expect(requestSnapshot).toEqual({
      authorization: "Bearer access_local_proxy",
      method: "GET",
      url: "/s/gitlab/whoami?from=browser",
    });
  });

  it("rejects non service routes without touching gateway", async () => {
    let gatewayTouched = false;
    const gateway = await startHTTPServer((_request, response) => {
      gatewayTouched = true;
      response.writeHead(200).end();
    });
    cleanup.push(
      async () =>
        await new Promise<void>((resolve, reject) => {
          gateway.close((error) => (error ? reject(error) : resolve()));
        }),
    );

    const controller = createLocalProxyController({
      maxPort: 18199,
      preferredPort: 18190,
    });
    cleanup.push(async () => {
      await controller.stop();
    });

    const status = await controller.start(createSession(serverOrigin(gateway)));
    const response = await fetch(`${status.baseURL}/gitlab/whoami`);

    expect(response.status).toBe(404);
    expect(gatewayTouched).toBe(false);
  });

  it("blocks unsafe requests from foreign browser origins", async () => {
    let gatewayTouched = false;
    const gateway = await startHTTPServer((_request, response) => {
      gatewayTouched = true;
      response.writeHead(200, { "Content-Type": "application/json" });
      response.end(JSON.stringify({ ok: true }));
    });
    cleanup.push(
      async () =>
        await new Promise<void>((resolve, reject) => {
          gateway.close((error) => (error ? reject(error) : resolve()));
        }),
    );

    const controller = createLocalProxyController({
      maxPort: 18209,
      preferredPort: 18200,
    });
    cleanup.push(async () => {
      await controller.stop();
    });

    const status = await controller.start(createSession(serverOrigin(gateway)));
    const response = await fetch(`${status.baseURL}/s/gitlab/api/v4/projects`, {
      body: JSON.stringify({ name: "blocked" }),
      headers: {
        "Content-Type": "application/json",
        Origin: "https://evil.example.com",
      },
      method: "POST",
    });

    expect(response.status).toBe(403);
    expect(gatewayTouched).toBe(false);
  });

  it("proxies websocket upgrade requests and injects bearer token", async () => {
    let requestSnapshot: RequestSnapshot | null = null;
    const gateway = createServer();
    gateway.on("upgrade", (request, socket) => {
      requestSnapshot = {
        authorization: String(request.headers.authorization ?? ""),
        method: String(request.method ?? ""),
        url: String(request.url ?? ""),
      };
      socket.write(
        "HTTP/1.1 101 Switching Protocols\r\nConnection: Upgrade\r\nUpgrade: websocket\r\n\r\n",
      );
      socket.end();
    });
    gateway.listen(0, "127.0.0.1");
    await once(gateway, "listening");
    cleanup.push(
      async () =>
        await new Promise<void>((resolve, reject) => {
          gateway.close((error) => (error ? reject(error) : resolve()));
        }),
    );

    const controller = createLocalProxyController({
      maxPort: 18219,
      preferredPort: 18210,
    });
    cleanup.push(async () => {
      await controller.stop();
    });

    const status = await controller.start(createSession(serverOrigin(gateway)));
    const socket = createConnection({ host: "127.0.0.1", port: status.port });
    cleanup.push(async () => {
      socket.destroy();
    });
    await once(socket, "connect");

    socket.write(
      [
        "GET /s/gitlab/-/cable HTTP/1.1",
        `Host: 127.0.0.1:${status.port}`,
        "Connection: Upgrade",
        "Upgrade: websocket",
        "Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==",
        "Sec-WebSocket-Version: 13",
        "",
        "",
      ].join("\r\n"),
    );

    const responseText = await readSocketData(socket);

    expect(responseText).toContain("101 Switching Protocols");
    expect(requestSnapshot).toEqual({
      authorization: "Bearer access_local_proxy",
      method: "GET",
      url: "/s/gitlab/-/cable",
    });
  });

  it("falls back to the next port when the preferred port is busy", async () => {
    const blocker = await startHTTPServer((_request, response) => {
      response.writeHead(204).end();
    });
    const blockerAddress = blocker.address();
    if (!blockerAddress || typeof blockerAddress === "string") {
      throw new Error("blocker address is unavailable");
    }

    cleanup.push(
      async () =>
        await new Promise<void>((resolve, reject) => {
          blocker.close((error) => (error ? reject(error) : resolve()));
        }),
    );

    const gateway = await startHTTPServer((_request, response) => {
      response.writeHead(200, { "Content-Type": "application/json" });
      response.end(JSON.stringify({ ok: true }));
    });
    cleanup.push(
      async () =>
        await new Promise<void>((resolve, reject) => {
          gateway.close((error) => (error ? reject(error) : resolve()));
        }),
    );

    const controller = createLocalProxyController({
      maxPort: blockerAddress.port + 3,
      preferredPort: blockerAddress.port,
    });
    cleanup.push(async () => {
      await controller.stop();
    });

    const status = await controller.start(createSession(serverOrigin(gateway)));

    expect(status.port).toBeGreaterThan(blockerAddress.port);
    expect(status.port).toBeLessThanOrEqual(blockerAddress.port + 3);
  });
});
