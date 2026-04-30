import type { LucideIcon } from "lucide-react";
import { ClipboardCheck, Layers3, Server, ShieldCheck, UsersRound } from "lucide-react";

import { LoginForm } from "../features/auth/login-form";

type LoginFeature = {
  description: string;
  Icon: LucideIcon;
  title: string;
};

const loginFeatures: LoginFeature[] = [
  {
    description: "账号、角色和例外授权统一收口",
    Icon: UsersRound,
    title: "Users",
  },
  {
    description: "显式配置上游与公开入口",
    Icon: Server,
    title: "Services",
  },
  {
    description: "每次关键操作都可追踪 requestId",
    Icon: ClipboardCheck,
    title: "Audit",
  },
];

export function LoginPage() {
  return (
    <div className="relative flex min-h-screen items-center justify-center overflow-hidden bg-[linear-gradient(135deg,color-mix(in_oklab,var(--bifrost-brand)_10%,var(--bifrost-bg))_0%,var(--bifrost-bg)_42%,color-mix(in_oklab,var(--bifrost-brand)_8%,var(--bifrost-bg))_100%)] px-4 py-8 text-text-primary sm:px-6 lg:px-8">
      <div
        aria-hidden="true"
        className="pointer-events-none absolute top-[7%] right-[4%] h-[104px] w-[128px] opacity-55 [background-image:radial-gradient(color-mix(in_oklab,var(--bifrost-brand)_52%,transparent)_1.4px,transparent_1.4px)] [background-size:16px_16px]"
      />
      <div
        aria-hidden="true"
        className="pointer-events-none absolute bottom-[7%] left-[3%] h-[88px] w-[120px] opacity-45 [background-image:radial-gradient(color-mix(in_oklab,var(--bifrost-brand)_48%,transparent)_1.3px,transparent_1.3px)] [background-size:16px_16px]"
      />
      <div
        aria-hidden="true"
        className="pointer-events-none absolute right-[-120px] bottom-[-150px] h-[360px] w-[360px] rounded-full border border-[color-mix(in_oklab,var(--bifrost-brand)_18%,transparent)]"
      />
      <div
        aria-hidden="true"
        className="pointer-events-none absolute right-[-82px] bottom-[-112px] h-[284px] w-[284px] rounded-full border border-[color-mix(in_oklab,var(--bifrost-brand)_14%,transparent)]"
      />

      <main className="relative z-10 w-full max-w-[1260px] rounded-[28px] border border-white/75 bg-white/70 p-3 shadow-[0_34px_120px_-48px_rgba(15,23,42,0.42)] backdrop-blur-sm">
        <div className="grid min-h-[560px] overflow-hidden rounded-[22px] border border-border-soft bg-surface lg:grid-cols-[minmax(0,1.35fr)_minmax(380px,0.85fr)]">
          <section
            aria-label="后台登录产品说明"
            className="relative hidden overflow-hidden bg-[linear-gradient(145deg,color-mix(in_oklab,var(--bifrost-brand)_8%,var(--bifrost-surface))_0%,color-mix(in_oklab,var(--bifrost-brand)_4%,var(--bifrost-surface))_48%,var(--bifrost-surface)_100%)] px-14 py-16 lg:flex lg:flex-col lg:justify-between"
          >
            <div
              aria-hidden="true"
              className="absolute right-14 top-16 flex h-16 w-16 rotate-45 items-center justify-center rounded-[16px] border border-white/60 bg-white/35 text-brand-soft shadow-[0_18px_60px_-28px_rgba(37,99,235,0.5)]"
            >
              <Layers3 className="h-8 w-8 -rotate-45 text-brand/20" />
            </div>
            <div
              aria-hidden="true"
              className="absolute inset-x-[-8%] bottom-0 h-[160px] bg-[linear-gradient(140deg,color-mix(in_oklab,var(--bifrost-brand)_8%,transparent),transparent_64%)]"
            />

            <div className="relative z-10 space-y-8">
              <div className="inline-flex items-center gap-2 rounded-full border border-[color-mix(in_oklab,var(--bifrost-brand)_28%,var(--bifrost-border))] bg-white/72 px-3.5 py-2 text-[13px] leading-[18px] font-medium text-brand shadow-[0_10px_30px_-22px_rgba(37,99,235,0.7)]">
                <span className="flex h-6 w-6 items-center justify-center rounded-full bg-brand-soft">
                  <ShieldCheck className="h-4 w-4" />
                </span>
                Compact security workspace
              </div>
              <div className="space-y-5">
                <h2 className="max-w-[540px] text-[34px] leading-[46px] font-semibold">
                  为私有服务访问配置一个清爽、可追溯的控制面。
                </h2>
                <p className="max-w-[520px] text-[16px] leading-[28px] text-text-secondary">
                  后台专注于账号、角色、设备、服务目录和审计，不承担流量转发，也不会影响终端网络环境。
                </p>
              </div>
            </div>

            <div className="relative z-10 grid gap-5 md:grid-cols-3">
              {loginFeatures.map(({ description, Icon, title }) => (
                <article
                  className="min-h-[148px] rounded-[14px] border border-border bg-white/82 p-5 shadow-[0_18px_48px_-34px_rgba(15,23,42,0.42)]"
                  key={title}
                >
                  <div className="mb-5 flex h-11 w-11 items-center justify-center rounded-[12px] bg-brand-soft text-brand">
                    <Icon className="h-5 w-5" />
                  </div>
                  <h3 className="mb-2 text-[16px] leading-[24px] font-semibold">{title}</h3>
                  <p className="text-[13px] leading-[22px] text-text-secondary">{description}</p>
                </article>
              ))}
            </div>
          </section>

          <section
            aria-label="管理员登录表单"
            className="flex items-center justify-center border-border-soft bg-surface px-5 py-10 sm:px-10 lg:border-l lg:px-14"
          >
            <LoginForm />
          </section>
        </div>
      </main>
    </div>
  );
}
