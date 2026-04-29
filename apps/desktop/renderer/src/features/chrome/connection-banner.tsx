import { Alert, Chip } from "@heroui/react";
import { Fingerprint, LockKeyhole, Wifi, WifiOff } from "lucide-react";

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
  const fingerprint = device?.fingerprint ? device.fingerprint.slice(-12) : "未生成设备身份";
  const ConnectionIcon = session ? Wifi : WifiOff;

  return (
    <section className="rounded-[14px] border border-border bg-surface px-3 py-3">
      <div className="flex items-start justify-between gap-3">
        <div className="flex min-w-0 items-start gap-3">
          <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-[10px] bg-surface-2 text-text-secondary">
            <ConnectionIcon className="h-4 w-4" />
          </div>
          <div className="min-w-0">
            <div className="text-[14px] leading-[22px] font-semibold">
              {session ? "已连接 Bifrost Gateway" : "等待登录"}
            </div>
            <div className="truncate text-[12px] leading-[18px] text-text-secondary">
              {session ? `${session.user.displayName} · ${session.user.username}` : "未建立会话"}
            </div>
          </div>
        </div>
        <Chip color={session ? "success" : "default"} size="sm" variant="soft">
          {session ? "Online" : "Offline"}
        </Chip>
      </div>

      <div className="mt-3 grid grid-cols-2 gap-2">
        <div className="flex min-w-0 items-center gap-2 rounded-[10px] border border-border-soft bg-surface-2 px-2 py-1.5 text-[12px] leading-[18px] text-text-secondary">
          <Fingerprint className="h-3.5 w-3.5 shrink-0" />
          <span className="truncate">{fingerprint}</span>
        </div>
        <div className="flex min-w-0 items-center gap-2 rounded-[10px] border border-border-soft bg-surface-2 px-2 py-1.5 text-[12px] leading-[18px] text-text-secondary">
          <LockKeyhole className="h-3.5 w-3.5 shrink-0" />
          <span className="truncate">
            {diagnostics?.encryptionAvailable ? "安全存储可用" : "安全存储待确认"}
          </span>
        </div>
      </div>

      {errorMessage ? (
        <Alert className="mt-3" status="danger">
          <Alert.Content>
            <Alert.Title>客户端状态异常</Alert.Title>
            <Alert.Description>{errorMessage}</Alert.Description>
          </Alert.Content>
        </Alert>
      ) : null}
    </section>
  );
}
