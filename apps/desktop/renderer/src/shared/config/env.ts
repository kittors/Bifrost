const fallbackGatewayBaseURL = "http://127.0.0.1:8080";

export function normalizeGatewayBaseURL(value: string) {
  const trimmed = value.trim();
  return trimmed.replace(/\/+$/, "") || fallbackGatewayBaseURL;
}
