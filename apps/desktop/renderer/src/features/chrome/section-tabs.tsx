import { Button } from "@bifrost/ui";

import { useDesktopSessionStore } from "../../entities/session/store";
import type { DesktopView } from "../../entities/session/types";

const tabs: Array<{ label: string; view: DesktopView }> = [
  { label: "服务", view: "services" },
  { label: "账号", view: "account" },
  { label: "设置", view: "settings" },
  { label: "诊断", view: "diagnostics" },
];

export function SectionTabs() {
  const { setView, view } = useDesktopSessionStore();

  return (
    <section className="grid grid-cols-4 gap-2">
      {tabs.map((tab) => (
        <Button
          key={tab.view}
          onClick={() => setView(tab.view)}
          size="sm"
          variant={tab.view === view ? "secondary" : "ghost"}
        >
          {tab.label}
        </Button>
      ))}
    </section>
  );
}
