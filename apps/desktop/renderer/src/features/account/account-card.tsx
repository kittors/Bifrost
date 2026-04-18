import { Button } from "@bifrost/ui";

import { logoutClientSession } from "../../entities/session/api";
import { useDesktopSessionStore } from "../../entities/session/store";
import { resolveApiErrorMessage } from "../../shared/lib/http";

export function AccountCard() {
  const { clearSession, device, session, setErrorMessage } = useDesktopSessionStore();

  if (!session) {
    return null;
  }

  return (
    <section className="rounded-[14px] border border-border bg-surface p-3">
      <div className="text-[14px] leading-[22px] font-semibold">账号与设备</div>
      <div className="mt-3 space-y-1 text-[12px] leading-[18px] text-text-secondary">
        <div>{session.user.displayName}</div>
        <div>{session.user.username}</div>
        <div>{session.user.roles.join(", ") || "-"}</div>
        <div>{device?.deviceId ?? "未写入 deviceId"}</div>
        <div>{device?.fingerprint ?? "未生成指纹"}</div>
      </div>
      <Button
        className="mt-3 w-full"
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
        退出登录
      </Button>
    </section>
  );
}
