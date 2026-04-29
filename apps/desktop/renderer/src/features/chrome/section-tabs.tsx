import { Tabs } from "@heroui/react";
import { Activity, Settings, UserRound, Waypoints } from "lucide-react";

import { useDesktopSessionStore } from "../../entities/session/store";
import type { DesktopView } from "../../entities/session/types";

const tabs: Array<{ icon: typeof Waypoints; label: string; view: DesktopView }> = [
  { icon: Waypoints, label: "服务", view: "services" },
  { icon: UserRound, label: "账号", view: "account" },
  { icon: Settings, label: "设置", view: "settings" },
  { icon: Activity, label: "诊断", view: "diagnostics" },
];

export function SectionTabs() {
  const { setView, view } = useDesktopSessionStore();

  return (
    <Tabs.Root selectedKey={view} onSelectionChange={(key) => setView(String(key) as DesktopView)}>
      <Tabs.List className="grid w-full grid-cols-4">
        {tabs.map((tab) => (
          <Tabs.Tab className="min-w-0 gap-1.5" id={tab.view} key={tab.view}>
            <tab.icon className="h-3.5 w-3.5" />
            {tab.label}
          </Tabs.Tab>
        ))}
      </Tabs.List>
    </Tabs.Root>
  );
}
