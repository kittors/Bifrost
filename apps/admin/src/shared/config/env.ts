const fallbackGatewayBaseURL = "http://142.171.208.80:18080";

export function getGatewayBaseURL() {
  const configured = import.meta.env.VITE_GATEWAY_BASE_URL?.trim();

  if (configured) {
    return configured.replace(/\/+$/, "");
  }

  return fallbackGatewayBaseURL;
}
