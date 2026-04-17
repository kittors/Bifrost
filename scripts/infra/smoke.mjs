const services = ["gitlab", "jenkins", "docs", "internal-admin"];
const gatewayPort = process.env.BIFROST_DEV_GATEWAY_PORT ?? "8080";
const adminPort = process.env.BIFROST_DEV_ADMIN_PORT ?? "5173";

async function assertJson(url, predicate, failureMessage) {
  const response = await fetch(url);
  if (!response.ok) {
    throw new Error(`${failureMessage}: ${response.status} ${response.statusText}`);
  }

  const payload = await response.json();
  if (!predicate(payload)) {
    throw new Error(`${failureMessage}: ${JSON.stringify(payload)}`);
  }
}

async function main() {
  await assertJson(
    `http://127.0.0.1:${gatewayPort}/healthz`,
    (payload) => payload.status === "ok",
    "gateway health check failed",
  );

  for (const serviceKey of services) {
    await assertJson(
      `http://127.0.0.1:${gatewayPort}/debug/upstreams/${serviceKey}`,
      (payload) => payload.serviceKey === serviceKey && payload.upstream?.serviceKey,
      `gateway upstream probe failed for ${serviceKey}`,
    );
  }

  const adminResponse = await fetch(`http://127.0.0.1:${adminPort}/`);
  if (!adminResponse.ok) {
    throw new Error(`admin-web not reachable: ${adminResponse.status} ${adminResponse.statusText}`);
  }

  const adminHtml = await adminResponse.text();
  if (!adminHtml.includes("Bifrost Admin Preview")) {
    throw new Error("admin-web entry did not return the preview page");
  }
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
