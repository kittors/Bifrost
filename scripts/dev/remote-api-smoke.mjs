import { generateKeyPairSync } from "node:crypto";

const gatewayBaseURL = (
  process.env.BIFROST_REMOTE_DEV_GATEWAY_URL ?? "http://142.171.208.80:18080"
).replace(/\/+$/, "");

const requestTimeoutMS = Number.parseInt(process.env.BIFROST_REMOTE_DEV_TIMEOUT_MS ?? "10000", 10);
const seedPassword = process.env.BIFROST_REMOTE_DEV_SEED_PASSWORD ?? "ChangeMe123!";

function generateDeviceIdentity() {
  const { publicKey } = generateKeyPairSync("ed25519");
  const publicKeyDER = publicKey.export({ format: "der", type: "spki" });
  const publicKeyRaw = Buffer.from(publicKeyDER).subarray(-32);
  const encodedPublicKey = Buffer.from(publicKeyRaw).toString("base64url");

  return {
    fingerprint: `smoke_${Date.now()}_${encodedPublicKey.slice(0, 16)}`,
    publicKey: encodedPublicKey,
  };
}

function assertCondition(condition, message) {
  if (!condition) {
    throw new Error(message);
  }
}

function buildURL(pathOrURL) {
  if (pathOrURL.startsWith("http://") || pathOrURL.startsWith("https://")) {
    return pathOrURL;
  }

  return `${gatewayBaseURL}${pathOrURL.startsWith("/") ? pathOrURL : `/${pathOrURL}`}`;
}

function describePayload(payload) {
  if (!payload || typeof payload !== "object") {
    return "";
  }

  const requestID = payload.meta?.requestId ? ` requestId=${payload.meta.requestId}` : "";
  const error = payload.error
    ? ` code=${payload.error.code ?? "-"} userMessage=${payload.error.userMessage ?? "-"}`
    : "";

  return `${requestID}${error}`;
}

async function requestJSON(name, input) {
  const url = buildURL(input.path);
  const headers = {
    accept: "application/json",
    ...(input.body === undefined ? {} : { "content-type": "application/json" }),
    ...(input.accessToken ? { authorization: `Bearer ${input.accessToken}` } : {}),
    ...input.headers,
  };
  const response = await fetch(url, {
    body: input.body === undefined ? undefined : JSON.stringify(input.body),
    headers,
    method: input.method ?? "GET",
    signal: AbortSignal.timeout(requestTimeoutMS),
  });
  const rawBody = await response.text();
  let payload = null;

  if (rawBody !== "") {
    try {
      payload = JSON.parse(rawBody);
    } catch (error) {
      throw new Error(
        `${name} 返回的不是 JSON：${error.message}；响应体：${rawBody.slice(0, 500)}`,
      );
    }
  }

  const expectedStatus = input.expectedStatus ?? 200;
  if (response.status !== expectedStatus) {
    throw new Error(
      `${name} 状态码异常：期望 ${expectedStatus}，实际 ${response.status}。${describePayload(
        payload,
      )} 响应体：${rawBody.slice(0, 500)}`,
    );
  }

  return { payload, response };
}

async function apiSuccess(name, input, validate) {
  const result = await requestJSON(name, input);
  assertCondition(
    result.payload?.success === true,
    `${name} 未返回 success=true。${describePayload(result.payload)}`,
  );
  assertCondition(result.payload.meta?.requestId, `${name} 未返回 meta.requestId`);

  if (validate) {
    validate(result.payload.data, result);
  }

  console.log(`成功：${name}（requestId=${result.payload.meta.requestId}）`);
  return result;
}

async function jsonSuccess(name, input, validate) {
  const result = await requestJSON(name, input);

  if (validate) {
    validate(result.payload, result);
  }

  console.log(`成功：${name}`);
  return result;
}

