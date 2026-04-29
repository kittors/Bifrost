import { once } from "node:events";
import {
  createServer,
  type IncomingHttpHeaders,
  type IncomingMessage,
  type Server,
  type ServerResponse,
} from "node:http";
import { connect as connectTCP } from "node:net";
import type { Duplex } from "node:stream";
import { connect as connectTLS } from "node:tls";

import type { DesktopLocalProxyStatus, DesktopSessionSnapshot } from "../shared/types";

type LocalProxyControllerOptions = {
  maxPort?: number;
  preferredPort?: number;
};

const loopbackHost = "127.0.0.1" as const;

const hopByHopHeaders = new Set([
  "authorization",
  "connection",
  "content-length",
  "host",
  "keep-alive",
  "proxy-authenticate",
  "proxy-authorization",
  "te",
  "trailer",
  "transfer-encoding",
  "upgrade",
]);

async function readRequestBody(request: IncomingMessage) {
  const chunks: Buffer[] = [];
  for await (const chunk of request) {
    chunks.push(Buffer.isBuffer(chunk) ? chunk : Buffer.from(chunk));
  }
  return Buffer.concat(chunks);
}

function buildForwardHeaders(headers: IncomingHttpHeaders, accessToken: string) {
  const forwardedHeaders = new Headers();

  for (const [key, value] of Object.entries(headers)) {
    if (hopByHopHeaders.has(key.toLowerCase()) || value === undefined) {
      continue;
    }

    if (Array.isArray(value)) {
      for (const item of value) {
        forwardedHeaders.append(key, item);
      }
      continue;
    }

    forwardedHeaders.set(key, value);
  }

  forwardedHeaders.set("Authorization", `Bearer ${accessToken}`);
  forwardedHeaders.set("X-Bifrost-Local-Proxy", "true");

  return forwardedHeaders;
}

function writeForwardedResponse(response: ServerResponse, upstreamResponse: Response) {
  for (const [key, value] of upstreamResponse.headers.entries()) {
    if (hopByHopHeaders.has(key.toLowerCase())) {
      continue;
    }
    response.setHeader(key, value);
  }

  response.statusCode = upstreamResponse.status;
}

function isSafeRequestMethod(method: string | undefined) {
  return method === "GET" || method === "HEAD" || method === "OPTIONS";
}

function isSameLocalOrigin(value: string, baseURL: string) {
  try {
    return new URL(value).origin === new URL(baseURL).origin;
  } catch {
    return false;
  }
}

function hasForeignBrowserOrigin(headers: IncomingHttpHeaders, baseURL: string) {
  const origin = Array.isArray(headers.origin) ? headers.origin[0] : headers.origin;
  if (origin && !isSameLocalOrigin(origin, baseURL)) {
    return true;
  }

  const referer = Array.isArray(headers.referer) ? headers.referer[0] : headers.referer;
  if (referer && !isSameLocalOrigin(referer, baseURL)) {
    return true;
  }

  return false;
}

function defaultPort(protocol: string) {
  return protocol === "https:" ? 443 : 80;
}

function buildGatewayURL(requestURL: string, gatewayBaseURL: string) {
  return new URL(requestURL, `${gatewayBaseURL.replace(/\/+$/, "")}/`);
}

function writeSocketJSONError(socket: Duplex, statusCode: number, message: string) {
  const payload = JSON.stringify({ error: message });
  socket.end(
    [
      `HTTP/1.1 ${statusCode} ${message}`,
      "Content-Type: application/json",
      `Content-Length: ${Buffer.byteLength(payload)}`,
      "Connection: close",
      "",
      payload,
    ].join("\r\n"),
  );
}

