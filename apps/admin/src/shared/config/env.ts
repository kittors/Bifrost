const defaultGatewayDevProxyBaseURL = "/__bifrost_gateway__";
const fallbackGatewayBaseURL = "http://142.171.208.80:18080";

function normalizeBaseURL(value: string) {
  return value.replace(/\/+$/, "");
}

export function getGatewayBaseURL() {
  if (import.meta.env.DEV) {
    const devProxyBaseURL = import.meta.env.VITE_GATEWAY_DEV_PROXY_BASE_URL?.trim();
    return normalizeBaseURL(devProxyBaseURL || defaultGatewayDevProxyBaseURL);
  }

  const configured = import.meta.env.VITE_GATEWAY_BASE_URL?.trim();

  if (configured) {
    return normalizeBaseURL(configured);
  }

  return fallbackGatewayBaseURL;
}
