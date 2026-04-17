import { LoginForm } from "../features/auth/login-form";

export function LoginPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-[radial-gradient(circle_at_top,color-mix(in_oklab,var(--bifrost-brand)_14%,transparent),transparent_35%),linear-gradient(180deg,var(--bifrost-bg),color-mix(in_oklab,var(--bifrost-surface)_92%,var(--bifrost-bg)))] px-4">
      <div className="w-full max-w-[1040px] rounded-[24px] border border-border-soft bg-[color-mix(in_oklab,var(--bifrost-surface)_88%,transparent)] p-4 shadow-[0_30px_120px_-48px_rgba(15,23,42,0.45)] backdrop-blur-sm lg:grid lg:grid-cols-[minmax(0,1.2fr)_420px] lg:gap-4 lg:p-5">
        <section className="hidden rounded-[20px] bg-[linear-gradient(180deg,color-mix(in_oklab,var(--bifrost-brand)_6%,var(--bifrost-surface)),var(--bifrost-surface))] p-8 lg:flex lg:flex-col lg:justify-between">
          <div className="space-y-4">
            <div className="inline-flex rounded-full border border-border bg-surface px-3 py-1 text-[12px] leading-[18px] text-text-secondary">
              Compact security workspace
            </div>
            <div className="space-y-2">
              <h2 className="max-w-[420px] text-[28px] leading-[36px] font-semibold tracking-[-0.02em]">
                为私有服务访问配置一个清爽、可追溯的控制面。
              </h2>
              <p className="max-w-[460px] text-[14px] leading-[22px] text-text-secondary">
                后台专注于账号、角色、设备、服务目录和审计，不承担流量转发，也不会影响终端网络环境。
              </p>
            </div>
          </div>

          <div className="grid gap-3 md:grid-cols-3">
            {[
              ["Users", "账号、角色和例外授权统一收口"],
              ["Services", "显式配置上游与公开入口"],
              ["Audit", "每次关键操作都可追踪 requestId"],
            ].map(([title, description]) => (
              <div className="rounded-[14px] border border-border bg-surface p-4" key={title}>
                <div className="mb-2 text-[13px] leading-[20px] font-semibold">{title}</div>
                <div className="text-[12px] leading-[18px] text-text-secondary">{description}</div>
              </div>
            ))}
          </div>
        </section>

        <section className="flex items-center justify-center p-3">
          <LoginForm />
        </section>
      </div>
    </div>
  );
}
