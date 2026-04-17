import { Button, Dialog, ErrorState, Input } from "@bifrost/ui";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

import { createAdminUser } from "../../entities/admin/api";
import type { AdminRole } from "../../entities/admin/types";
import { normalizeUnknownError } from "../../shared/lib/http";

const createUserSchema = z.object({
  displayName: z.string().min(1, "请输入显示名"),
  email: z.email("请输入有效邮箱地址"),
  password: z.string().min(8, "密码至少需要 8 位"),
  roleIds: z.array(z.string()).min(1, "请至少选择一个角色"),
  username: z.string().min(1, "请输入用户名"),
});

type CreateUserValues = z.infer<typeof createUserSchema>;

type CreateUserDialogProps = {
  accessToken: string;
  onCreated: () => Promise<void> | void;
  onOpenChange: (open: boolean) => void;
  open: boolean;
  roleOptions: AdminRole[];
};

// 将创建用户的表单和提交逻辑独立出去，避免页面组件同时承担列表和表单职责。
export function CreateUserDialog({
  accessToken,
  onCreated,
  onOpenChange,
  open,
  roleOptions,
}: CreateUserDialogProps) {
  const form = useForm<CreateUserValues>({
    defaultValues: {
      displayName: "",
      email: "",
      password: "",
      roleIds: [],
      username: "",
    },
    resolver: zodResolver(createUserSchema),
  });

  const createUserMutation = useMutation({
    mutationFn: (values: CreateUserValues) =>
      createAdminUser({
        accessToken,
        ...values,
      }),
    onSuccess: async () => {
      toast.success("用户已创建");
      form.reset();
      onOpenChange(false);
      await onCreated();
    },
  });

  const createUserError = createUserMutation.error
    ? normalizeUnknownError(createUserMutation.error)
    : null;
  const selectedRoleIDs = form.watch("roleIds");

  return (
    <Dialog.Root onOpenChange={onOpenChange} open={open}>
      <Dialog.Trigger asChild>
        <Button size="md">创建用户</Button>
      </Dialog.Trigger>
      <Dialog.Content>
        <Dialog.Header>
          <Dialog.Title>创建用户</Dialog.Title>
          <Dialog.Description>创建后台或客户端可使用的新账号，并分配初始角色。</Dialog.Description>
        </Dialog.Header>

        <form
          className="mt-4 space-y-4"
          onSubmit={form.handleSubmit(async (values) => {
            await createUserMutation.mutateAsync(values);
          })}
        >
          <div className="grid gap-4 md:grid-cols-2">
            <label className="space-y-1.5" htmlFor="create-user-username">
              <span className="text-[12px] leading-[18px] text-text-secondary">用户名</span>
              <Input id="create-user-username" placeholder="alice" {...form.register("username")} />
              {form.formState.errors.username ? (
                <p className="text-[12px] leading-[18px] text-danger">
                  {form.formState.errors.username.message}
                </p>
              ) : null}
            </label>
            <label className="space-y-1.5" htmlFor="create-user-display-name">
              <span className="text-[12px] leading-[18px] text-text-secondary">显示名</span>
              <Input
                id="create-user-display-name"
                placeholder="Alice"
                {...form.register("displayName")}
              />
              {form.formState.errors.displayName ? (
                <p className="text-[12px] leading-[18px] text-danger">
                  {form.formState.errors.displayName.message}
                </p>
              ) : null}
            </label>
          </div>

          <label className="space-y-1.5" htmlFor="create-user-email">
            <span className="text-[12px] leading-[18px] text-text-secondary">邮箱</span>
            <Input
              id="create-user-email"
              placeholder="alice@example.com"
              {...form.register("email")}
            />
            {form.formState.errors.email ? (
              <p className="text-[12px] leading-[18px] text-danger">
                {form.formState.errors.email.message}
              </p>
            ) : null}
          </label>

          <label className="space-y-1.5" htmlFor="create-user-password">
            <span className="text-[12px] leading-[18px] text-text-secondary">初始密码</span>
            <Input
              id="create-user-password"
              placeholder="ChangeMe123!"
              type="password"
              {...form.register("password")}
            />
            {form.formState.errors.password ? (
              <p className="text-[12px] leading-[18px] text-danger">
                {form.formState.errors.password.message}
              </p>
            ) : null}
          </label>

          <div className="space-y-2">
            <div className="text-[12px] leading-[18px] text-text-secondary">角色</div>
            <div className="grid gap-2 md:grid-cols-2">
              {roleOptions.map((role) => {
                const checked = selectedRoleIDs.includes(role.id);

                return (
                  <label
                    className="flex items-start gap-3 rounded-[10px] border border-border bg-surface-2 px-3 py-2"
                    key={role.id}
                  >
                    <input
                      checked={checked}
                      className="mt-1 h-4 w-4 accent-[var(--bifrost-brand)]"
                      onChange={(event) => {
                        const current = form.getValues("roleIds");

                        // 角色多选直接在当前表单状态上增删，避免额外引入重复状态源。
                        form.setValue(
                          "roleIds",
                          event.target.checked
                            ? [...current, role.id]
                            : current.filter((item) => item !== role.id),
                          { shouldValidate: true },
                        );
                      }}
                      type="checkbox"
                    />
                    <div className="min-w-0">
                      <div className="text-[13px] leading-[20px] font-medium">
                        {role.displayName}
                      </div>
                      <div className="text-[12px] leading-[18px] text-text-secondary">
                        {role.name}
                      </div>
                    </div>
                  </label>
                );
              })}
            </div>
            {form.formState.errors.roleIds ? (
              <p className="text-[12px] leading-[18px] text-danger">
                {form.formState.errors.roleIds.message}
              </p>
            ) : null}
          </div>

          {createUserError ? (
            <ErrorState
              description={createUserError.userMessage}
              requestId={createUserError.requestId || undefined}
              title="创建失败"
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
            <Button disabled={createUserMutation.isPending} type="submit">
              {createUserMutation.isPending ? "提交中..." : "创建用户"}
            </Button>
          </Dialog.Footer>
        </form>
      </Dialog.Content>
    </Dialog.Root>
  );
}
