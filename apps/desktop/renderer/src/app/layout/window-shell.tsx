import type { PropsWithChildren } from "react";
import { ShieldCheck } from "lucide-react";

type WindowShellProps = PropsWithChildren<{
  mode?: "app" | "login";
}>;

// 桌面窗口内容保持卡片式密度，不引入后台式大工作区布局。
export function WindowShell({ children, mode = "app" }: WindowShellProps) {
  if (mode === "login") {
    return (
      <main
        className="relative min-h-screen overflow-x-hidden overflow-y-auto bg-[radial-gradient(circle_at_50%_10%,rgba(69,144,255,0.12),transparent_34%),linear-gradient(145deg,#fbfdff_0%,#eef6ff_58%,#fbfdff_100%)] text-text-primary"
        data-testid="desktop-login-surface"
      >
        <LoginNetworkBackdrop />

        <div className="relative z-10 mx-auto flex min-h-screen w-full max-w-[560px] flex-col justify-center px-9 py-10">
          {children}
        </div>
      </main>
    );
  }

  return (
    <main className="min-h-screen bg-[radial-gradient(circle_at_top,color-mix(in_oklab,var(--bifrost-brand)_10%,transparent),transparent_34%),linear-gradient(180deg,var(--bifrost-bg),color-mix(in_oklab,var(--bifrost-surface)_94%,var(--bifrost-bg)))] px-4 py-4 text-text-primary">
      <div className="mx-auto flex min-h-[calc(100vh-32px)] max-w-[440px] flex-col overflow-hidden rounded-[18px] border border-border-soft bg-[color-mix(in_oklab,var(--bifrost-surface)_90%,transparent)] shadow-[0_28px_80px_-42px_rgba(15,23,42,0.38)] backdrop-blur-sm">
        <header className="flex h-[52px] shrink-0 items-center justify-between border-b border-border-soft px-4">
          <div className="flex items-center gap-3">
            <div className="flex h-9 w-9 items-center justify-center rounded-[10px] bg-brand-soft text-brand">
              <ShieldCheck className="h-4 w-4" />
            </div>
            <div>
              <div className="text-[14px] leading-[20px] font-semibold">Bifrost Client</div>
              <div className="text-[12px] leading-[18px] text-text-secondary">Local Access</div>
            </div>
          </div>
          <div className="rounded-full border border-border bg-surface px-2 py-1 text-[11px] leading-[16px] text-text-secondary">
            Desktop
          </div>
        </header>

        <div className="flex min-h-0 flex-1 flex-col gap-3 p-3">{children}</div>
      </div>
    </main>
  );
}

function LoginNetworkBackdrop() {
  return (
    <svg
      aria-hidden="true"
      className="pointer-events-none absolute inset-0 z-0 h-full min-h-[560px] w-full text-white/70"
      data-testid="login-network-backdrop"
      fill="none"
      viewBox="0 0 420 560"
      xmlns="http://www.w3.org/2000/svg"
    >
      <g opacity="0.42" stroke="currentColor" strokeLinecap="round" strokeWidth="1.2">
        <path d="M328 72L370 128L420 88" />
        <path d="M370 128L420 160" />
        <path d="M334 228L372 194L420 214" />
        <path d="M24 520L42 480L90 535" />
        <path d="M42 480L104 510" />
      </g>
      <g fill="white" opacity="0.72">
        <circle cx="328" cy="72" r="3" />
        <circle cx="370" cy="128" r="3.6" />
        <circle cx="420" cy="160" r="3" />
        <circle cx="372" cy="194" r="2.8" />
        <circle cx="42" cy="480" r="3" />
        <circle cx="104" cy="510" r="2.8" />
      </g>
    </svg>
  );
}
