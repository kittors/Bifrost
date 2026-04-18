import type {
  DesktopDeviceIdentity,
  DesktopDiagnosticsSnapshot,
  DesktopSessionSnapshot,
} from "../../../../electron/shared/types";

type ConnectionBannerProps = {
  device: DesktopDeviceIdentity | null;
  diagnostics: DesktopDiagnosticsSnapshot | null;
  errorMessage: string | null;
  session: DesktopSessionSnapshot | null;
};

export function ConnectionBanner({
  device,
  diagnostics,
  errorMessage,
  session,
}: ConnectionBannerProps) {
  return (
    <section className="rounded-[14px] border border-border bg-surface px-3 py-3">
      <div className="flex items-start justify-between gap-3">
        <div>
          <div className="text-[14px] leading-[22px] font-semibold">
            {session ? "已连接 Bifrost Gateway" : "等待登录"}
          </div>
          <div className="text-[12px] leading-[18px] text-text-secondary">
            {session
              ? `${session.user.displayName} · ${device?.fingerprint ?? "未生成设备身份"}`
              : "登录后可查看授权服务并在系统浏览器中打开"}
          </div>
        </div>
        <div className="rounded-full border border-border bg-surface-2 px-2 py-1 text-[11px] leading-[16px] text-text-secondary">
          {diagnostics?.encryptionAvailable ? "安全存储可用" : "安全存储待确认"}
        </div>
      </div>

      {errorMessage ? (
        <div className="mt-3 rounded-[10px] border border-danger/20 bg-[color-mix(in_oklab,var(--bifrost-danger)_10%,transparent)] px-3 py-2 text-[12px] leading-[18px] text-danger">
          {errorMessage}
        </div>
      ) : null}
    </section>
  );
}
