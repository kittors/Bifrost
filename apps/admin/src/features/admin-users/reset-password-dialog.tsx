import { Button, Dialog, ErrorState, Input } from "@bifrost/ui";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

import { resetAdminUserPassword } from "../../entities/admin/api";
import type { AdminUser } from "../../entities/admin/types";
import { normalizeUnknownError } from "../../shared/lib/http";

const resetPasswordSchema = z.object({
  password: z.string().min(8, "密码至少需要 8 位"),
});

type ResetPasswordValues = z.infer<typeof resetPasswordSchema>;

type ResetPasswordDialogProps = {
  accessToken: string;
  onOpenChange: (open: boolean) => void;
  onReset: () => Promise<void> | void;
  open: boolean;
  user: AdminUser | null;
};

// 重置密码单独拆出，避免用户详情抽屉同时承担敏感表单状态。
export function ResetPasswordDialog({
  accessToken,
  onOpenChange,
  onReset,
  open,
  user,
}: ResetPasswordDialogProps) {
  const form = useForm<ResetPasswordValues>({
    defaultValues: {
      password: "",
    },
    resolver: zodResolver(resetPasswordSchema),
  });

  const resetPasswordMutation = useMutation({
    mutationFn: (values: ResetPasswordValues) =>
      resetAdminUserPassword({
        accessToken,
        password: values.password,
        userID: user?.id ?? "",
      }),
    onSuccess: async () => {
      toast.success("密码已重置");
      form.reset();
      onOpenChange(false);
      await onReset();
    },
  });

  const resetPasswordError = resetPasswordMutation.error
    ? normalizeUnknownError(resetPasswordMutation.error)
    : null;

  return (
    <Dialog.Root onOpenChange={onOpenChange} open={open}>
      <Dialog.Content>
        <Dialog.Header>
          <Dialog.Title>重置密码</Dialog.Title>
          <Dialog.Description>
            为 {user?.displayName ?? "当前用户"} 设置新密码，并立即撤销其已有会话。
          </Dialog.Description>
        </Dialog.Header>

        <form
          className="mt-4 space-y-4"
          onSubmit={form.handleSubmit(async (values) => {
            await resetPasswordMutation.mutateAsync(values);
          })}
        >
          <label className="space-y-1.5" htmlFor="reset-user-password">
            <span className="text-[12px] leading-[18px] text-text-secondary">新密码</span>
            <Input
              id="reset-user-password"
              placeholder="NewPassword123!"
              type="password"
              {...form.register("password")}
            />
            {form.formState.errors.password ? (
              <p className="text-[12px] leading-[18px] text-danger">
                {form.formState.errors.password.message}
              </p>
            ) : null}
          </label>

          {resetPasswordError ? (
            <ErrorState
              description={resetPasswordError.userMessage}
              requestId={resetPasswordError.requestId || undefined}
              title="密码重置失败"
            />
          ) : null}

          <Dialog.Footer>
            <Button
              onClick={() => {
                onOpenChange(false);
              }}
              variant="secondary"
            >
              取消
            </Button>
            <Button disabled={resetPasswordMutation.isPending || !user} type="submit">
              {resetPasswordMutation.isPending ? "提交中..." : "确认重置"}
            </Button>
          </Dialog.Footer>
        </form>
      </Dialog.Content>
    </Dialog.Root>
  );
}
