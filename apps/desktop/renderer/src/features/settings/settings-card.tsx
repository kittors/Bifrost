import {
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Chip,
  Input,
} from "@heroui/react";
import { MoonStar, Settings, SunMedium } from "lucide-react";

import { useDesktopSessionStore } from "../../entities/session/store";

export function SettingsCard() {
  const { gatewayBaseURL, setGatewayBaseURL, setTheme, theme } = useDesktopSessionStore();

  return (
    <Card>
      <CardHeader className="pb-2">
        <div className="flex items-center gap-3">
          <div className="flex h-9 w-9 items-center justify-center rounded-[10px] bg-brand-soft text-brand">
            <Settings className="h-4 w-4" />
          </div>
          <div>
            <CardTitle>设置</CardTitle>
            <CardDescription>客户端外观与 Gateway 地址。</CardDescription>
          </div>
        </div>
        <Chip color="default" size="sm" variant="soft">
          {theme === "light" ? "Light" : "Dark"}
        </Chip>
      </CardHeader>
      <CardContent className="space-y-3">
        <div className="space-y-1">
          <span className="text-[12px] leading-[18px] text-text-secondary">Gateway 地址</span>
          <Input
            value={gatewayBaseURL}
            onChange={(event) => setGatewayBaseURL(event.target.value)}
          />
        </div>
        <div className="flex gap-2">
          <Button
            className="flex-1"
            onClick={() => setTheme("light")}
            size="sm"
            variant={theme === "light" ? "secondary" : "ghost"}
          >
            <SunMedium className="h-4 w-4" />
            Light
          </Button>
          <Button
            className="flex-1"
            onClick={() => setTheme("dark")}
            size="sm"
            variant={theme === "dark" ? "secondary" : "ghost"}
          >
            <MoonStar className="h-4 w-4" />
            Dark
          </Button>
        </div>
      </CardContent>
    </Card>
  );
}
