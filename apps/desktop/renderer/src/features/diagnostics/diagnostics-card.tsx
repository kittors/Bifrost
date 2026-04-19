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
    <section className="rounded-[14px] border border-border bg-surface p-3">
      <div className="text-[14px] leading-[22px] font-semibold">诊断</div>
      <div className="mt-3 grid gap-2 text-[12px] leading-[18px] text-text-secondary">
        <div>平台：{diagnostics?.platform ?? "-"}</div>
        <div>安全存储：{diagnostics?.encryptionAvailable ? "可用" : "不可用"}</div>
        <div>
          本地入口：
          {localProxyStatus.running
            ? `${localProxyStatus.baseURL}（127.0.0.1:${localProxyStatus.port}）`
            : "未启动"}
        </div>
        <div>系统代理：{diagnostics?.proxyManagedByBifrost ? "已接管" : "未修改"}</div>
        <div>系统 DNS：{diagnostics?.dnsManagedByBifrost ? "已接管" : "未修改"}</div>
        <div>系统路由：{diagnostics?.routeManagedByBifrost ? "已接管" : "未修改"}</div>
      </div>
    </section>
  );
}
