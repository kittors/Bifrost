import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useState } from "react";

import { listAdminDevices, requireAccessToken } from "../entities/admin/api";
import { DeviceDetailDrawer } from "../features/admin-devices/device-detail-drawer";
import { DevicesFilterBar } from "../features/admin-devices/devices-filter-bar";
import { DevicesTable } from "../features/admin-devices/devices-table";
import { getCurrentAdminSession } from "../features/auth/store";
import { QueryErrorState } from "../shared/ui/query-error-state";

export function DevicesPage() {
  const session = getCurrentAdminSession();
  const accessToken = requireAccessToken(session);
  const queryClient = useQueryClient();
  const [keyword, setKeyword] = useState("");
  const [status, setStatus] = useState("");
  const [selectedDeviceID, setSelectedDeviceID] = useState<string | null>(null);

  const devicesQuery = useQuery({
    queryFn: () => listAdminDevices({ accessToken, keyword, status }),
    queryKey: ["admin-devices", accessToken, keyword, status],
  });

  const rows = devicesQuery.data?.items ?? [];
  const totalDevices = devicesQuery.data?.total ?? 0;

  return (
    <div className="space-y-4">
      {/* 页面只保留查询与区块编排，设备表格和抽屉细节都在 feature 层维护。 */}
      <section>
        <h1 className="text-[16px] leading-[24px] font-semibold">设备管理</h1>
        <p className="text-[12px] leading-[18px] text-text-secondary">
          查看已绑定设备、系统版本和当前信任状态。
        </p>
      </section>

      <DevicesFilterBar
        keyword={keyword}
        onKeywordChange={setKeyword}
        onStatusChange={setStatus}
        status={status}
      />

      {devicesQuery.error ? (
        <QueryErrorState error={devicesQuery.error} title="设备列表加载失败" />
      ) : null}

      {devicesQuery.isLoading ? (
        <div className="rounded-[14px] border border-border bg-surface px-4 py-8 text-[13px] leading-[20px] text-text-secondary">
          正在加载设备列表...
        </div>
      ) : (
        <DevicesTable onOpenDetails={setSelectedDeviceID} rows={rows} totalDevices={totalDevices} />
      )}

      <DeviceDetailDrawer
        accessToken={accessToken}
        deviceID={selectedDeviceID}
        onOpenChange={(open) => {
          if (!open) {
            setSelectedDeviceID(null);
          }
        }}
        onUpdated={async () => {
          await queryClient.invalidateQueries({ queryKey: ["admin-devices"] });
        }}
        open={Boolean(selectedDeviceID)}
      />
    </div>
  );
}
