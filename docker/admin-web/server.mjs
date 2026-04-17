import { readFileSync } from "node:fs";
import { createServer } from "node:http";

const port = Number.parseInt(process.env.PORT ?? "5173", 10);
const indexHtml = readFileSync(new URL("./index.html", import.meta.url), "utf8");

const server = createServer((request, response) => {
  const url = new URL(request.url ?? "/", `http://${request.headers.host ?? "localhost"}`);

  if (url.pathname === "/health") {
    response.writeHead(200, { "content-type": "application/json; charset=utf-8" });
    response.end(JSON.stringify({ ok: true, service: "admin-web" }));
    return;
  }

  response.writeHead(200, { "content-type": "text/html; charset=utf-8" });
  response.end(indexHtml);
});

server.listen(port, "0.0.0.0", () => {
  console.log(`admin-web listening on ${port}`);
});
