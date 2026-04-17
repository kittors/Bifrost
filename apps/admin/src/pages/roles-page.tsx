import { Button, Dialog, Drawer, EmptyState, Input, Table } from "@bifrost/ui";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useMemo, useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

import {
  createAdminRole,
  listAdminRoles,
  listAdminServices,
  replaceRoleServices,
  requireAccessToken,
} from "../entities/admin/api";
import type { AdminRole } from "../entities/admin/types";
import { getCurrentAdminSession } from "../features/auth/store";
import { QueryErrorState } from "../shared/ui/query-error-state";

const createRoleSchema = z.object({
  description: z.string().min(1, "请输入角色描述"),
  displayName: z.string().min(1, "请输入显示名"),
  name: z.string().min(1, "请输入角色名"),
});

type CreateRoleValues = z.infer<typeof createRoleSchema>;

export function RolesPage() {
  const session = getCurrentAdminSession();
  const accessToken = requireAccessToken(session);
  const queryClient = useQueryClient();
  const [keyword, setKeyword] = useState("");
  const [createOpen, setCreateOpen] = useState(false);
  const [permissionRole, setPermissionRole] = useState<AdminRole | null>(null);
  const [selectedServiceIDs, setSelectedServiceIDs] = useState<string[]>([]);

  const rolesQuery = useQuery({
    queryFn: () => listAdminRoles({ accessToken, keyword }),
    queryKey: ["admin-roles", accessToken, keyword],
  });
  const servicesQuery = useQuery({
    queryFn: () => listAdminServices({ accessToken, pageSize: 200 }),
    queryKey: ["admin-services", accessToken, "role-config"],
  });

  const createRoleForm = useForm<CreateRoleValues>({
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
      setCreateOpen(false);
      createRoleForm.reset();
      await queryClient.invalidateQueries({ queryKey: ["admin-roles"] });
    },
  });

  const replaceRoleServicesMutation = useMutation({
    mutationFn: () =>
      replaceRoleServices({
        accessToken,
        roleID: permissionRole?.id ?? "",
        serviceIDs: selectedServiceIDs,
      }),
    onSuccess: async () => {
      toast.success("角色服务授权已提交");
      setPermissionRole(null);
      setSelectedServiceIDs([]);
      await queryClient.invalidateQueries({ queryKey: ["admin-roles"] });
    },
  });

  const rows = rolesQuery.data?.items ?? [];
  const services = servicesQuery.data?.items ?? [];
  const caption = useMemo(
    () => `当前共有 ${rolesQuery.data?.total ?? 0} 个角色`,
    [rolesQuery.data?.total],
  );

  return (
    <div className="space-y-4">
      <section className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-[16px] leading-[24px] font-semibold">角色管理</h1>
          <p className="text-[12px] leading-[18px] text-text-secondary">
            维护角色清单，并对角色可访问服务做显式配置。
          </p>
        </div>

        <Dialog.Root onOpenChange={setCreateOpen} open={createOpen}>
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
              onSubmit={createRoleForm.handleSubmit(async (values) => {
                await createRoleMutation.mutateAsync(values);
              })}
            >
              <label className="block space-y-1.5" htmlFor="create-role-name">
                <span className="text-[12px] leading-[18px] text-text-secondary">角色名</span>
                <Input
                  id="create-role-name"
                  placeholder="developer"
                  {...createRoleForm.register("name")}
                />
              </label>
              <label className="block space-y-1.5" htmlFor="create-role-display-name">
                <span className="text-[12px] leading-[18px] text-text-secondary">显示名</span>
                <Input
                  id="create-role-display-name"
                  placeholder="研发"
                  {...createRoleForm.register("displayName")}
                />
              </label>
              <label className="block space-y-1.5" htmlFor="create-role-description">
                <span className="text-[12px] leading-[18px] text-text-secondary">描述</span>
                <Input
                  id="create-role-description"
                  placeholder="研发相关私有服务"
                  {...createRoleForm.register("description")}
                />
              </label>

              <Dialog.Footer>
                <Button onClick={() => setCreateOpen(false)} variant="secondary">
                  取消
                </Button>
                <Button disabled={createRoleMutation.isPending} type="submit">
                  {createRoleMutation.isPending ? "提交中..." : "创建角色"}
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
            placeholder="搜索角色名或显示名"
            value={keyword}
          />
          {keyword ? (
            <Button onClick={() => setKeyword("")} size="sm" variant="ghost">
              清空筛选
            </Button>
          ) : null}
        </div>
      </section>

      {rolesQuery.error ? (
        <QueryErrorState error={rolesQuery.error} title="角色列表加载失败" />
      ) : null}

      <section className="overflow-hidden rounded-[14px] border border-border bg-surface">
        {rows.length === 0 && !rolesQuery.isLoading ? (
          <div className="px-4 py-8">
            <EmptyState
              description="当前没有可展示的角色记录。"
              title={keyword ? "未匹配到角色" : "暂无角色"}
            />
          </div>
        ) : (
          <Table.Root>
            <Table.Caption>{caption}</Table.Caption>
            <Table.Header>
              <Table.Row>
                <Table.Head>角色</Table.Head>
                <Table.Head>描述</Table.Head>
                <Table.Head className="text-right">操作</Table.Head>
              </Table.Row>
            </Table.Header>
            <Table.Body>
              {rows.map((role) => (
                <Table.Row key={role.id}>
                  <Table.Cell>
                    <div className="space-y-0.5">
                      <div className="font-medium">{role.displayName}</div>
                      <div className="text-[12px] leading-[18px] text-text-secondary">
                        {role.name}
                      </div>
                    </div>
                  </Table.Cell>
                  <Table.Cell>{role.description}</Table.Cell>
                  <Table.Cell className="text-right">
                    <Button
                      onClick={() => {
                        setPermissionRole(role);
                        setSelectedServiceIDs([]);
                      }}
                      size="sm"
                      variant="secondary"
                    >
                      授权服务
                    </Button>
                  </Table.Cell>
                </Table.Row>
              ))}
            </Table.Body>
          </Table.Root>
        )}
      </section>

      <Drawer.Root
        onOpenChange={(open) => {
          if (!open) {
            setPermissionRole(null);
            setSelectedServiceIDs([]);
          }
        }}
        open={Boolean(permissionRole)}
      >
        <Drawer.Content className="w-[min(640px,calc(100vw-24px))]">
          <Drawer.Header>
            <Drawer.Title>{permissionRole?.displayName ?? "角色服务授权"}</Drawer.Title>
            <Drawer.Description>
              当前后端仍未提供角色服务摘要查询接口，这里会以你勾选的服务集合执行覆盖写入。
            </Drawer.Description>
          </Drawer.Header>

          {replaceRoleServicesMutation.error ? (
            <div className="mt-4">
              <QueryErrorState error={replaceRoleServicesMutation.error} title="角色授权提交失败" />
            </div>
          ) : null}

          <div className="mt-4 grid gap-2 overflow-y-auto">
            {services.map((service) => {
              const checked = selectedServiceIDs.includes(service.id);

              return (
                <label
                  className="flex items-start gap-3 rounded-[10px] border border-border bg-surface-2 px-3 py-2"
                  key={service.id}
                >
                  <input
                    checked={checked}
                    className="mt-1 h-4 w-4 accent-[var(--bifrost-brand)]"
                    onChange={(event) => {
                      setSelectedServiceIDs((current) =>
                        event.target.checked
                          ? [...current, service.id]
                          : current.filter((item) => item !== service.id),
                      );
                    }}
                    type="checkbox"
                  />
                  <div className="min-w-0">
                    <div className="text-[13px] leading-[20px] font-medium">{service.name}</div>
                    <div className="text-[12px] leading-[18px] text-text-secondary">
                      {service.key} · {service.group}
                    </div>
                  </div>
                </label>
              );
            })}
          </div>

          <Drawer.Footer>
            <Button onClick={() => setPermissionRole(null)} variant="secondary">
              取消
            </Button>
            <Button
              disabled={replaceRoleServicesMutation.isPending || !permissionRole}
              onClick={async () => {
                await replaceRoleServicesMutation.mutateAsync();
              }}
            >
              {replaceRoleServicesMutation.isPending ? "提交中..." : "保存授权"}
            </Button>
          </Drawer.Footer>
        </Drawer.Content>
      </Drawer.Root>
    </div>
  );
}
