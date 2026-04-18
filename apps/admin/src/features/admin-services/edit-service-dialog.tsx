import { Button, Dialog, ErrorState, Input } from "@bifrost/ui";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { useEffect } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

import { getAdminService, updateAdminService } from "../../entities/admin/api";
import type { AdminService } from "../../entities/admin/types";
import { normalizeUnknownError } from "../../shared/lib/http";

const editServiceSchema = z.object({
  description: z.string().min(1, "请输入服务描述"),
  group: z.string().min(1, "请输入分组"),
  name: z.string().min(1, "请输入服务名称"),
  protocol: z.string().min(1, "请输入协议"),
  publicPath: z.string().min(1, "请输入公开路径"),
  upstreamUrl: z.string().url("请输入有效的上游地址"),
});

type EditServiceValues = z.infer<typeof editServiceSchema>;

type EditServiceDialogProps = {
  accessToken: string;
  onOpenChange: (open: boolean) => void;
  onSaved: () => Promise<void> | void;
  open: boolean;
  service: AdminService | null;
};

// 服务编辑弹窗独立维护表单状态，列表只负责触发。
export function EditServiceDialog({
  accessToken,
  onOpenChange,
  onSaved,
  open,
  service,
}: EditServiceDialogProps) {
  const form = useForm<EditServiceValues>({
    defaultValues: {
      description: "",
      group: "",
      name: "",
      protocol: "",
      publicPath: "",
      upstreamUrl: "",
    },
    resolver: zodResolver(editServiceSchema),
  });

  const serviceQuery = useQuery({
    enabled: open && Boolean(service),
    queryFn: () =>
      getAdminService({
        accessToken,
        serviceID: service?.id ?? "",
      }),
    queryKey: ["admin-service-detail", accessToken, service?.id],
  });

  useEffect(() => {
    if (serviceQuery.data) {
      form.reset({
        description: serviceQuery.data.description,
        group: serviceQuery.data.group,
        name: serviceQuery.data.name,
        protocol: serviceQuery.data.protocol,
        publicPath: serviceQuery.data.publicPath,
        upstreamUrl: serviceQuery.data.upstreamUrl,
      });
    }
  }, [form, serviceQuery.data]);

  const updateServiceMutation = useMutation({
    mutationFn: (values: EditServiceValues) =>
      updateAdminService({
        accessToken,
        serviceID: service?.id ?? "",
        ...values,
      }),
    onSuccess: async () => {
      toast.success("服务已更新");
      onOpenChange(false);
      await onSaved();
    },
  });

  const serviceError = serviceQuery.error ? normalizeUnknownError(serviceQuery.error) : null;
  const updateServiceError = updateServiceMutation.error
    ? normalizeUnknownError(updateServiceMutation.error)
    : null;

  return (
    <Dialog.Root onOpenChange={onOpenChange} open={open}>
      <Dialog.Content>
        <Dialog.Header>
          <Dialog.Title>编辑服务</Dialog.Title>
          <Dialog.Description>维护服务名称、上游地址和对外公开路径。</Dialog.Description>
        </Dialog.Header>

        <form
          className="mt-4 space-y-4"
          onSubmit={form.handleSubmit(async (values) => {
            await updateServiceMutation.mutateAsync(values);
          })}
        >
          <div className="grid gap-4 md:grid-cols-2">
            <Field label="服务名称">
              <Input placeholder="Docs Portal" {...form.register("name")} />
            </Field>
            <Field label="服务分组">
              <Input placeholder="shared" {...form.register("group")} />
            </Field>
          </div>

          <Field label="服务描述">
            <Input placeholder="共享文档服务" {...form.register("description")} />
          </Field>

          <div className="grid gap-4 md:grid-cols-2">
            <Field label="协议">
              <Input placeholder="http" {...form.register("protocol")} />
            </Field>
            <Field label="公开路径">
              <Input placeholder="/s/docs" {...form.register("publicPath")} />
            </Field>
          </div>

          <Field label="上游地址">
            <Input placeholder="http://mock-docs:8080" {...form.register("upstreamUrl")} />
          </Field>

          {serviceError ? (
            <ErrorState
              description={serviceError.userMessage}
              requestId={serviceError.requestId || undefined}
              title="服务详情加载失败"
            />
          ) : null}

          {updateServiceError ? (
            <ErrorState
              description={updateServiceError.userMessage}
              requestId={updateServiceError.requestId || undefined}
              title="服务更新失败"
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
            <Button disabled={updateServiceMutation.isPending || !service} type="submit">
              {updateServiceMutation.isPending ? "提交中..." : "保存变更"}
            </Button>
          </Dialog.Footer>
        </form>
      </Dialog.Content>
    </Dialog.Root>
  );
}

function Field({ children, label }: { children: ReactNode; label: string }) {
  return (
    <div className="space-y-1.5">
      <span className="text-[12px] leading-[18px] text-text-secondary">{label}</span>
      {children}
    </div>
  );
}
