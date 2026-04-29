import {
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Chip,
} from "@heroui/react";
import { LogOut, UserRound } from "lucide-react";

import { logoutClientSession } from "../../entities/session/api";
import { useDesktopSessionStore } from "../../entities/session/store";
import { resolveApiErrorMessage } from "../../shared/lib/http";

export function AccountCard() {
  const { clearSession, device, session, setErrorMessage } = useDesktopSessionStore();

  if (!session) {
    return null;
  }

  return (
    <Card>
      <CardHeader className="pb-2">
        <div className="flex items-center gap-3">
          <div className="flex h-9 w-9 items-center justify-center rounded-[10px] bg-brand-soft text-brand">
            <UserRound className="h-4 w-4" />
          </div>
          <div>
            <CardTitle>账号与设备</CardTitle>
            <CardDescription>{session.user.username}</CardDescription>
          </div>
        </div>
        <Chip color="success" size="sm" variant="soft">
          Trusted
        </Chip>
      </CardHeader>
      <CardContent className="space-y-3">
        <div className="grid gap-2 text-[12px] leading-[18px]">
          <InfoRow label="显示名称" value={session.user.displayName} />
          <InfoRow label="角色" value={session.user.roles.join(", ") || "-"} />
          <InfoRow label="Device ID" value={device?.deviceId ?? "未写入 deviceId"} />
          <InfoRow label="Fingerprint" value={device?.fingerprint ?? "未生成指纹"} />
        </div>
        <Button
          className="w-full"
          onClick={async () => {
            try {
              await logoutClientSession({
                accessToken: session.accessToken,
                baseURL: session.gatewayBaseURL,
              });
            } catch (error) {
              setErrorMessage(resolveApiErrorMessage(error, "退出登录失败"));
            } finally {
              await clearSession();
            }
          }}
          size="sm"
          variant="secondary"
        >
          <LogOut className="h-4 w-4" />
          退出登录
        </Button>
      </CardContent>
    </Card>
  );
}

function InfoRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="grid grid-cols-[88px_minmax(0,1fr)] gap-2 rounded-[10px] border border-border-soft bg-surface-2 px-2 py-1.5">
      <span className="text-text-muted">{label}</span>
      <span className="truncate text-text-secondary">{value}</span>
    </div>
  );
}
