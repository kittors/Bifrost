const defaultTimeoutMs = 60_000;

function delay(ms) {
  return new Promise((resolve) => {
    setTimeout(resolve, ms);
  });
}

// 轮询等待 HTTP 健康检查，避免 E2E 在服务还未就绪时启动。
export async function waitForHTTP(url, timeoutMs = defaultTimeoutMs) {
  const startedAt = Date.now();

  while (Date.now() - startedAt < timeoutMs) {
    try {
      const response = await fetch(url);
      if (response.ok) {
        return;
      }
    } catch {
      // 服务尚未可连通时继续重试，不在这里提前中断。
    }

    await delay(1_000);
  }

  throw new Error(`timed out waiting for ${url}`);
}
