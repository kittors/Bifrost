import { Button, Input } from "@heroui/react";
import { Dialog } from "../../shared/ui/dialog";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

import { createAdminRole } from "../../entities/admin/api";

const createRoleSchema = z.object({
  description: z.string().min(1, "请输入角色描述"),
  displayName: z.string().min(1, "请输入显示名"),
  name: z.string().min(1, "请输入角色名"),
});

type CreateRoleValues = z.infer<typeof createRoleSchema>;

type CreateRoleDialogProps = {
  accessToken: string;
  onCreated: () => Promise<void> | void;
  onOpenChange: (open: boolean) => void;
  open: boolean;
};

// 创建角色表单与页面列表解耦，页面只负责控制弹窗开关和刷新时机。
export function CreateRoleDialog({
  accessToken,
  onCreated,
  onOpenChange,
  open,
}: CreateRoleDialogProps) {
  const form = useForm<CreateRoleValues>({
    defaultValues: {
      description: "",
      displayName: "",
      name: "",
    },
    resolver: zodResolver(createRoleSchema),
  });

  const createRoleMutation = useMutation({
    mutationFn: (values: CreateRoleValues) =>
      createAdminRole({
        accessToken,
        ...values,
      }),
    onSuccess: async () => {
      toast.success("角色已创建");
      form.reset();
      onOpenChange(false);
      await onCreated();
    },
  });

  return (
    <Dialog.Root onOpenChange={onOpenChange} open={open}>
      <Dialog.Trigger asChild>
        <Button>创建角色</Button>
      </Dialog.Trigger>
      <Dialog.Content>
        <Dialog.Header>
          <Dialog.Title>创建角色</Dialog.Title>
          <Dialog.Description>为一组用户建立统一的服务访问语义。</Dialog.Description>
        </Dialog.Header>

        <form
          className="mt-4 space-y-4"
          onSubmit={form.handleSubmit(async (values) => {
            await createRoleMutation.mutateAsync(values);
          })}
        >
          <label className="block space-y-1.5" htmlFor="create-role-name">
            <span className="text-[12px] leading-[18px] text-text-secondary">角色名</span>
            <Input id="create-role-name" placeholder="developer" {...form.register("name")} />
          </label>
          <label className="block space-y-1.5" htmlFor="create-role-display-name">
            <span className="text-[12px] leading-[18px] text-text-secondary">显示名</span>
            <Input
              id="create-role-display-name"
              placeholder="研发"
              {...form.register("displayName")}
            />
          </label>
          <label className="block space-y-1.5" htmlFor="create-role-description">
            <span className="text-[12px] leading-[18px] text-text-secondary">描述</span>
            <Input
              id="create-role-description"
              placeholder="研发相关私有服务"
              {...form.register("description")}
            />
          </label>

          <Dialog.Footer>
            <Button
              onClick={() => {
                onOpenChange(false);
              }}
              variant="secondary"
            >
              取消
            </Button>
            <Button isDisabled={createRoleMutation.isPending} type="submit">
              {createRoleMutation.isPending ? "提交中..." : "创建角色"}
            </Button>
          </Dialog.Footer>
        </form>
      </Dialog.Content>
    </Dialog.Root>
  );
}
