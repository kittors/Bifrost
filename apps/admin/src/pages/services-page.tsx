import { Button, Dialog, EmptyState, Input, Table } from "@bifrost/ui";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useMemo, useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

import { createAdminService, listAdminServices, requireAccessToken } from "../entities/admin/api";
import { getCurrentAdminSession } from "../features/auth/store";
import { QueryErrorState } from "../shared/ui/query-error-state";
import { StatusBadge } from "../shared/ui/status-badge";

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

export function ServicesPage() {
  const session = getCurrentAdminSession();
  const accessToken = requireAccessToken(session);
  const queryClient = useQueryClient();
  const [keyword, setKeyword] = useState("");
  const [status, setStatus] = useState("");
  const [createOpen, setCreateOpen] = useState(false);

  const servicesQuery = useQuery({
    queryFn: () => listAdminServices({ accessToken, keyword, status }),
    queryKey: ["admin-services", accessToken, keyword, status],
  });

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
      setCreateOpen(false);
      form.reset();
      await queryClient.invalidateQueries({ queryKey: ["admin-services"] });
    },
  });

  const rows = servicesQuery.data?.items ?? [];
  const caption = useMemo(
    () => `当前共有 ${servicesQuery.data?.total ?? 0} 个服务`,
    [servicesQuery.data?.total],
  );

  return (
    <div className="space-y-4">
      <section className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-[16px] leading-[24px] font-semibold">服务目录</h1>
          <p className="text-[12px] leading-[18px] text-text-secondary">
            管理私有服务入口、上游地址和公开路径。
          </p>
        </div>

        <Dialog.Root onOpenChange={setCreateOpen} open={createOpen}>
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
                  onChange={(event) => form.setValue("enabled", event.target.checked)}
                  type="checkbox"
                />
                <span className="text-[13px] leading-[20px]">创建后立即启用</span>
              </label>

              <Dialog.Footer>
                <Button onClick={() => setCreateOpen(false)} variant="secondary">
                  取消
                </Button>
                <Button disabled={createServiceMutation.isPending} type="submit">
                  {createServiceMutation.isPending ? "提交中..." : "创建服务"}
                </Button>
              </Dialog.Footer>
            </form>
          </Dialog.Content>
        </Dialog.Root>
      </section>

      <section className="rounded-[14px] border border-border bg-surface p-4">
        <div className="flex flex-wrap items-center gap-2">
          <Input
            className="max-w-[280px]"
            onChange={(event) => setKeyword(event.target.value)}
            placeholder="搜索服务名或标识"
            value={keyword}
          />
          <select
            className="h-[32px] rounded-[6px] border border-border bg-surface px-3 text-[13px] leading-[20px]"
            onChange={(event) => setStatus(event.target.value)}
            value={status}
          >
            <option value="">全部状态</option>
            <option value="enabled">Enabled</option>
            <option value="disabled">Disabled</option>
          </select>
        </div>
      </section>

      {servicesQuery.error ? (
        <QueryErrorState error={servicesQuery.error} title="服务列表加载失败" />
      ) : null}

      <section className="overflow-hidden rounded-[14px] border border-border bg-surface">
        {rows.length === 0 && !servicesQuery.isLoading ? (
          <div className="px-4 py-8">
            <EmptyState description="当前没有服务记录。" title="暂无服务" />
          </div>
        ) : (
          <Table.Root>
            <Table.Caption>{caption}</Table.Caption>
            <Table.Header>
              <Table.Row>
                <Table.Head>服务</Table.Head>
                <Table.Head>上游地址</Table.Head>
                <Table.Head>状态</Table.Head>
              </Table.Row>
            </Table.Header>
            <Table.Body>
              {rows.map((service) => (
                <Table.Row key={service.id}>
                  <Table.Cell>
                    <div className="space-y-0.5">
                      <div className="font-medium">{service.name}</div>
                      <div className="text-[12px] leading-[18px] text-text-secondary">
                        {service.key} · {service.publicPath}
                      </div>
                    </div>
                  </Table.Cell>
                  <Table.Cell>{service.upstreamUrl}</Table.Cell>
                  <Table.Cell>
                    <StatusBadge status={service.status} />
                  </Table.Cell>
                </Table.Row>
              ))}
            </Table.Body>
          </Table.Root>
        )}
      </section>
    </div>
  );
}
