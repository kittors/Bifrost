import { Pagination } from "@heroui/react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";

import { listAdminRoles, listAdminServices, requireAccessToken } from "../entities/admin/api";
import type { AdminRole } from "../entities/admin/types";
import { CreateRoleDialog } from "../features/admin-roles/create-role-dialog";
import { EditRoleDialog } from "../features/admin-roles/edit-role-dialog";
import { RoleServicesDrawer } from "../features/admin-roles/role-services-drawer";
import { RolesFilterBar } from "../features/admin-roles/roles-filter-bar";
import { RolesTable } from "../features/admin-roles/roles-table";
import { getCurrentAdminSession } from "../features/auth/store";
import { QueryErrorState } from "../shared/ui/query-error-state";

const rolesPageSize = 20;

export function RolesPage() {
  const session = getCurrentAdminSession();
  const accessToken = requireAccessToken(session);
  const queryClient = useQueryClient();
  const [keyword, setKeyword] = useState("");
  const [page, setPage] = useState(1);
  const [createOpen, setCreateOpen] = useState(false);
  const [editingRole, setEditingRole] = useState<AdminRole | null>(null);
  const [permissionRole, setPermissionRole] = useState<AdminRole | null>(null);

  const rolesQuery = useQuery({
    queryFn: () => listAdminRoles({ accessToken, keyword, page, pageSize: rolesPageSize }),
    queryKey: ["admin-roles", accessToken, keyword, page],
  });
  const servicesQuery = useQuery({
    queryFn: () => listAdminServices({ accessToken, pageSize: 200 }),
    queryKey: ["admin-services", accessToken, "role-config"],
  });

  const rows = rolesQuery.data?.items ?? [];
  const services = servicesQuery.data?.items ?? [];
  const totalRoles = rolesQuery.data?.total ?? 0;
  const pageCount = Math.max(1, Math.ceil(totalRoles / rolesPageSize));
  const safePage = Math.min(Math.max(page, 1), pageCount);

  return (
    <div className="space-y-4">
      {/* 页面层保留查询与区块编排，弹窗和抽屉细节全部下沉到 feature 目录。 */}
      <section className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-[16px] leading-[24px] font-semibold">角色管理</h1>
          <p className="text-[12px] leading-[18px] text-text-secondary">
            维护角色清单，并对角色可访问服务做显式配置。
          </p>
        </div>

        <CreateRoleDialog
          accessToken={accessToken}
          onCreated={async () => {
            await queryClient.invalidateQueries({ queryKey: ["admin-roles"] });
          }}
          onOpenChange={setCreateOpen}
          open={createOpen}
        />
      </section>

      <RolesFilterBar
        keyword={keyword}
        onKeywordChange={(value) => {
          setKeyword(value);
          setPage(1);
        }}
        onReset={() => {
          setKeyword("");
          setPage(1);
        }}
      />

      {rolesQuery.error ? (
        <QueryErrorState error={rolesQuery.error} title="角色列表加载失败" />
      ) : null}

      {rolesQuery.isLoading ? (
        <div className="rounded-[14px] border border-border bg-surface px-4 py-8 text-[13px] leading-[20px] text-text-secondary">
          正在加载角色列表...
        </div>
      ) : (
        <RolesTable
          keyword={keyword}
          onEdit={setEditingRole}
          onOpenPermissions={setPermissionRole}
          rows={rows}
          totalRoles={totalRoles}
        />
      )}

      <Pagination
        aria-label="角色列表分页"
        className="flex flex-wrap items-center justify-between gap-3 rounded-[12px] bg-surface px-4 py-3"
        size="sm"
      >
        <Pagination.Summary className="text-[12px] leading-[18px] text-text-secondary">
          第 {safePage} / {pageCount} 页，共 {totalRoles} 项
        </Pagination.Summary>
        <Pagination.Content className="flex items-center gap-1">
          <Pagination.Item>
            <Pagination.Previous
              isDisabled={safePage <= 1}
              onPress={() => {
                setPage(safePage - 1);
              }}
            >
              上一页
            </Pagination.Previous>
          </Pagination.Item>
          <Pagination.Item>
            <Pagination.Link isActive>{safePage}</Pagination.Link>
          </Pagination.Item>
          <Pagination.Item>
            <Pagination.Next
              isDisabled={safePage >= pageCount}
              onPress={() => {
                setPage(safePage + 1);
              }}
            >
              下一页
            </Pagination.Next>
          </Pagination.Item>
        </Pagination.Content>
      </Pagination>

      <EditRoleDialog
        accessToken={accessToken}
        onOpenChange={(open) => {
          if (!open) {
            setEditingRole(null);
          }
        }}
        onSaved={async () => {
          await queryClient.invalidateQueries({ queryKey: ["admin-roles"] });
        }}
        open={Boolean(editingRole)}
        role={editingRole}
      />

      <RoleServicesDrawer
        accessToken={accessToken}
        onOpenChange={(open) => {
          if (!open) {
            setPermissionRole(null);
          }
        }}
        onSaved={async () => {
          await queryClient.invalidateQueries({ queryKey: ["admin-roles"] });
        }}
        open={Boolean(permissionRole)}
        role={permissionRole}
        services={services}
      />
    </div>
  );
}
