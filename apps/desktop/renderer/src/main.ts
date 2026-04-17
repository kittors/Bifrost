import "@bifrost/design-tokens/app.css";
import { type ApiResponse, createApiSuccess } from "@bifrost/contracts";
import { Button } from "@bifrost/ui";

export const rendererName = "bifrost-desktop-renderer";

export type DesktopBootstrapResponse = ApiResponse<{
  rendererName: string;
}>;

export const desktopBootstrapResponse = createApiSuccess({
  data: { rendererName },
  requestId: "req_desktop_bootstrap",
  timestamp: "2026-04-17T00:00:00Z",
});

export const desktopCanConsumeSharedUi = Boolean(Button);
