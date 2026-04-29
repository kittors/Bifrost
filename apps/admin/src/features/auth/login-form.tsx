import { Button, Input } from "@heroui/react";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { useRouter } from "@tanstack/react-router";
import { ShieldCheck } from "lucide-react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

import { normalizeUnknownError } from "../../shared/lib/http";
import { ErrorState } from "../../shared/ui/error-state";
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
  const form = useForm<AdminLoginValues>({
    defaultValues: {
      password: "",
      username: "",
    },
    resolver: zodResolver(adminLoginSchema),
  });

  const mutation = useMutation({
    mutationFn: adminLogin,
    onSuccess: async (session) => {
      setSession(session);
      toast.success("管理员会话已建立");
      await router.navigate({ to: "/" });
    },
  });

  const error = mutation.error ? normalizeUnknownError(mutation.error) : null;

  return (
    <div className="w-full max-w-[420px] rounded-[14px] border border-border bg-surface p-6 shadow-[0_28px_80px_-36px_rgba(15,23,42,0.35)]">
      <div className="mb-6 flex items-start justify-between gap-4">
        <div className="space-y-2">
          <div className="inline-flex h-10 w-10 items-center justify-center rounded-[10px] bg-brand-soft text-brand">
            <ShieldCheck className="h-5 w-5" />
          </div>
          <div className="space-y-1">
            <h1 className="text-[16px] leading-[24px] font-semibold">Bifrost Admin</h1>
            <p className="text-[13px] leading-[20px] text-text-secondary">
              登录安全控制台，管理用户、策略、服务与审计。
            </p>
          </div>
        </div>
        <div className="rounded-full border border-border bg-surface-2 px-2 py-1 text-[12px] leading-[18px] text-text-secondary">
          Secure Console
        </div>
      </div>

      <form
        className="space-y-4"
        onSubmit={form.handleSubmit(async (values) => {
          await mutation.mutateAsync(values);
        })}
      >
        <label className="block space-y-1.5" htmlFor="admin-login-username">
          <span className="text-[12px] leading-[18px] font-medium text-text-secondary">用户名</span>
          <Input
            autoComplete="username"
            id="admin-login-username"
            placeholder="admin"
            {...form.register("username")}
          />
          {form.formState.errors.username ? (
            <p className="text-[12px] leading-[18px] text-danger">
              {form.formState.errors.username.message}
            </p>
          ) : null}
        </label>

        <label className="block space-y-1.5" htmlFor="admin-login-password">
          <span className="text-[12px] leading-[18px] font-medium text-text-secondary">密码</span>
          <Input
            autoComplete="current-password"
            id="admin-login-password"
            placeholder="请输入管理员密码"
            type="password"
            {...form.register("password")}
          />
          {form.formState.errors.password ? (
            <p className="text-[12px] leading-[18px] text-danger">
              {form.formState.errors.password.message}
            </p>
          ) : null}
        </label>

        {error ? (
          <ErrorState
            description={error.userMessage}
            requestId={error.requestId || undefined}
            title="登录失败"
          />
        ) : null}

        <Button className="w-full" isDisabled={mutation.isPending} size="lg" type="submit">
          {mutation.isPending ? "登录中..." : "登录后台"}
        </Button>
      </form>
    </div>
  );
}
