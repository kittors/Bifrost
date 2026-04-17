import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";

import { listAdminRoles, listAdminUsers, requireAccessToken } from "../entities/admin/api";
import { CreateUserDialog } from "../features/admin-users/create-user-dialog";
import { UsersFilterBar } from "../features/admin-users/users-filter-bar";
import { UsersTable } from "../features/admin-users/users-table";
import { getCurrentAdminSession } from "../features/auth/store";
import { QueryErrorState } from "../shared/ui/query-error-state";

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

  const roleOptions = rolesQuery.data?.items ?? [];
  const userRows = usersQuery.data?.items ?? [];
  const totalUsers = usersQuery.data?.total ?? 0;

  return (
    <div className="space-y-4">
      {/* 页面层只负责组装区块，不再承载表单与表格细节。 */}
      <section className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-[16px] leading-[24px] font-semibold">用户管理</h1>
          <p className="text-[12px] leading-[18px] text-text-secondary">
            管理用户账号、角色分配和登录入口。
          </p>
        </div>
        <CreateUserDialog
          accessToken={accessToken}
          onCreated={async () => {
            await queryClient.invalidateQueries({ queryKey: ["admin-users"] });
          }}
          onOpenChange={setCreateOpen}
          open={createOpen}
          roleOptions={roleOptions}
        />
      </section>

      <UsersFilterBar
        keyword={keyword}
        onKeywordChange={setKeyword}
        onReset={() => {
          setKeyword("");
          setStatus("");
        }}
        onStatusChange={setStatus}
        status={status}
      />

      {usersQuery.error ? (
        <QueryErrorState error={usersQuery.error} title="用户列表加载失败" />
      ) : null}

      {usersQuery.isLoading ? (
        <div className="rounded-[14px] border border-border bg-surface px-4 py-8 text-[13px] leading-[20px] text-text-secondary">
          正在加载用户列表...
        </div>
      ) : (
        <UsersTable keyword={keyword} rows={userRows} status={status} totalUsers={totalUsers} />
      )}
    </div>
  );
}
