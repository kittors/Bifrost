import { Button } from "@heroui/react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { toast } from "sonner";

import { getAdminDevice, setAdminDeviceStatus } from "../../entities/admin/api";
import { QueryErrorState } from "../../shared/ui/query-error-state";
import { StatusBadge } from "../../shared/ui/status-badge";
import { Drawer } from "../../shared/ui/drawer";

type DeviceDetailDrawerProps = {
  accessToken: string;
  deviceID: string | null;
  onOpenChange: (open: boolean) => void;
  onUpdated: () => Promise<void> | void;
  open: boolean;
};

// 设备详情抽屉只负责单台设备的展示与信任状态切换。
export function DeviceDetailDrawer({
  accessToken,
  deviceID,
  onOpenChange,
  onUpdated,
  open,
}: DeviceDetailDrawerProps) {
  const deviceQuery = useQuery({
    enabled: open && Boolean(deviceID),
    queryFn: () =>
      getAdminDevice({
        accessToken,
        deviceID: deviceID ?? "",
      }),
    queryKey: ["admin-device-detail", accessToken, deviceID],
  });

  const device = deviceQuery.data ?? null;
  const nextStatus = device?.status === "trusted" ? "disabled" : "trusted";
  const actionLabel = nextStatus === "disabled" ? "禁用设备" : "重新信任";

  const statusMutation = useMutation({
    mutationFn: () =>
      setAdminDeviceStatus({
        accessToken,
        deviceID: device?.id ?? "",
        status: nextStatus,
      }),
    onSuccess: async () => {
      toast.success("设备状态已更新");
      await deviceQuery.refetch();
      await onUpdated();
    },
  });

  return (
    <Drawer.Root onOpenChange={onOpenChange} open={open}>
      <Drawer.Content>
        <Drawer.Header>
          <Drawer.Title>{device?.name ?? "设备详情"}</Drawer.Title>
          <Drawer.Description>查看设备归属、指纹和当前信任状态。</Drawer.Description>
        </Drawer.Header>

        {deviceQuery.error ? (
          <div className="mt-4">
            <QueryErrorState error={deviceQuery.error} title="设备详情加载失败" />
          </div>
        ) : null}

        {device ? (
          <div className="mt-4 space-y-3">
            <DetailRow label="设备系统" value={device.os} />
            <DetailRow label="客户端版本" value={device.clientVersion} />
            <DetailRow label="所属用户" value={device.userUsername} />
            <DetailRow label="公钥指纹" value={device.publicKeyFingerprint} />
            <div className="flex items-center justify-between rounded-[12px] border border-border bg-surface-2 px-3 py-2">
              <span className="text-[12px] leading-[18px] text-text-secondary">状态</span>
              <StatusBadge status={device.status} />
            </div>
          </div>
        ) : (
          <div className="mt-4 rounded-[12px] border border-border bg-surface-2 px-3 py-4 text-[13px] leading-[20px] text-text-secondary">
            正在加载设备详情...
          </div>
        )}

        {statusMutation.error ? (
          <div className="mt-4">
            <QueryErrorState error={statusMutation.error} title="设备状态更新失败" />
          </div>
        ) : null}

        <Drawer.Footer>
          <Button
            onClick={() => {
              onOpenChange(false);
            }}
            size="sm"
            variant="ghost"
          >
            关闭
          </Button>
          <Button
            isDisabled={!device || statusMutation.isPending}
            onClick={async () => {
              await statusMutation.mutateAsync();
            }}
            size="sm"
            variant={nextStatus === "disabled" ? "danger" : "primary"}
          >
            {statusMutation.isPending ? "处理中..." : actionLabel}
          </Button>
        </Drawer.Footer>
      </Drawer.Content>
    </Drawer.Root>
  );
}

function DetailRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between gap-3 rounded-[12px] border border-border bg-surface-2 px-3 py-2">
      <span className="text-[12px] leading-[18px] text-text-secondary">{label}</span>
      <span className="truncate text-[13px] leading-[20px] font-medium">{value || "-"}</span>
    </div>
  );
}
