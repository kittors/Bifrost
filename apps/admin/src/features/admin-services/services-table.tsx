import { Button, Table } from "@heroui/react";

import type { AdminService } from "../../entities/admin/types";
import { EmptyState } from "../../shared/ui/empty-state";
import { StatusBadge } from "../../shared/ui/status-badge";

type ServicesTableProps = {
  onEdit: (service: AdminService) => void;
  onToggleStatus: (service: AdminService) => Promise<void>;
  pendingServiceID: string | null;
  rows: AdminService[];
  totalServices: number;
};

export function ServicesTable({
  onEdit,
  onToggleStatus,
  pendingServiceID,
  rows,
  totalServices,
}: ServicesTableProps) {
  const caption = `当前共有 ${totalServices} 个服务`;

  return (
    <section className="overflow-hidden rounded-[12px] bg-surface">
      {rows.length === 0 ? (
        <div className="px-4 py-8">
          <EmptyState description="当前没有服务记录。" title="暂无服务" />
        </div>
      ) : (
        <Table className="w-full">
          <Table.ScrollContainer className="overflow-x-auto">
            <Table.Content aria-label="服务目录数据表" className="w-full text-left">
              <Table.Header className="[&_tr]:border-b [&_tr]:border-border">
                <Table.Column className="h-[36px] px-3 text-[12px] leading-[18px] font-medium text-text-secondary">
                  服务
                </Table.Column>
                <Table.Column className="h-[36px] px-3 text-[12px] leading-[18px] font-medium text-text-secondary">
                  上游地址
                </Table.Column>
                <Table.Column className="h-[36px] px-3 text-[12px] leading-[18px] font-medium text-text-secondary">
                  状态
                </Table.Column>
                <Table.Column className="h-[36px] px-3 text-right text-[12px] leading-[18px] font-medium text-text-secondary">
                  操作
                </Table.Column>
              </Table.Header>
              <Table.Body className="[&_tr:last-child]:border-0">
                {rows.map((service) => (
                  <Table.Row
                    className="h-[36px] border-b border-border transition-colors"
                    key={service.id}
                  >
                    <Table.Cell className="px-3 text-[13px] leading-[20px] text-text-primary">
                      <div className="space-y-0.5">
                        <div className="font-medium">{service.name}</div>
                        <div className="text-[12px] leading-[18px] text-text-secondary">
                          {service.key} · {service.publicPath}
                        </div>
                      </div>
                    </Table.Cell>
                    <Table.Cell className="px-3 text-[13px] leading-[20px] text-text-primary">
                      {service.upstreamUrl}
                    </Table.Cell>
                    <Table.Cell className="px-3 text-[13px] leading-[20px] text-text-primary">
                      <StatusBadge status={service.status} />
                    </Table.Cell>
                    <Table.Cell className="px-3 text-right text-[13px] leading-[20px] text-text-primary">
                      <div className="flex justify-end gap-2">
                        <Button
                          onClick={() => {
                            onEdit(service);
                          }}
                          size="sm"
                          variant="ghost"
                        >
                          编辑
                        </Button>
                        <Button
                          isDisabled={pendingServiceID === service.id}
                          onClick={async () => {
                            await onToggleStatus(service);
                          }}
                          size="sm"
                          variant={service.status === "enabled" ? "danger" : "secondary"}
                        >
                          {pendingServiceID === service.id
                            ? "处理中..."
                            : service.status === "enabled"
                              ? "禁用"
                              : "启用"}
                        </Button>
                      </div>
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
