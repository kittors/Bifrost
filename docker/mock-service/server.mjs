import { createServer } from "node:http";

const port = Number.parseInt(process.env.PORT ?? "8080", 10);
const serviceName = process.env.SERVICE_NAME ?? "mock-service";
const serviceKey = process.env.SERVICE_KEY ?? "mock";

function sendJson(response, statusCode, payload) {
  response.writeHead(statusCode, {
    "content-type": "application/json; charset=utf-8",
  });
  response.end(JSON.stringify(payload));
}

const server = createServer(async (request, response) => {
  const url = new URL(request.url ?? "/", `http://${request.headers.host ?? "localhost"}`);

  if (url.pathname === "/health") {
    sendJson(response, 200, { ok: true, serviceKey, serviceName });
    return;
  }

  if (url.pathname === "/whoami") {
    sendJson(response, 200, { serviceKey, serviceName, path: url.pathname });
    return;
  }

  if (url.pathname === "/headers") {
    sendJson(response, 200, {
      headers: request.headers,
      serviceKey,
      serviceName,
    });
    return;
  }

  if (url.pathname === "/slow") {
    const delayMs = Number.parseInt(url.searchParams.get("delayMs") ?? "1500", 10);
    await new Promise((resolve) => setTimeout(resolve, delayMs));
    sendJson(response, 200, { delayed: true, delayMs, serviceKey, serviceName });
    return;
  }

  if (url.pathname === "/echo" && request.method === "POST") {
    let body = "";
    request.setEncoding("utf8");
    request.on("data", (chunk) => {
      body += chunk;
    });
    request.on("end", () => {
      sendJson(response, 200, {
        body,
        method: request.method,
        serviceKey,
        serviceName,
      });
    });
    return;
  }

  sendJson(response, 200, {
    ok: true,
    serviceKey,
    serviceName,
    path: url.pathname,
  });
});

server.listen(port, "0.0.0.0", () => {
  console.log(`${serviceName} listening on ${port}`);
});
