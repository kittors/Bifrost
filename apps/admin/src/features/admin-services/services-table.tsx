import { Button, EmptyState, Table } from "@bifrost/ui";

import type { AdminService } from "../../entities/admin/types";
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
    <section className="overflow-hidden rounded-[14px] border border-border bg-surface">
      {rows.length === 0 ? (
        <div className="px-4 py-8">
          <EmptyState description="当前没有服务记录。" title="暂无服务" />
        </div>
      ) : (
        <Table.Root>
          <Table.Caption>{caption}</Table.Caption>
          <Table.Header>
            <Table.Row>
              <Table.Head>服务</Table.Head>
              <Table.Head>上游地址</Table.Head>
              <Table.Head>状态</Table.Head>
              <Table.Head className="text-right">操作</Table.Head>
            </Table.Row>
          </Table.Header>
          <Table.Body>
            {rows.map((service) => (
              <Table.Row key={service.id}>
                <Table.Cell>
                  <div className="space-y-0.5">
                    <div className="font-medium">{service.name}</div>
                    <div className="text-[12px] leading-[18px] text-text-secondary">
                      {service.key} · {service.publicPath}
                    </div>
                  </div>
                </Table.Cell>
                <Table.Cell>{service.upstreamUrl}</Table.Cell>
                <Table.Cell>
                  <StatusBadge status={service.status} />
                </Table.Cell>
                <Table.Cell className="text-right">
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
                      disabled={pendingServiceID === service.id}
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
        </Table.Root>
      )}
    </section>
  );
}
