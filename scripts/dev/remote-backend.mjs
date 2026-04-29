const gatewayBaseURL = (
  process.env.BIFROST_REMOTE_DEV_GATEWAY_URL ?? "http://142.171.208.80:18080"
).replace(/\/+$/, "");

const checks = [
  { name: "Gateway health", path: "/healthz" },
  { name: "Gateway readiness", path: "/readyz" },
  { name: "Private GitLab upstream via Gateway", path: "/debug/upstreams/gitlab" },
  { name: "Private Jenkins upstream via Gateway", path: "/debug/upstreams/jenkins" },
  { name: "Private Docs upstream via Gateway", path: "/debug/upstreams/docs" },
];

async function checkEndpoint(check) {
  const url = `${gatewayBaseURL}${check.path}`;
  const response = await fetch(url, {
    headers: {
      accept: "application/json",
    },
    signal: AbortSignal.timeout(5000),
  });

  if (!response.ok) {
    throw new Error(`${check.name} failed: ${response.status} ${response.statusText} (${url})`);
  }

  return response.json();
}

for (const check of checks) {
  const payload = await checkEndpoint(check);
  console.log(`ok ${check.name}: ${JSON.stringify(payload)}`);
}

console.log("");
console.log(`Remote dev backend is ready: ${gatewayBaseURL}`);
console.log(`Use VITE_GATEWAY_BASE_URL=${gatewayBaseURL} for local Admin Web.`);