function appendPath(baseURL, path) {
  const normalizedBase = baseURL.endsWith("/") ? baseURL : `${baseURL}/`;
  return new URL(path.replace(/^\/+/, ""), normalizedBase).toString();
}

function cookieHeaderFrom(response) {
  const cookie = response.headers.get("set-cookie");
  if (!cookie) {
    return "";
  }

  return cookie.split(";")[0];
}

async function cleanupSmokeDevice(adminAccessToken, deviceID) {
  if (!adminAccessToken || !deviceID) {
    return;
  }

  await apiSuccess("禁用 smoke 临时设备", {
    accessToken: adminAccessToken,
    body: { status: "disabled" },
    method: "POST",
    path: `/api/v1/admin/devices/${deviceID}/status`,
  });
}

async function main() {
  let adminAccessToken = "";
  let smokeDeviceID = "";

  try {
    await jsonSuccess("网关健康检查", { path: "/healthz" }, (payload) => {
      assertCondition(payload.status === "ok", `网关健康检查状态异常：${JSON.stringify(payload)}`);
    });

    await jsonSuccess("网关就绪检查", { path: "/readyz" }, (payload) => {
      assertCondition(
        payload.status === "ready",
        `网关就绪检查状态异常：${JSON.stringify(payload)}`,
      );
      assertCondition(payload.upstreams?.gitlab, "网关就绪检查未返回 gitlab upstream");
      assertCondition(payload.upstreams?.jenkins, "网关就绪检查未返回 jenkins upstream");
      assertCondition(payload.upstreams?.docs, "网关就绪检查未返回 docs upstream");
    });

    const adminLogin = await apiSuccess(
      "管理员登录",
      {
        body: { password: seedPassword, username: "admin" },
        method: "POST",
        path: "/api/v1/admin/auth/login",
      },
      (data) => {
        assertCondition(data.user?.username === "admin", "管理员登录返回的用户不是 admin");
        assertCondition(data.user?.roles?.includes("role_admin"), "管理员登录未返回 role_admin");
        assertCondition(data.accessToken, "管理员登录未返回 accessToken");
        assertCondition(data.refreshToken, "管理员登录未返回 refreshToken");
      },
    );
    adminAccessToken = adminLogin.payload.data.accessToken;

    await apiSuccess(
      "管理员当前用户",
      {
        accessToken: adminAccessToken,
        path: "/api/v1/admin/auth/me",
      },
      (data) => {
        assertCondition(data.user?.username === "admin", "管理员当前用户返回异常");
      },
    );

    const refreshedAdmin = await apiSuccess(
      "管理员刷新会话",
      {
        body: { refreshToken: adminLogin.payload.data.refreshToken },
        method: "POST",
        path: "/api/v1/admin/auth/refresh",
      },
      (data) => {
        assertCondition(data.accessToken, "管理员刷新会话未返回 accessToken");
        assertCondition(data.refreshToken, "管理员刷新会话未返回 refreshToken");
      },
    );
    adminAccessToken = refreshedAdmin.payload.data.accessToken;

    await apiSuccess(
      "管理员用户列表",
      {
        accessToken: adminAccessToken,
        path: "/api/v1/admin/users?page=1&pageSize=20",
      },
      (data) => {
        const usernames = data.items?.map((user) => user.username) ?? [];
        assertCondition(usernames.includes("admin"), "管理员用户列表未包含 admin");
        assertCondition(usernames.includes("alice"), "管理员用户列表未包含 alice");
      },
    );

    await apiSuccess(
      "管理员用户详情",
      {
        accessToken: adminAccessToken,
        path: "/api/v1/admin/users/user_admin",
      },
      (data) => {
        assertCondition(data.username === "admin", "管理员用户详情返回异常");
      },
    );

    await apiSuccess(
      "管理员角色列表",
      {
        accessToken: adminAccessToken,
        path: "/api/v1/admin/roles?page=1&pageSize=20",
      },
      (data) => {
        const roleNames = data.items?.map((role) => role.name) ?? [];
        assertCondition(roleNames.includes("admin"), "管理员角色列表未包含 admin");
        assertCondition(roleNames.includes("developer"), "管理员角色列表未包含 developer");
      },
    );

    await apiSuccess(
      "管理员服务列表",
      {
        accessToken: adminAccessToken,
        path: "/api/v1/admin/services?page=1&pageSize=20",
      },
      (data) => {
        const serviceKeys = data.items?.map((service) => service.key) ?? [];
        assertCondition(serviceKeys.includes("gitlab"), "管理员服务列表未包含 gitlab");
        assertCondition(serviceKeys.includes("docs"), "管理员服务列表未包含 docs");
      },
    );

    await apiSuccess(
      "管理员服务详情",
      {
        accessToken: adminAccessToken,
        path: "/api/v1/admin/services/service_gitlab",
      },
      (data) => {
        assertCondition(data.key === "gitlab", "管理员服务详情返回异常");
      },
    );

    await apiSuccess(
      "管理员设备列表",
      {
        accessToken: adminAccessToken,
        path: "/api/v1/admin/devices?page=1&pageSize=20&userId=user_alice",
      },
      (data) => {
        assertCondition(Array.isArray(data.items), "管理员设备列表未返回 items");
      },
    );

    await apiSuccess(
      "管理员审计列表",
      {
        accessToken: adminAccessToken,
        path: "/api/v1/admin/audit-events?page=1&pageSize=20",
      },
      (data) => {
        assertCondition(Array.isArray(data.items), "管理员审计列表未返回 items");
      },
    );

    await apiSuccess(
      "管理员用户服务覆盖列表",
      {
        accessToken: adminAccessToken,
        path: "/api/v1/admin/users/user_alice/service-overrides",
      },
      (data) => {
        assertCondition(Array.isArray(data.items), "管理员用户服务覆盖列表未返回 items");
      },
    );

    const identity = generateDeviceIdentity();
    const clientBootstrap = await apiSuccess(
      "客户端设备 bootstrap",
      {
        body: {
          clientVersion: "0.1.0-remote-smoke",
          deviceName: "远端接口 smoke 临时设备",
          deviceOs: process.platform,
          password: seedPassword,
          publicKey: identity.publicKey,
          publicKeyFingerprint: identity.fingerprint,
          username: "alice",
        },
        method: "POST",
        path: "/api/v1/client/devices/bootstrap",
      },
      (data) => {
        assertCondition(data.user?.username === "alice", "客户端 bootstrap 返回的用户不是 alice");
        assertCondition(data.device?.deviceId, "客户端 bootstrap 未返回 deviceId");
        assertCondition(data.accessToken, "客户端 bootstrap 未返回 accessToken");
        assertCondition(data.refreshToken, "客户端 bootstrap 未返回 refreshToken");
      },
    );
    smokeDeviceID = clientBootstrap.payload.data.device.deviceId;
    let clientAccessToken = clientBootstrap.payload.data.accessToken;
    const clientRefreshToken = clientBootstrap.payload.data.refreshToken;

    await apiSuccess(
      "客户端当前用户",
      {
        accessToken: clientAccessToken,
        path: "/api/v1/client/me",
      },
      (data) => {
        assertCondition(data.user?.username === "alice", "客户端当前用户返回异常");
      },
    );

    const clientServices = await apiSuccess(
      "客户端服务列表",
      {
        accessToken: clientAccessToken,
        path: "/api/v1/client/services",
      },
      (data) => {
        const serviceKeys = data.items?.map((service) => service.key) ?? [];
        assertCondition(serviceKeys.includes("gitlab"), "客户端服务列表未包含 gitlab");
        assertCondition(serviceKeys.includes("docs"), "客户端服务列表未包含 docs");
        assertCondition(!serviceKeys.includes("jenkins"), "客户端服务列表不应包含 jenkins");
      },
    );

    const gitlabService = clientServices.payload.data.items.find(
      (service) => service.key === "gitlab",
    );
    assertCondition(gitlabService?.id, "客户端服务列表未找到 gitlab 服务 ID");

    await apiSuccess(
      "客户端服务详情",
      {
        accessToken: clientAccessToken,
        path: `/api/v1/client/services/${gitlabService.id}`,
      },
      (data) => {
        assertCondition(data.key === "gitlab", "客户端服务详情返回异常");
      },
    );

    const refreshedClient = await apiSuccess(
      "客户端刷新会话",
      {
        body: { deviceId: smokeDeviceID, refreshToken: clientRefreshToken },
        method: "POST",
        path: "/api/v1/client/auth/refresh",
      },
      (data) => {
        assertCondition(data.accessToken, "客户端刷新会话未返回 accessToken");
        assertCondition(data.refreshToken, "客户端刷新会话未返回 refreshToken");
      },
    );
    clientAccessToken = refreshedClient.payload.data.accessToken;

    const accessURLResult = await apiSuccess(
      "客户端创建 GitLab 访问 URL",
      {
        accessToken: clientAccessToken,
        method: "POST",
        path: `/api/v1/client/services/${gitlabService.id}/access-url`,
      },
      (data) => {
        assertCondition(data.url?.includes("/s/gitlab"), "GitLab 访问 URL 返回异常");
        assertCondition(data.expiresIn > 0, "GitLab 访问 URL 未返回有效 expiresIn");
      },
    );
    const accessCookie = cookieHeaderFrom(accessURLResult.response);
    assertCondition(accessCookie, "客户端创建 GitLab 访问 URL 未返回访问 cookie");

    await requestJSON("客户端请求 Jenkins 访问 URL 应被拒绝", {
      accessToken: clientAccessToken,
      expectedStatus: 403,
      method: "POST",
      path: "/api/v1/client/services/service_jenkins/access-url",
    }).then(({ payload }) => {
      assertCondition(payload.success === false, "Jenkins 拒绝访问未返回 success=false");
      assertCondition(payload.error?.code === "POLICY_ACCESS_DENIED", "Jenkins 拒绝访问错误码异常");
      console.log(
        `成功：客户端请求 Jenkins 访问 URL 被策略拒绝（requestId=${payload.meta?.requestId}）`,
      );
    });

    await jsonSuccess(
      "GitLab 代理访问",
      {
        headers: { cookie: accessCookie },
        path: appendPath(accessURLResult.payload.data.url, "whoami"),
      },
      (payload) => {
        assertCondition(payload.serviceKey === "gitlab", "GitLab 代理访问返回异常");
        assertCondition(payload.serviceName?.includes("GitLab"), "GitLab 代理访问上游服务名异常");
      },
    );

    for (const { path, serviceKey } of [
      { path: "/debug/upstreams/gitlab", serviceKey: "gitlab" },
      { path: "/debug/upstreams/jenkins", serviceKey: "jenkins" },
      { path: "/debug/upstreams/docs", serviceKey: "docs" },
      { path: "/debug/upstreams/internal-admin", serviceKey: "internal-admin" },
    ]) {
      await jsonSuccess(
        `私有上游探测：${serviceKey}`,
        {
          path,
        },
        (payload) => {
          assertCondition(payload.serviceKey === serviceKey, `${serviceKey} 私有上游探测返回异常`);
        },
      );
    }
  } finally {
    await cleanupSmokeDevice(adminAccessToken, smokeDeviceID);
  }
}

try {
  console.log(`远端 dev Gateway：${gatewayBaseURL}`);
  await main();
  console.log("");
  console.log("远端 dev 接口 smoke 测试全部通过。");
} catch (error) {
  console.error("");
  console.error(
    `远端 dev 接口 smoke 测试失败：${error instanceof Error ? error.message : String(error)}`,
  );
  process.exitCode = 1;
}
