import { useQuery } from "@tanstack/react-query";
import { useEffect } from "react";

import { loadLocalDeviceIdentity } from "../entities/device/api";
import { useDesktopSessionStore } from "../entities/session/store";
import { AccountCard } from "../features/account/account-card";
import { LoginCard } from "../features/auth/login-card";
import { ConnectionBanner } from "../features/chrome/connection-banner";
import { SectionTabs } from "../features/chrome/section-tabs";
import { DiagnosticsCard } from "../features/diagnostics/diagnostics-card";
import { ServicesCard } from "../features/services/services-card";
import { SettingsCard } from "../features/settings/settings-card";
import { WindowShell } from "./layout/window-shell";

export function DesktopApp() {
  const {
    device,
    errorMessage,
    hydrateFromSecureStore,
    session,
    setDevice,
    setErrorMessage,
    theme,
    view,
  } = useDesktopSessionStore();

  useEffect(() => {
    void hydrateFromSecureStore();
    void loadLocalDeviceIdentity()
      .then(setDevice)
      .catch(() => {
        setErrorMessage("本地设备身份读取失败");
      });
  }, [hydrateFromSecureStore, setDevice, setErrorMessage]);

  useEffect(() => {
    document.documentElement.setAttribute("data-theme", theme);
  }, [theme]);

  const diagnosticsQuery = useQuery({
    queryFn: () => window.bifrostDesktop.diagnostics.snapshot(),
    queryKey: ["desktop-diagnostics"],
  });

  return (
    <WindowShell>
      <ConnectionBanner
        device={device}
        diagnostics={diagnosticsQuery.data ?? null}
        errorMessage={errorMessage}
        session={session}
      />

      {session ? (
        <>
          <SectionTabs />
          {view === "services" ? <ServicesCard /> : null}
          {view === "account" ? <AccountCard /> : null}
          {view === "settings" ? <SettingsCard /> : null}
          {view === "diagnostics" ? (
            <DiagnosticsCard diagnostics={diagnosticsQuery.data ?? null} />
          ) : null}
        </>
      ) : (
        <LoginCard />
      )}
    </WindowShell>
  );
}
