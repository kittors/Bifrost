import "@bifrost/design-tokens/app.css";
import { type ApiResponse, createApiSuccess } from "@bifrost/contracts";
import { Button } from "@bifrost/ui";

export const appName = "Bifrost Admin";

export type AdminBootstrapResponse = ApiResponse<{
  appName: string;
}>;

export const adminBootstrapResponse = createApiSuccess({
  data: { appName },
  requestId: "req_admin_bootstrap",
  timestamp: "2026-04-17T00:00:00Z",
});

export const adminCanConsumeSharedUi = Boolean(Button);
