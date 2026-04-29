const gatewayBaseURL = (
  process.env.BIFROST_REMOTE_DEV_GATEWAY_URL ?? "http://142.171.208.80:18080"
).replace(/\/+$/, "");

const checks = [
  { name: "网关健康检查", path: "/healthz" },
  { name: "网关就绪检查", path: "/readyz" },
  { name: "通过网关访问私有 GitLab 上游", path: "/debug/upstreams/gitlab" },
  { name: "通过网关访问私有 Jenkins 上游", path: "/debug/upstreams/jenkins" },
  { name: "通过网关访问私有 Docs 上游", path: "/debug/upstreams/docs" },
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
    throw new Error(`${check.name}失败：${response.status} ${response.statusText}（${url}）`);
  }

  return response.json();
}

for (const check of checks) {
  const payload = await checkEndpoint(check);
  console.log(`成功：${check.name}：${JSON.stringify(payload)}`);
}

console.log("");
console.log(`远端 dev 后端已就绪：${gatewayBaseURL}`);
console.log("注意：这条命令只检查远端后端，不会启动后台管理页面。");
console.log("启动后台管理页面请执行：pnpm dev:admin");
console.log(`后台页面会通过 VITE_GATEWAY_BASE_URL=${gatewayBaseURL} 访问后端接口。`);
