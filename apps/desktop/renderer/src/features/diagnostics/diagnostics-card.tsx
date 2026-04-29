import { Card, CardContent, CardDescription, CardHeader, CardTitle, Chip } from "@heroui/react";
import { Activity, CheckCircle2, CircleMinus } from "lucide-react";

import type {
  DesktopDiagnosticsSnapshot,
  DesktopLocalProxyStatus,
} from "../../../../electron/shared/types";

type DiagnosticsCardProps = {
  diagnostics: DesktopDiagnosticsSnapshot | null;
  localProxyStatus: DesktopLocalProxyStatus;
};

export function DiagnosticsCard({ diagnostics, localProxyStatus }: DiagnosticsCardProps) {
  return (
    <Card>
      <CardHeader className="pb-2">
        <div className="flex items-center gap-3">
          <div className="flex h-9 w-9 items-center justify-center rounded-[10px] bg-brand-soft text-brand">
            <Activity className="h-4 w-4" />
          </div>
          <div>
            <CardTitle>诊断</CardTitle>
            <CardDescription>{diagnostics?.platform ?? "unknown"}</CardDescription>
          </div>
        </div>
        <Chip color={localProxyStatus.running ? "success" : "warning"} size="sm" variant="soft">
          {localProxyStatus.running ? "Proxy On" : "Proxy Off"}
        </Chip>
      </CardHeader>
      <CardContent className="grid gap-2 text-[12px] leading-[18px]">
        <DiagnosticRow
          active={Boolean(diagnostics?.encryptionAvailable)}
          label="安全存储"
          value={diagnostics?.encryptionAvailable ? "可用" : "不可用"}
        />
        <DiagnosticRow
          active={localProxyStatus.running}
          label="本地入口"
          value={
            localProxyStatus.running
              ? `${localProxyStatus.baseURL}（127.0.0.1:${localProxyStatus.port}）`
              : "未启动"
          }
        />
        <DiagnosticRow
          active={!diagnostics?.proxyManagedByBifrost}
          label="系统代理"
          value={diagnostics?.proxyManagedByBifrost ? "已接管" : "未修改"}
        />
        <DiagnosticRow
          active={!diagnostics?.dnsManagedByBifrost}
          label="系统 DNS"
          value={diagnostics?.dnsManagedByBifrost ? "已接管" : "未修改"}
        />
        <DiagnosticRow
          active={!diagnostics?.routeManagedByBifrost}
          label="系统路由"
          value={diagnostics?.routeManagedByBifrost ? "已接管" : "未修改"}
        />
      </CardContent>
    </Card>
  );
}

function DiagnosticRow({
  active,
  label,
  value,
}: {
  active: boolean;
  label: string;
  value: string;
}) {
  const Icon = active ? CheckCircle2 : CircleMinus;

  return (
    <div className="grid grid-cols-[18px_72px_minmax(0,1fr)] items-center gap-2 rounded-[10px] border border-border-soft bg-surface-2 px-2 py-1.5">
      <Icon className={active ? "h-3.5 w-3.5 text-success" : "h-3.5 w-3.5 text-text-muted"} />
      <span className="text-text-muted">{label}</span>
      <span className="truncate text-text-secondary">{value}</span>
    </div>
  );
}
