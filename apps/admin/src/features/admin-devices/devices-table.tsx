import { Button, EmptyState, Table } from "@bifrost/ui";

import type { AdminDevice } from "../../entities/admin/types";
import { StatusBadge } from "../../shared/ui/status-badge";

type DevicesTableProps = {
  onOpenDetails: (deviceID: string) => void;
  rows: AdminDevice[];
  totalDevices: number;
};

export function DevicesTable({ onOpenDetails, rows, totalDevices }: DevicesTableProps) {
  const caption = `当前共有 ${totalDevices} 台设备`;

  return (
    <section className="overflow-hidden rounded-[14px] border border-border bg-surface">
      {rows.length === 0 ? (
        <div className="px-4 py-8">
          <EmptyState description="当前没有设备记录。" title="暂无设备" />
        </div>
      ) : (
        <Table.Root>
          <Table.Caption>{caption}</Table.Caption>
          <Table.Header>
            <Table.Row>
              <Table.Head>设备</Table.Head>
              <Table.Head>用户</Table.Head>
              <Table.Head>指纹</Table.Head>
              <Table.Head>状态</Table.Head>
              <Table.Head className="text-right">操作</Table.Head>
            </Table.Row>
          </Table.Header>
          <Table.Body>
            {rows.map((device) => (
              <Table.Row key={device.id}>
                <Table.Cell>
                  <div className="space-y-0.5">
                    <div className="font-medium">{device.name}</div>
                    <div className="text-[12px] leading-[18px] text-text-secondary">
                      {device.os} · {device.clientVersion}
                    </div>
                  </div>
                </Table.Cell>
                <Table.Cell>{device.userUsername}</Table.Cell>
                <Table.Cell>{device.publicKeyFingerprint}</Table.Cell>
                <Table.Cell>
                  <StatusBadge status={device.status} />
                </Table.Cell>
                <Table.Cell className="text-right">
                  <Button
                    onClick={() => {
                      onOpenDetails(device.id);
                    }}
                    size="sm"
                    variant="ghost"
                  >
                    详情
                  </Button>
                </Table.Cell>
              </Table.Row>
            ))}
          </Table.Body>
        </Table.Root>
      )}
    </section>
  );
}
