import { Button, Input, toast } from "@heroui/react";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";

import { updateAdminRole } from "../../entities/admin/api";
import type { AdminRole } from "../../entities/admin/types";
import { normalizeUnknownError } from "../../shared/lib/http";
import { ErrorState } from "../../shared/ui/error-state";
import { Dialog } from "../../shared/ui/dialog";

const editRoleSchema = z.object({
  description: z.string().min(1, "请输入角色描述"),
  displayName: z.string().min(1, "请输入显示名"),
});

type EditRoleValues = z.infer<typeof editRoleSchema>;

type EditRoleDialogProps = {
  accessToken: string;
  onOpenChange: (open: boolean) => void;
  onSaved: () => Promise<void> | void;
  open: boolean;
  role: AdminRole | null;
};

// 角色编辑弹窗只维护可编辑字段，角色名保持不可变以减少策略引用风险。
export function EditRoleDialog({
  accessToken,
  onOpenChange,
  onSaved,
  open,
  role,
}: EditRoleDialogProps) {
  const form = useForm<EditRoleValues>({
    defaultValues: {
      description: "",
      displayName: "",
    },
    resolver: zodResolver(editRoleSchema),
  });

  useEffect(() => {
    if (role) {
      form.reset({
        description: role.description,
        displayName: role.displayName,
      });
    }
  }, [form, role]);

  const updateRoleMutation = useMutation({
    mutationFn: (values: EditRoleValues) =>
      updateAdminRole({
        accessToken,
        roleID: role?.id ?? "",
        ...values,
      }),
    onSuccess: async () => {
      toast.success("角色已更新");
      onOpenChange(false);
      await onSaved();
    },
  });

  const updateRoleError = updateRoleMutation.error
    ? normalizeUnknownError(updateRoleMutation.error)
    : null;

  return (
    <Dialog.Root onOpenChange={onOpenChange} open={open}>
      <Dialog.Content>
        <Dialog.Header>
          <Dialog.Title>编辑角色</Dialog.Title>
          <Dialog.Description>
            更新 {role?.name ?? "当前角色"} 的展示名称和说明，不改变角色标识。
          </Dialog.Description>
        </Dialog.Header>

        <form
          className="mt-4 space-y-4"
          onSubmit={form.handleSubmit(async (values) => {
            await updateRoleMutation.mutateAsync(values);
          })}
        >
          <label className="block space-y-1.5" htmlFor="edit-role-display-name">
            <span className="text-[12px] leading-[18px] text-text-secondary">显示名</span>
            <Input
              id="edit-role-display-name"
              placeholder="研发团队"
              {...form.register("displayName")}
            />
          </label>

          <label className="block space-y-1.5" htmlFor="edit-role-description">
            <span className="text-[12px] leading-[18px] text-text-secondary">描述</span>
            <Input
              id="edit-role-description"
              placeholder="研发私有服务访问角色"
              {...form.register("description")}
            />
          </label>

          {updateRoleError ? (
            <ErrorState
              description={updateRoleError.userMessage}
              requestId={updateRoleError.requestId || undefined}
              title="角色更新失败"
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
            <Button isDisabled={updateRoleMutation.isPending || !role} type="submit">
              {updateRoleMutation.isPending ? "提交中..." : "保存角色"}
            </Button>
          </Dialog.Footer>
        </form>
      </Dialog.Content>
    </Dialog.Root>
  );
}
