import type { PropsWithChildren } from "react";

// 桌面窗口内容保持卡片式密度，不引入后台式大工作区布局。
export function WindowShell({ children }: PropsWithChildren) {
  return (
    <main className="min-h-screen bg-[radial-gradient(circle_at_top,color-mix(in_oklab,var(--bifrost-brand)_12%,transparent),transparent_34%),linear-gradient(180deg,var(--bifrost-bg),color-mix(in_oklab,var(--bifrost-surface)_94%,var(--bifrost-bg)))] px-4 py-4 text-text-primary">
      <div className="mx-auto flex min-h-[calc(100vh-32px)] max-w-[420px] flex-col gap-3 rounded-[18px] border border-border-soft bg-[color-mix(in_oklab,var(--bifrost-surface)_88%,transparent)] p-4 shadow-[0_28px_80px_-42px_rgba(15,23,42,0.38)] backdrop-blur-sm">
        {children}
      </div>
    </main>
  );
}
