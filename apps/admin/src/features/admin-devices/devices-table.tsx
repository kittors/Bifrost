import { Button, Table } from "@heroui/react";

import type { AdminDevice } from "../../entities/admin/types";
import { EmptyState } from "../../shared/ui/empty-state";
import { StatusBadge } from "../../shared/ui/status-badge";

type DevicesTableProps = {
  onOpenDetails: (deviceID: string) => void;
  rows: AdminDevice[];
  totalDevices: number;
};

export function DevicesTable({ onOpenDetails, rows, totalDevices }: DevicesTableProps) {
  const caption = `当前共有 ${totalDevices} 台设备`;

  return (
    <section className="overflow-hidden rounded-[12px] bg-surface">
      {rows.length === 0 ? (
        <div className="px-4 py-8">
          <EmptyState description="当前没有设备记录。" title="暂无设备" />
        </div>
      ) : (
        <Table className="w-full">
          <Table.ScrollContainer className="overflow-x-auto">
            <Table.Content aria-label="设备管理数据表" className="w-full text-left">
              <Table.Header className="[&_tr]:border-b [&_tr]:border-border">
                <Table.Column className="h-[36px] px-3 text-[12px] leading-[18px] font-medium text-text-secondary">
                  设备
                </Table.Column>
                <Table.Column className="h-[36px] px-3 text-[12px] leading-[18px] font-medium text-text-secondary">
                  用户
                </Table.Column>
                <Table.Column className="h-[36px] px-3 text-[12px] leading-[18px] font-medium text-text-secondary">
                  指纹
                </Table.Column>
                <Table.Column className="h-[36px] px-3 text-[12px] leading-[18px] font-medium text-text-secondary">
                  状态
                </Table.Column>
                <Table.Column className="h-[36px] px-3 text-right text-[12px] leading-[18px] font-medium text-text-secondary">
                  操作
                </Table.Column>
              </Table.Header>
              <Table.Body className="[&_tr:last-child]:border-0">
                {rows.map((device) => (
                  <Table.Row
                    className="h-[36px] border-b border-border transition-colors"
                    key={device.id}
                  >
                    <Table.Cell className="px-3 text-[13px] leading-[20px] text-text-primary">
                      <div className="space-y-0.5">
                        <div className="font-medium">{device.name}</div>
                        <div className="text-[12px] leading-[18px] text-text-secondary">
                          {device.os} · {device.clientVersion}
                        </div>
                      </div>
                    </Table.Cell>
                    <Table.Cell className="px-3 text-[13px] leading-[20px] text-text-primary">
                      {device.userUsername}
                    </Table.Cell>
                    <Table.Cell className="px-3 text-[13px] leading-[20px] text-text-primary">
                      {device.publicKeyFingerprint}
                    </Table.Cell>
                    <Table.Cell className="px-3 text-[13px] leading-[20px] text-text-primary">
                      <StatusBadge status={device.status} />
                    </Table.Cell>
                    <Table.Cell className="px-3 text-right text-[13px] leading-[20px] text-text-primary">
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
            </Table.Content>
          </Table.ScrollContainer>
          <Table.Footer className="px-3 pb-3 pt-2 text-[12px] leading-[18px] text-text-muted">
            {caption}
          </Table.Footer>
        </Table>
      )}
    </section>
  );
}
