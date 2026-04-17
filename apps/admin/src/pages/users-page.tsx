import { Button, Dialog, EmptyState, ErrorState, Input, Table } from "@bifrost/ui";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useMemo, useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

import {
  createAdminUser,
  listAdminRoles,
  listAdminUsers,
  requireAccessToken,
} from "../entities/admin/api";
import { getCurrentAdminSession } from "../features/auth/store";
import { formatList, formatStatusLabel } from "../shared/lib/format";
import { normalizeUnknownError } from "../shared/lib/http";
import { StatusBadge } from "../shared/ui/status-badge";

const createUserSchema = z.object({
  displayName: z.string().min(1, "请输入显示名"),
  email: z.email("请输入有效邮箱地址"),
  password: z.string().min(8, "密码至少需要 8 位"),
  roleIds: z.array(z.string()).min(1, "请至少选择一个角色"),
  username: z.string().min(1, "请输入用户名"),
});

type CreateUserValues = z.infer<typeof createUserSchema>;

export function UsersPage() {
  const session = getCurrentAdminSession();
  const accessToken = requireAccessToken(session);
  const queryClient = useQueryClient();
  const [keyword, setKeyword] = useState("");
  const [status, setStatus] = useState("");
  const [createOpen, setCreateOpen] = useState(false);

  const usersQuery = useQuery({
    queryFn: () =>
      listAdminUsers({
        accessToken,
        keyword,
        status,
      }),
    queryKey: ["admin-users", accessToken, keyword, status],
  });

  const rolesQuery = useQuery({
    queryFn: () => listAdminRoles({ accessToken }),
    queryKey: ["admin-roles", accessToken],
  });

  const createUserForm = useForm<CreateUserValues>({
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
      setCreateOpen(false);
      createUserForm.reset();
      await queryClient.invalidateQueries({ queryKey: ["admin-users"] });
    },
  });

  const createUserError = createUserMutation.error
    ? normalizeUnknownError(createUserMutation.error)
    : null;

  const roleOptions = rolesQuery.data?.items ?? [];
  const userRows = usersQuery.data?.items ?? [];
  const totalUsers = usersQuery.data?.total ?? 0;
  const selectedRoleIDs = createUserForm.watch("roleIds");

  const resultCaption = useMemo(() => {
    if (keyword || status) {
      return `当前筛选命中 ${totalUsers} 个用户`;
    }

    return `当前共有 ${totalUsers} 个用户`;
  }, [keyword, status, totalUsers]);

  return (
    <div className="space-y-4">
      <section className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-[16px] leading-[24px] font-semibold">用户管理</h1>
          <p className="text-[12px] leading-[18px] text-text-secondary">
            管理用户账号、角色分配和登录入口。
          </p>
        </div>
        <Dialog.Root onOpenChange={setCreateOpen} open={createOpen}>
          <Dialog.Trigger asChild>
            <Button size="md">创建用户</Button>
          </Dialog.Trigger>
          <Dialog.Content>
            <Dialog.Header>
              <Dialog.Title>创建用户</Dialog.Title>
              <Dialog.Description>
                创建后台或客户端可使用的新账号，并分配初始角色。
              </Dialog.Description>
            </Dialog.Header>

            <form
              className="mt-4 space-y-4"
              onSubmit={createUserForm.handleSubmit(async (values) => {
                await createUserMutation.mutateAsync(values);
              })}
            >
              <div className="grid gap-4 md:grid-cols-2">
                <label className="space-y-1.5" htmlFor="create-user-username">
                  <span className="text-[12px] leading-[18px] text-text-secondary">用户名</span>
                  <Input
                    id="create-user-username"
                    placeholder="alice"
                    {...createUserForm.register("username")}
                  />
                  {createUserForm.formState.errors.username ? (
                    <p className="text-[12px] leading-[18px] text-danger">
                      {createUserForm.formState.errors.username.message}
                    </p>
                  ) : null}
                </label>
                <label className="space-y-1.5" htmlFor="create-user-display-name">
                  <span className="text-[12px] leading-[18px] text-text-secondary">显示名</span>
                  <Input
                    id="create-user-display-name"
                    placeholder="Alice"
                    {...createUserForm.register("displayName")}
                  />
                  {createUserForm.formState.errors.displayName ? (
                    <p className="text-[12px] leading-[18px] text-danger">
                      {createUserForm.formState.errors.displayName.message}
                    </p>
                  ) : null}
                </label>
              </div>

              <label className="space-y-1.5" htmlFor="create-user-email">
                <span className="text-[12px] leading-[18px] text-text-secondary">邮箱</span>
                <Input
                  id="create-user-email"
                  placeholder="alice@example.com"
                  {...createUserForm.register("email")}
                />
                {createUserForm.formState.errors.email ? (
                  <p className="text-[12px] leading-[18px] text-danger">
                    {createUserForm.formState.errors.email.message}
                  </p>
                ) : null}
              </label>

              <label className="space-y-1.5" htmlFor="create-user-password">
                <span className="text-[12px] leading-[18px] text-text-secondary">初始密码</span>
                <Input
                  id="create-user-password"
                  placeholder="ChangeMe123!"
                  type="password"
                  {...createUserForm.register("password")}
                />
                {createUserForm.formState.errors.password ? (
                  <p className="text-[12px] leading-[18px] text-danger">
                    {createUserForm.formState.errors.password.message}
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
                            const current = createUserForm.getValues("roleIds");
                            createUserForm.setValue(
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
                {createUserForm.formState.errors.roleIds ? (
                  <p className="text-[12px] leading-[18px] text-danger">
                    {createUserForm.formState.errors.roleIds.message}
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
                    setCreateOpen(false);
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
      </section>

      <section className="rounded-[14px] border border-border bg-surface p-4">
        <div className="flex flex-wrap items-center gap-2">
          <Input
            className="max-w-[280px]"
            onChange={(event) => {
              setKeyword(event.target.value);
            }}
            placeholder="搜索用户名、显示名或邮箱"
            value={keyword}
          />
          <select
            className="h-[32px] rounded-[6px] border border-border bg-surface px-3 text-[13px] leading-[20px] text-text-primary"
            onChange={(event) => {
              setStatus(event.target.value);
            }}
            value={status}
          >
            <option value="">全部状态</option>
            <option value="enabled">Enabled</option>
            <option value="disabled">Disabled</option>
          </select>
          {(keyword || status) && (
            <Button
              onClick={() => {
                setKeyword("");
                setStatus("");
              }}
              size="sm"
              variant="ghost"
            >
              清空筛选
            </Button>
          )}
        </div>
      </section>

      <section className="overflow-hidden rounded-[14px] border border-border bg-surface">
        {usersQuery.isLoading ? (
          <div className="px-4 py-8 text-[13px] leading-[20px] text-text-secondary">
            正在加载用户列表...
          </div>
        ) : userRows.length === 0 ? (
          <div className="px-4 py-8">
            <EmptyState
              description="当前条件下没有可展示的用户记录。"
              title={keyword || status ? "未匹配到用户" : "暂无用户"}
            />
          </div>
        ) : (
          <Table.Root>
            <Table.Caption>{resultCaption}</Table.Caption>
            <Table.Header>
              <Table.Row>
                <Table.Head>用户</Table.Head>
                <Table.Head>邮箱</Table.Head>
                <Table.Head>角色</Table.Head>
                <Table.Head>状态</Table.Head>
              </Table.Row>
            </Table.Header>
            <Table.Body>
              {userRows.map((user) => (
                <Table.Row key={user.id}>
                  <Table.Cell>
                    <div className="space-y-0.5">
                      <div className="font-medium">{user.displayName}</div>
                      <div className="text-[12px] leading-[18px] text-text-secondary">
                        {user.username}
                      </div>
                    </div>
                  </Table.Cell>
                  <Table.Cell>{user.email}</Table.Cell>
                  <Table.Cell>{formatList(user.roles)}</Table.Cell>
                  <Table.Cell>
                    <StatusBadge status={formatStatusLabel(user.status).toLowerCase()} />
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