function buildUpgradeRequestHeaders(
  headers: IncomingHttpHeaders,
  gatewayURL: URL,
  accessToken: string,
) {
  const forwardedHeaders = new Map<string, string>();

  for (const [key, value] of Object.entries(headers)) {
    const normalizedKey = key.toLowerCase();
    if (
      value === undefined ||
      normalizedKey === "authorization" ||
      normalizedKey === "host" ||
      normalizedKey === "proxy-connection" ||
      normalizedKey === "proxy-authenticate" ||
      normalizedKey === "proxy-authorization"
    ) {
      continue;
    }

    const normalizedValue = Array.isArray(value) ? value.join(", ") : value;
    forwardedHeaders.set(key, normalizedValue);
  }

  forwardedHeaders.set("Host", gatewayURL.host);
  forwardedHeaders.set("Authorization", `Bearer ${accessToken}`);
  forwardedHeaders.set("X-Bifrost-Local-Proxy", "true");

  return forwardedHeaders;
}

async function connectGatewaySocket(gatewayURL: URL): Promise<Duplex> {
  const port = Number(gatewayURL.port || defaultPort(gatewayURL.protocol));

  if (gatewayURL.protocol === "https:") {
    const socket = connectTLS({
      host: gatewayURL.hostname,
      port,
      servername: gatewayURL.hostname,
    });
    await Promise.race([
      once(socket, "secureConnect"),
      once(socket, "error").then(([error]) => {
        throw error;
      }),
    ]);
    return socket;
  }

  const socket = connectTCP({
    host: gatewayURL.hostname,
    port,
  });
  await Promise.race([
    once(socket, "connect"),
    once(socket, "error").then(([error]) => {
      throw error;
    }),
  ]);
  return socket;
}

