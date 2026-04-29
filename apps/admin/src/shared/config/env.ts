const fallbackGatewayBaseURL = "http://127.0.0.1:8080";

export function getGatewayBaseURL() {
  const configured = import.meta.env.VITE_GATEWAY_BASE_URL?.trim();

  if (configured) {
    return configured.replace(/\/+$/, "");
  }

  return fallbackGatewayBaseURL;
}
