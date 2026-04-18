import { Button, Input } from "@bifrost/ui";
import { useState } from "react";
import type { DesktopSessionSnapshot } from "../../../../electron/shared/types";
import { bootstrapClientDevice, ensureLocalDeviceIdentity } from "../../entities/device/api";
import { clientLogin } from "../../entities/session/api";
import { useDesktopSessionStore } from "../../entities/session/store";
import { resolveApiErrorMessage } from "../../shared/lib/http";

const clientVersion = "0.1.0";

// 登录卡片负责首登 bootstrap 和已绑定设备登录两条路径。
export function LoginCard() {
  const { gatewayBaseURL, saveSession, setErrorMessage, setGatewayBaseURL, updateDeviceID } =
    useDesktopSessionStore();
  const [username, setUsername] = useState("alice");
  const [password, setPassword] = useState("ChangeMe123!");
  const [isSubmitting, setIsSubmitting] = useState(false);

  return (
    <section className="flex flex-1 flex-col justify-center gap-3 rounded-[14px] border border-border bg-surface p-4">
      <div className="space-y-1">
        <div className="text-[16px] leading-[24px] font-semibold">客户端登录</div>
        <p className="text-[12px] leading-[18px] text-text-secondary">
          只在当前窗口内保存会话状态，不修改系统代理、DNS 或路由。
        </p>
      </div>

      <Input value={gatewayBaseURL} onChange={(event) => setGatewayBaseURL(event.target.value)} />
      <Input
        value={username}
        onChange={(event) => setUsername(event.target.value)}
        placeholder="用户名"
      />
      <Input
        value={password}
        onChange={(event) => setPassword(event.target.value)}
        placeholder="密码"
        type="password"
      />

      <Button
        disabled={isSubmitting}
        onClick={async () => {
          setIsSubmitting(true);
          setErrorMessage(null);

          try {
            const identity = await ensureLocalDeviceIdentity();
            let result:
              | Awaited<ReturnType<typeof clientLogin>>
              | Awaited<ReturnType<typeof bootstrapClientDevice>>;
            let deviceID = identity.deviceId;

            if (deviceID) {
              result = await clientLogin({
                baseURL: gatewayBaseURL,
                clientVersion,
                deviceId: deviceID,
                password,
                username,
              });
            } else {
              const bootstrap = await bootstrapClientDevice(gatewayBaseURL, {
                clientVersion,
                deviceName: "Bifrost Desktop",
                deviceOs: navigator.platform || "unknown",
                password,
                publicKey: identity.publicKey,
                publicKeyFingerprint: identity.fingerprint,
                username,
              });
              deviceID = bootstrap.device.deviceId;
              await updateDeviceID(deviceID);
              result = bootstrap;
            }

            const session: DesktopSessionSnapshot = {
              accessToken: result.accessToken,
              deviceId: deviceID,
              expiresAt: new Date(Date.now() + result.expiresIn * 1000).toISOString(),
              gatewayBaseURL,
              refreshToken: result.refreshToken,
              user: result.user,
            };

            await saveSession(session);
          } catch (error) {
            setErrorMessage(resolveApiErrorMessage(error, "登录失败"));
          } finally {
            setIsSubmitting(false);
          }
        }}
      >
        {isSubmitting ? "登录中..." : "登录并绑定设备"}
      </Button>
    </section>
  );
}
