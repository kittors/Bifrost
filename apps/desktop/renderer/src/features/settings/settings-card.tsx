import { Button, Input } from "@bifrost/ui";

import { useDesktopSessionStore } from "../../entities/session/store";

export function SettingsCard() {
  const { gatewayBaseURL, setGatewayBaseURL, setTheme, theme } = useDesktopSessionStore();

  return (
    <section className="rounded-[14px] border border-border bg-surface p-3">
      <div className="text-[14px] leading-[22px] font-semibold">设置</div>
      <div className="mt-3 space-y-3">
        <div className="space-y-1">
          <span className="text-[12px] leading-[18px] text-text-secondary">Gateway 地址</span>
          <Input
            value={gatewayBaseURL}
            onChange={(event) => setGatewayBaseURL(event.target.value)}
          />
        </div>
        <div className="flex gap-2">
          <Button
            onClick={() => setTheme("light")}
            size="sm"
            variant={theme === "light" ? "secondary" : "ghost"}
          >
            Light
          </Button>
          <Button
            onClick={() => setTheme("dark")}
            size="sm"
            variant={theme === "dark" ? "secondary" : "ghost"}
          >
            Dark
          </Button>
        </div>
      </div>
    </section>
  );
}
