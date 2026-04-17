import { Button, Dialog, Input } from "@bifrost/ui";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

import { createAdminService } from "../../entities/admin/api";

const createServiceSchema = z.object({
  description: z.string().min(1, "请输入服务描述"),
  enabled: z.boolean(),
  group: z.string().min(1, "请输入服务分组"),
  key: z.string().min(1, "请输入服务标识"),
  name: z.string().min(1, "请输入服务名称"),
  protocol: z.string().min(1, "请输入协议"),
  publicPath: z.string().min(1, "请输入公共路径"),
  upstreamUrl: z.string().url("请输入有效的上游地址"),
});

type CreateServiceValues = z.infer<typeof createServiceSchema>;

type CreateServiceDialogProps = {
  accessToken: string;
  onCreated: () => Promise<void> | void;
  onOpenChange: (open: boolean) => void;
  open: boolean;
};

// 服务创建弹窗独立维护自己的表单状态，页面只负责列表刷新。
export function CreateServiceDialog({
  accessToken,
  onCreated,
  onOpenChange,
  open,
}: CreateServiceDialogProps) {
  const form = useForm<CreateServiceValues>({
    defaultValues: {
      description: "",
      enabled: true,
      group: "development",
      key: "",
      name: "",
      protocol: "https",
      publicPath: "",
      upstreamUrl: "",
    },
    resolver: zodResolver(createServiceSchema),
  });

  const createServiceMutation = useMutation({
    mutationFn: (values: CreateServiceValues) =>
      createAdminService({
        accessToken,
        ...values,
      }),
    onSuccess: async () => {
      toast.success("服务已创建");
      form.reset();
      onOpenChange(false);
      await onCreated();
    },
  });

  return (
    <Dialog.Root onOpenChange={onOpenChange} open={open}>
      <Dialog.Trigger asChild>
        <Button>创建服务</Button>
      </Dialog.Trigger>
      <Dialog.Content>
        <Dialog.Header>
          <Dialog.Title>创建服务</Dialog.Title>
          <Dialog.Description>显式登记可被网关代理访问的私有 Web 服务。</Dialog.Description>
        </Dialog.Header>

        <form
          className="mt-4 space-y-4"
          onSubmit={form.handleSubmit(async (values) => {
            await createServiceMutation.mutateAsync(values);
          })}
        >
          <div className="grid gap-4 md:grid-cols-2">
            <label className="block space-y-1.5" htmlFor="create-service-key">
              <span className="text-[12px] leading-[18px] text-text-secondary">服务标识</span>
              <Input id="create-service-key" placeholder="gitlab" {...form.register("key")} />
            </label>
            <label className="block space-y-1.5" htmlFor="create-service-name">
              <span className="text-[12px] leading-[18px] text-text-secondary">服务名称</span>
              <Input id="create-service-name" placeholder="GitLab" {...form.register("name")} />
            </label>
          </div>

          <div className="grid gap-4 md:grid-cols-2">
            <label className="block space-y-1.5" htmlFor="create-service-group">
              <span className="text-[12px] leading-[18px] text-text-secondary">分组</span>
              <Input
                id="create-service-group"
                placeholder="development"
                {...form.register("group")}
              />
            </label>
            <label className="block space-y-1.5" htmlFor="create-service-protocol">
              <span className="text-[12px] leading-[18px] text-text-secondary">协议</span>
              <Input
                id="create-service-protocol"
                placeholder="https"
                {...form.register("protocol")}
              />
            </label>
          </div>

          <label className="block space-y-1.5" htmlFor="create-service-upstream">
            <span className="text-[12px] leading-[18px] text-text-secondary">上游地址</span>
            <Input
              id="create-service-upstream"
              placeholder="http://mock-gitlab:8080"
              {...form.register("upstreamUrl")}
            />
          </label>
          <label className="block space-y-1.5" htmlFor="create-service-path">
            <span className="text-[12px] leading-[18px] text-text-secondary">公共路径</span>
            <Input
              id="create-service-path"
              placeholder="/s/gitlab"
              {...form.register("publicPath")}
            />
          </label>
          <label className="block space-y-1.5" htmlFor="create-service-description">
            <span className="text-[12px] leading-[18px] text-text-secondary">描述</span>
            <Input
              id="create-service-description"
              placeholder="研发代码平台"
              {...form.register("description")}
            />
          </label>

          <label className="flex items-center gap-3 rounded-[10px] border border-border bg-surface-2 px-3 py-2">
            <input
              checked={form.watch("enabled")}
              className="h-4 w-4 accent-[var(--bifrost-brand)]"
              onChange={(event) => {
                form.setValue("enabled", event.target.checked);
              }}
              type="checkbox"
            />
            <span className="text-[13px] leading-[20px]">创建后立即启用</span>
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
            <Button disabled={createServiceMutation.isPending} type="submit">
              {createServiceMutation.isPending ? "提交中..." : "创建服务"}
            </Button>
          </Dialog.Footer>
        </form>
      </Dialog.Content>
    </Dialog.Root>
  );
}