export function createLocalProxyController(options: LocalProxyControllerOptions = {}) {
  const preferredPort = options.preferredPort ?? 18080;
  const maxPort = options.maxPort ?? 18099;

  let activeServer: Server | null = null;
  let activeSession: DesktopSessionSnapshot | null = null;
  let activeStatus: DesktopLocalProxyStatus | null = null;
  const activeSockets = new Set<Duplex>();

  async function handleRequest(request: IncomingMessage, response: ServerResponse) {
    if (!activeSession || !activeStatus) {
      response.writeHead(503, { "Content-Type": "application/json" });
      response.end(JSON.stringify({ error: "local proxy is not ready" }));
      return;
    }

    const requestURL = request.url ?? "/";
    if (!requestURL.startsWith("/s/")) {
      response.writeHead(404, { "Content-Type": "application/json" });
      response.end(JSON.stringify({ error: "local proxy route not found" }));
      return;
    }

    // 浏览器跨站写操作不能借本地代理带设备会话；curl 等本地工具通常不带 Origin/Referer。
    if (
      !isSafeRequestMethod(request.method) &&
      hasForeignBrowserOrigin(request.headers, activeStatus.baseURL)
    ) {
      response.writeHead(403, { "Content-Type": "application/json" });
      response.end(JSON.stringify({ error: "foreign browser origin is not allowed" }));
      return;
    }

    const gatewayURL = buildGatewayURL(requestURL, activeSession.gatewayBaseURL);
    const requestBody =
      request.method === "GET" || request.method === "HEAD"
        ? undefined
        : await readRequestBody(request);

    const upstreamResponse = await fetch(gatewayURL, {
      body: requestBody,
      headers: buildForwardHeaders(request.headers, activeSession.accessToken),
      method: request.method ?? "GET",
      redirect: "manual",
    });

    writeForwardedResponse(response, upstreamResponse);
    const payload = Buffer.from(await upstreamResponse.arrayBuffer());
    response.end(payload);
  }

  async function handleUpgrade(request: IncomingMessage, socket: Duplex, head: Buffer) {
    if (!activeSession || !activeStatus) {
      writeSocketJSONError(socket, 503, "local proxy is not ready");
      return;
    }

    const requestURL = request.url ?? "/";
    if (!requestURL.startsWith("/s/")) {
      writeSocketJSONError(socket, 404, "local proxy route not found");
      return;
    }

    const gatewayURL = buildGatewayURL(requestURL, activeSession.gatewayBaseURL);

    try {
      const upstreamSocket = await connectGatewaySocket(gatewayURL);

      // upgrade 请求需要保留握手头，但鉴权必须在主进程中覆盖注入。
      const requestLines = [
        `${request.method ?? "GET"} ${gatewayURL.pathname}${gatewayURL.search} HTTP/1.1`,
        ...Array.from(
          buildUpgradeRequestHeaders(request.headers, gatewayURL, activeSession.accessToken),
          ([key, value]) => `${key}: ${value}`,
        ),
        "",
        "",
      ];

      upstreamSocket.write(requestLines.join("\r\n"));
      if (head.length > 0) {
        upstreamSocket.write(head);
      }

      upstreamSocket.pipe(socket);
      socket.pipe(upstreamSocket);

      const closeSockets = () => {
        socket.destroy();
        upstreamSocket.destroy();
      };

      socket.once("error", closeSockets);
      upstreamSocket.once("error", closeSockets);
      socket.once("close", () => upstreamSocket.destroy());
      upstreamSocket.once("close", () => socket.destroy());
    } catch (error) {
      writeSocketJSONError(
        socket,
        502,
        error instanceof Error ? error.message : "local proxy upstream request failed",
      );
    }
  }

  async function bindServer(port: number) {
    const server = createServer((request, response) => {
      void handleRequest(request, response).catch((error: unknown) => {
        response.writeHead(502, { "Content-Type": "application/json" });
        response.end(
          JSON.stringify({
            error: error instanceof Error ? error.message : "local proxy upstream request failed",
          }),
        );
      });
    });
    server.on("connection", (socket) => {
      activeSockets.add(socket);
      socket.once("close", () => {
        activeSockets.delete(socket);
      });
    });
    server.on("upgrade", (request, socket, head) => {
      void handleUpgrade(request, socket, head).catch(() => {
        socket.destroy();
      });
    });

    server.listen(port, loopbackHost);

    try {
      await Promise.race([
        once(server, "listening"),
        once(server, "error").then(([error]) => {
          throw error;
        }),
      ]);
      return server;
    } catch (error) {
      server.close();
      throw error;
    }
  }

  return {
    async openService(publicPath: string) {
      if (!activeStatus) {
        throw new Error("local proxy is not running");
      }

      const normalizedPath = publicPath.startsWith("/") ? publicPath : `/${publicPath}`;
      if (!normalizedPath.startsWith("/s/")) {
        throw new Error("local proxy only opens /s service paths");
      }
      return new URL(normalizedPath.replace(/\/?$/, "/"), `${activeStatus.baseURL}/`).toString();
    },
    async start(session: DesktopSessionSnapshot) {
      if (
        activeServer &&
        activeStatus &&
        activeSession?.gatewayBaseURL === session.gatewayBaseURL
      ) {
        activeSession = session;
        return activeStatus;
      }

      await this.stop();

      for (let port = preferredPort; port <= maxPort; port += 1) {
        try {
          const server = await bindServer(port);
          activeServer = server;
          activeSession = session;
          activeStatus = {
            baseURL: `http://${loopbackHost}:${port}`,
            host: loopbackHost,
            port,
            running: true,
          };
          return activeStatus;
        } catch (error) {
          if (!(error instanceof Error) || !("code" in error) || error.code !== "EADDRINUSE") {
            throw error;
          }
        }
      }

      throw new Error(`no available local proxy port in range ${preferredPort}-${maxPort}`);
    },
    status() {
      return (
        activeStatus ?? {
          baseURL: "",
          host: loopbackHost,
          port: 0,
          running: false,
        }
      );
    },
    async stop() {
      if (!activeServer) {
        activeSession = null;
        activeStatus = null;
        return;
      }

      const server = activeServer;
      activeServer = null;
      activeSession = null;
      activeStatus = null;
      for (const socket of activeSockets) {
        socket.destroy();
      }
      activeSockets.clear();

      await new Promise<void>((resolve, reject) => {
        server.close((error) => {
          if (error) {
            reject(error);
            return;
          }
          resolve();
        });
      });
    },
  };
}
