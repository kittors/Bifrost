import { Button, Input, toast } from "@heroui/react";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { useRouter } from "@tanstack/react-router";
import { Eye, EyeOff, LockKeyhole, ShieldCheck, UserRound } from "lucide-react";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { normalizeUnknownError } from "../../shared/lib/http";
import { adminLogin } from "./api";
import { useAdminSessionStore } from "./store";

const adminLoginSchema = z.object({
  password: z.string().min(1, "请输入密码"),
  username: z.string().min(1, "请输入用户名"),
});

type AdminLoginValues = z.infer<typeof adminLoginSchema>;

export function LoginForm() {
  const router = useRouter();
  const setSession = useAdminSessionStore((state) => state.setSession);
  const [isPasswordVisible, setIsPasswordVisible] = useState(false);
  const form = useForm<AdminLoginValues>({
    defaultValues: {
      password: "",
      username: "",
    },
    resolver: zodResolver(adminLoginSchema),
  });

  const mutation = useMutation({
    mutationFn: adminLogin,
    onError: (cause) => {
      const error = normalizeUnknownError(cause);

      toast.danger("登录失败", {
        description: error.requestId
          ? `${error.userMessage}（requestId: ${error.requestId}）`
          : error.userMessage,
      });
    },
    onSuccess: async (session) => {
      setSession(session);
      toast.success("管理员会话已建立");
      await router.navigate({ to: "/" });
    },
  });

  return (
    <div className="w-full max-w-[430px]">
      <div className="mb-8 space-y-5">
        <div className="inline-flex h-14 w-14 items-center justify-center rounded-[18px] bg-brand-soft text-brand shadow-[0_18px_42px_-28px_rgba(37,99,235,0.7)]">
          <ShieldCheck className="h-8 w-8" />
        </div>
        <div className="space-y-2">
          <h1 className="text-[24px] leading-[32px] font-semibold">Bifrost Admin</h1>
          <p className="max-w-[300px] text-[15px] leading-[24px] text-text-secondary">
            登录安全控制台，管理用户、策略、服务与审计。
          </p>
        </div>
      </div>

      <form
        className="space-y-5"
        onSubmit={form.handleSubmit((values) => {
          mutation.mutate(values);
        })}
      >
        <div className="space-y-2">
          <label
            className="block text-[13px] leading-[18px] font-medium text-text-secondary"
            htmlFor="admin-login-username"
          >
            用户名
          </label>
          <div className="relative">
            <UserRound
              aria-hidden="true"
              className="pointer-events-none absolute left-4 top-1/2 h-5 w-5 -translate-y-1/2 text-text-muted"
            />
            <Input
              aria-invalid={Boolean(form.formState.errors.username)}
              autoComplete="username"
              className="h-[48px] w-full rounded-[12px] bg-white pl-12 pr-4 text-[15px] shadow-[0_12px_30px_-26px_rgba(15,23,42,0.42)] ring-1 ring-inset ring-[color-mix(in_oklab,var(--bifrost-border)_76%,var(--bifrost-brand)_14%)] placeholder:text-text-muted"
              fullWidth
              id="admin-login-username"
              placeholder="admin"
              {...form.register("username")}
            />
          </div>
          {form.formState.errors.username ? (
            <p className="text-[12px] leading-[18px] text-danger">
              {form.formState.errors.username.message}
            </p>
          ) : null}
        </div>

        <div className="space-y-2">
          <label
            className="block text-[13px] leading-[18px] font-medium text-text-secondary"
            htmlFor="admin-login-password"
          >
            密码
          </label>
          <div className="relative">
            <LockKeyhole
              aria-hidden="true"
              className="pointer-events-none absolute left-4 top-1/2 h-5 w-5 -translate-y-1/2 text-text-muted"
            />
            <Input
              aria-invalid={Boolean(form.formState.errors.password)}
              autoComplete="current-password"
              className="h-[48px] w-full rounded-[12px] bg-white pl-12 pr-12 text-[15px] shadow-[0_12px_30px_-26px_rgba(15,23,42,0.42)] ring-1 ring-inset ring-[color-mix(in_oklab,var(--bifrost-border)_76%,var(--bifrost-brand)_14%)] placeholder:text-text-muted"
              fullWidth
              id="admin-login-password"
              placeholder="请输入管理员密码"
              type={isPasswordVisible ? "text" : "password"}
              {...form.register("password")}
            />
            <Button
              aria-label={isPasswordVisible ? "隐藏密码" : "显示密码"}
              className="absolute right-2 top-1/2 h-8 w-8 min-w-0 -translate-y-1/2 rounded-[8px] p-0 text-text-muted hover:text-text-secondary"
              isIconOnly
              onPress={() => {
                setIsPasswordVisible((value) => !value);
              }}
              size="sm"
              type="button"
              variant="ghost"
            >
              {isPasswordVisible ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
            </Button>
          </div>
          {form.formState.errors.password ? (
            <p className="text-[12px] leading-[18px] text-danger">
              {form.formState.errors.password.message}
            </p>
          ) : null}
        </div>

        <Button
          className="mt-3 h-[52px] w-full rounded-[10px] bg-brand text-[16px] font-semibold text-white shadow-[0_18px_38px_-22px_rgba(37,99,235,0.85)] hover:bg-brand-hover"
          isDisabled={mutation.isPending}
          size="lg"
          type="submit"
        >
          {mutation.isPending ? "登录中..." : "登录后台"}
        </Button>
      </form>
    </div>
  );
}
