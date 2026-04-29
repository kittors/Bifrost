import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";

import {
  listAdminRoles,
  listAdminServices,
  listAdminUsers,
  requireAccessToken,
} from "../entities/admin/api";
import type { AdminUser } from "../entities/admin/types";
import { CreateUserDialog } from "../features/admin-users/create-user-dialog";
import { UserDetailDrawer } from "../features/admin-users/user-detail-drawer";
import { UserServiceOverridesDrawer } from "../features/admin-users/user-service-overrides-drawer";
import { UsersFilterBar } from "../features/admin-users/users-filter-bar";
import { UsersTable } from "../features/admin-users/users-table";
import { getCurrentAdminSession } from "../features/auth/store";
import { PaginationBar } from "../shared/ui/pagination-bar";
import { QueryErrorState } from "../shared/ui/query-error-state";

const usersPageSize = 20;

export function UsersPage() {
  const session = getCurrentAdminSession();
  const accessToken = requireAccessToken(session);
  const queryClient = useQueryClient();
  const [keyword, setKeyword] = useState("");
  const [page, setPage] = useState(1);
  const [status, setStatus] = useState("");
  const [createOpen, setCreateOpen] = useState(false);
  const [selectedUserID, setSelectedUserID] = useState<string | null>(null);
  const [overrideUser, setOverrideUser] = useState<AdminUser | null>(null);

  const usersQuery = useQuery({
    queryFn: () =>
      listAdminUsers({
        accessToken,
        keyword,
        page,
        pageSize: usersPageSize,
        status,
      }),
    queryKey: ["admin-users", accessToken, keyword, page, status],
  });

  const rolesQuery = useQuery({
    queryFn: () => listAdminRoles({ accessToken }),
    queryKey: ["admin-roles", accessToken],
  });
  const servicesQuery = useQuery({
    queryFn: () => listAdminServices({ accessToken, pageSize: 200 }),
    queryKey: ["admin-services", accessToken, "user-overrides"],
  });

  const roleOptions = rolesQuery.data?.items ?? [];
  const serviceOptions = servicesQuery.data?.items ?? [];
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
        onKeywordChange={(value) => {
          setKeyword(value);
          setPage(1);
        }}
        onReset={() => {
          setKeyword("");
          setPage(1);
          setStatus("");
        }}
        onStatusChange={(value) => {
          setStatus(value);
          setPage(1);
        }}
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
        <UsersTable
          keyword={keyword}
          onOpenDetails={setSelectedUserID}
          onOpenOverrides={setOverrideUser}
          rows={userRows}
          status={status}
          totalUsers={totalUsers}
        />
      )}

      <PaginationBar
        onPageChange={setPage}
        page={page}
        pageSize={usersPageSize}
        total={totalUsers}
      />

      <UserDetailDrawer
        accessToken={accessToken}
        onOpenChange={(open) => {
          if (!open) {
            setSelectedUserID(null);
          }
        }}
        onUpdated={async () => {
          await queryClient.invalidateQueries({ queryKey: ["admin-users"] });
        }}
        open={Boolean(selectedUserID)}
        userID={selectedUserID}
      />

      <UserServiceOverridesDrawer
        accessToken={accessToken}
        onOpenChange={(open) => {
          if (!open) {
            setOverrideUser(null);
          }
        }}
        onSaved={async () => {
          await queryClient.invalidateQueries({ queryKey: ["admin-users"] });
        }}
        open={Boolean(overrideUser)}
        services={serviceOptions}
        user={overrideUser}
      />
    </div>
  );
}
