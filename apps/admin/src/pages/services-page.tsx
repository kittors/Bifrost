import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";

import { listAdminServices, requireAccessToken } from "../entities/admin/api";
import { CreateServiceDialog } from "../features/admin-services/create-service-dialog";
import { ServicesFilterBar } from "../features/admin-services/services-filter-bar";
import { ServicesTable } from "../features/admin-services/services-table";
import { getCurrentAdminSession } from "../features/auth/store";
import { QueryErrorState } from "../shared/ui/query-error-state";

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

  const rows = servicesQuery.data?.items ?? [];
  const totalServices = servicesQuery.data?.total ?? 0;

  return (
    <div className="space-y-4">
      {/* 页面仅保留列表查询和区块编排，创建弹窗已拆到独立 feature。 */}
      <section className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-[16px] leading-[24px] font-semibold">服务目录</h1>
          <p className="text-[12px] leading-[18px] text-text-secondary">
            管理私有服务入口、上游地址和公开路径。
          </p>
        </div>

        <CreateServiceDialog
          accessToken={accessToken}
          onCreated={async () => {
            await queryClient.invalidateQueries({ queryKey: ["admin-services"] });
          }}
          onOpenChange={setCreateOpen}
          open={createOpen}
        />
      </section>

      <ServicesFilterBar
        keyword={keyword}
        onKeywordChange={setKeyword}
        onStatusChange={setStatus}
        status={status}
      />

      {servicesQuery.error ? (
        <QueryErrorState error={servicesQuery.error} title="服务列表加载失败" />
      ) : null}

      {servicesQuery.isLoading ? (
        <div className="rounded-[14px] border border-border bg-surface px-4 py-8 text-[13px] leading-[20px] text-text-secondary">
          正在加载服务列表...
        </div>
      ) : (
        <ServicesTable rows={rows} totalServices={totalServices} />
      )}
    </div>
  );
}
