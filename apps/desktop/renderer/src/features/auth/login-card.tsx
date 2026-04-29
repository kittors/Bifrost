import { Alert, Button, Card, CardContent, InputGroup, Label, TextField } from "@heroui/react";
import { useState } from "react";
import type { ComponentProps, FormEvent, ReactNode } from "react";
import { Eye, EyeOff, Globe2, LockKeyhole, UserRound } from "lucide-react";
import type { LucideIcon } from "lucide-react";
import type { DesktopSessionSnapshot } from "../../../../electron/shared/types";
import { bootstrapClientDevice, ensureLocalDeviceIdentity } from "../../entities/device/api";
import { clientLogin } from "../../entities/session/api";
import { useDesktopSessionStore } from "../../entities/session/store";
import { resolveApiErrorMessage } from "../../shared/lib/http";

const clientVersion = "0.1.0";
const desktopLoginIconClassName = "h-4 w-4";

// 登录卡片负责首登 bootstrap 和已绑定设备登录两条路径。
export function LoginCard() {
  const {
    errorMessage,
    gatewayBaseURL,
    saveSession,
    setErrorMessage,
    setGatewayBaseURL,
    updateDeviceID,
  } = useDesktopSessionStore();
  const [username, setUsername] = useState("alice");
  const [password, setPassword] = useState("ChangeMe123!");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isPasswordVisible, setIsPasswordVisible] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
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
        deviceID = bootstrap.device?.deviceId;
        if (!deviceID) {
          throw new Error("服务端未返回设备绑定结果，请稍后重试");
        }
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
  }

  return (
    <div className="grid w-full gap-7">
      <header className="mx-auto grid w-fit grid-cols-[56px_minmax(0,auto)] items-center gap-4">
        <BifrostLogo className="h-14 w-14 shrink-0" />
        <div className="min-w-0">
          <h1 className="text-[26px] leading-[32px] font-semibold text-text-primary">
            Bifrost Desktop
          </h1>
          <p className="mt-1 text-[14px] leading-[20px] font-medium text-text-secondary">
            安全 · 稳定 · 高效
          </p>
        </div>
      </header>

      <Card className="w-full" variant="default">
        <CardContent>
          <form className="grid min-w-0 gap-4" onSubmit={handleSubmit}>
            <LoginField
              autoComplete="url"
              icon={Globe2}
              label="服务端地址"
              value={gatewayBaseURL}
              onChange={(event) => setGatewayBaseURL(event.target.value)}
              placeholder="http://142.171.208.80:18080"
            />
            <LoginField
              autoComplete="username"
              icon={UserRound}
              label="用户名"
              value={username}
              onChange={(event) => setUsername(event.target.value)}
              placeholder="输入用户名"
            />
            <LoginField
              autoComplete="current-password"
              icon={LockKeyhole}
              label="密码"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              placeholder="输入密码"
              suffix={
                <Button
                  aria-label={isPasswordVisible ? "隐藏密码" : "显示密码"}
                  className="h-7 min-h-7 w-7 min-w-7 rounded-full p-0 text-text-secondary focus-visible:shadow-none"
                  isIconOnly
                  size="sm"
                  type="button"
                  variant="ghost"
                  onPress={() => setIsPasswordVisible((value) => !value)}
                >
                  {isPasswordVisible ? (
                    <EyeOff aria-hidden="true" className="h-[18px] w-[18px]" />
                  ) : (
                    <Eye aria-hidden="true" className="h-[18px] w-[18px]" />
                  )}
                </Button>
              }
              type={isPasswordVisible ? "text" : "password"}
            />

            {errorMessage ? (
              <Alert
                className="rounded-[12px] border border-danger/20 bg-[color-mix(in_oklab,var(--bifrost-danger)_8%,transparent)] px-3 py-2"
                status="danger"
              >
                <Alert.Content>
                  <Alert.Description className="text-[12px] leading-[18px]">
                    {errorMessage}
                  </Alert.Description>
                </Alert.Content>
              </Alert>
            ) : null}

            <Button
              className="mt-1 font-semibold"
              fullWidth
              isDisabled={isSubmitting}
              size="sm"
              type="submit"
            >
              {isSubmitting ? "登录中..." : "登录"}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}

type LoginFieldProps = Omit<ComponentProps<typeof InputGroup.Input>, "className"> & {
  icon: LucideIcon;
  label: string;
  suffix?: ReactNode;
};

function LoginField({ icon: Icon, label, suffix, ...inputProps }: LoginFieldProps) {
  return (
    <TextField className="min-w-0" fullWidth>
      <Label>{label}</Label>
      <InputGroup className="min-w-0" fullWidth>
        <InputGroup.Prefix>
          <Icon aria-hidden="true" className={desktopLoginIconClassName} />
        </InputGroup.Prefix>
        <InputGroup.Input className="min-w-0" {...inputProps} />
        {suffix ? <InputGroup.Suffix>{suffix}</InputGroup.Suffix> : null}
      </InputGroup>
    </TextField>
  );
}

type BifrostLogoProps = {
  className?: string;
};

function BifrostLogo({ className }: BifrostLogoProps) {
  return (
    <svg
      aria-label="Bifrost logo"
      className={className}
      fill="none"
      role="img"
      viewBox="0 0 72 72"
      xmlns="http://www.w3.org/2000/svg"
    >
      <title>Bifrost logo</title>
      <defs>
        <linearGradient id="bifrost-logo-stroke" x1="14" x2="58" y1="12" y2="60">
          <stop stopColor="#4BA3FF" />
          <stop offset="0.52" stopColor="#2778F6" />
          <stop offset="1" stopColor="#1851E8" />
        </linearGradient>
        <linearGradient id="bifrost-logo-fill" x1="23" x2="49" y1="22" y2="50">
          <stop stopColor="#EAF4FF" />
          <stop offset="1" stopColor="#D8E8FF" />
        </linearGradient>
      </defs>
      <path
        d="M36 8.5L59 21.75V50.25L36 63.5L13 50.25V21.75L36 8.5Z"
        fill="url(#bifrost-logo-fill)"
        stroke="url(#bifrost-logo-stroke)"
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="3"
      />
      <path
        d="M22 28L36 20L50 28V44L36 52L22 44V28Z"
        stroke="url(#bifrost-logo-stroke)"
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="2.4"
      />
      <path
        d="M23.5 43.5L36 36L48.5 43.5M23.5 28.5L36 36L48.5 28.5M36 20.5V51.5"
        stroke="url(#bifrost-logo-stroke)"
        strokeLinecap="round"
        strokeLinejoin="round"
        strokeWidth="2.2"
      />
      <circle cx="36" cy="36" fill="#2675F4" r="6.2" />
      <circle cx="36" cy="20.5" fill="#1E66F2" r="3.8" />
      <circle cx="36" cy="51.5" fill="#1E66F2" r="3.8" />
    </svg>
  );
}
