import { Button } from "@bifrost/ui";
import { Link, Outlet, useLocation, useRouter } from "@tanstack/react-router";
import {
  ClipboardList,
  MonitorSmartphone,
  MoonStar,
  ServerCog,
  Shield,
  ShieldCheck,
  SunMedium,
  Users,
} from "lucide-react";
import { useEffect, useState } from "react";
import { toast } from "sonner";

import { adminLogout } from "../../features/auth/api";
import { useAdminSessionStore } from "../../features/auth/store";

type ThemeMode = "light" | "dark";

export const adminThemeStorageKey = "bifrost.admin.theme";

export function readAdminTheme(): ThemeMode {
  if (typeof window === "undefined") {
    return "light";
  }

  const stored = window.localStorage.getItem(adminThemeStorageKey);
  return stored === "dark" ? "dark" : "light";
}

export function applyAdminTheme(theme: ThemeMode) {
  document.documentElement.setAttribute("data-theme", theme);
  window.localStorage.setItem(adminThemeStorageKey, theme);
}

const navigationItems = [
  { icon: ShieldCheck, label: "概览", to: "/" },
  { icon: Users, label: "用户", to: "/users" },
  { icon: Shield, label: "角色", to: "/roles" },
  { icon: MonitorSmartphone, label: "设备", to: "/devices" },
  { icon: ServerCog, label: "服务", to: "/services" },
  { icon: ClipboardList, label: "审计", to: "/audit-events" },
] as const;

const pageTitles: Record<string, string> = {
  "/": "系统概览",
  "/audit-events": "审计记录",
  "/devices": "设备管理",
  "/roles": "角色管理",
  "/services": "服务目录",
  "/users": "用户管理",
};

export function AdminShell() {
  const router = useRouter();
  const location = useLocation();
  const session = useAdminSessionStore((state) => state.session);
  const clearSession = useAdminSessionStore((state) => state.clearSession);
  const [theme, setTheme] = useState<ThemeMode>(readAdminTheme);

  useEffect(() => {
    applyAdminTheme(theme);
  }, [theme]);

  return (
    <div className="min-h-screen bg-[radial-gradient(circle_at_top_left,color-mix(in_oklab,var(--bifrost-brand)_10%,transparent),transparent_28%),linear-gradient(180deg,var(--bifrost-bg),color-mix(in_oklab,var(--bifrost-surface)_80%,var(--bifrost-bg)))] text-text-primary">
      <div className="grid min-h-screen grid-cols-[232px_minmax(0,1fr)]">
        <aside className="border-r border-border-soft bg-[color-mix(in_oklab,var(--bifrost-surface)_86%,transparent)] px-4 py-5 backdrop-blur-sm">
          <div className="mb-8 flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-[10px] bg-brand-soft text-brand">
              <Shield className="h-5 w-5" />
            </div>
            <div>
              <div className="text-[14px] leading-[22px] font-semibold">Bifrost</div>
              <div className="text-[12px] leading-[18px] text-text-secondary">Admin Console</div>
            </div>
          </div>

          <nav className="space-y-1.5">
            {navigationItems.map((item) => {
              const Icon = item.icon;
              const active = location.pathname === item.to;

              return (
                <Link
                  activeProps={{
                    className:
                      "border-brand/20 bg-brand-soft text-brand shadow-[inset_0_0_0_1px_color-mix(in_oklab,var(--bifrost-brand)_14%,transparent)]",
                  }}
                  className="flex h-10 items-center gap-3 rounded-[10px] border border-transparent px-3 text-[13px] leading-[20px] text-text-secondary transition-colors hover:bg-surface hover:text-text-primary"
                  key={item.to}
                  to={item.to}
                >
                  <Icon className={`h-4 w-4 ${active ? "text-brand" : ""}`} />
                  <span>{item.label}</span>
                </Link>
              );
            })}
          </nav>
        </aside>

        <div className="min-w-0">
          <header className="flex h-[52px] items-center justify-between border-b border-border-soft px-6">
            <div>
              <div className="text-[15px] leading-[22px] font-semibold">
                {pageTitles[location.pathname] ?? "Bifrost Admin"}
              </div>
            </div>
            <div className="flex items-center gap-2">
              <Button
                onClick={() => {
                  setTheme((current) => (current === "light" ? "dark" : "light"));
                }}
                size="sm"
                variant="ghost"
              >
                {theme === "light" ? (
                  <MoonStar className="h-4 w-4" />
                ) : (
                  <SunMedium className="h-4 w-4" />
                )}
                <span>{theme === "light" ? "Dark" : "Light"}</span>
              </Button>
              <div className="rounded-[10px] border border-border bg-surface px-3 py-1.5 text-right">
                <div className="text-[13px] leading-[20px] font-medium">
                  {session?.user.displayName}
                </div>
                <div className="text-[12px] leading-[18px] text-text-secondary">
                  {session?.user.username}
                </div>
              </div>
              <Button
                onClick={async () => {
                  const accessToken = session?.accessToken;
                  clearSession();
                  if (accessToken) {
                    try {
                      await adminLogout(accessToken);
                    } catch {
                      // Logout is best-effort because local state is already cleared.
                    }
                  }
                  toast.success("管理员会话已退出");
                  await router.navigate({ to: "/login" });
                }}
                size="sm"
                variant="secondary"
              >
                退出
              </Button>
            </div>
          </header>

          <main className="px-6 py-5">
            <Outlet />
          </main>
        </div>
      </div>
    </div>
  );
}
