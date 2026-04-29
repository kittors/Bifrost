const fallbackGatewayBaseURL = "http://142.171.208.80:18080";

export function normalizeGatewayBaseURL(value: string) {
  const trimmed = value.trim();
  return trimmed.replace(/\/+$/, "") || fallbackGatewayBaseURL;
}
