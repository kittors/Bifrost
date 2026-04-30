import { Button, ListBox, Pagination, Select, Table } from "@heroui/react";
import { useQuery } from "@tanstack/react-query";
import { useMemo, useState } from "react";

import { listAdminAuditEvents, requireAccessToken } from "../entities/admin/api";
import type { AdminAuditEvent } from "../entities/admin/types";
import { getCurrentAdminSession } from "../features/auth/store";
import { formatAuditType } from "../shared/lib/format";
import { QueryErrorState } from "../shared/ui/query-error-state";
import { StatusBadge } from "../shared/ui/status-badge";
import { Drawer } from "../shared/ui/drawer";
import { EmptyState } from "../shared/ui/empty-state";

const auditPageSize = 20;

export function AuditEventsPage() {
  const session = getCurrentAdminSession();
  const accessToken = requireAccessToken(session);
  const [page, setPage] = useState(1);
  const [result, setResult] = useState("");
  const [selectedEvent, setSelectedEvent] = useState<AdminAuditEvent | null>(null);

  const auditQuery = useQuery({
    queryFn: () => listAdminAuditEvents({ accessToken, page, pageSize: auditPageSize, result }),
    queryKey: ["admin-audit-events", accessToken, page, result],
  });

  const rows = auditQuery.data?.items ?? [];
  const totalAuditEvents = auditQuery.data?.total ?? 0;
  const pageCount = Math.max(1, Math.ceil(totalAuditEvents / auditPageSize));
  const safePage = Math.min(Math.max(page, 1), pageCount);
  const selectedResultKey = result || "all";
  const caption = useMemo(() => `当前共有 ${totalAuditEvents} 条审计记录`, [totalAuditEvents]);

  return (
    <div className="space-y-4">
      <section>
        <h1 className="text-[16px] leading-[24px] font-semibold">审计记录</h1>
        <p className="text-[12px] leading-[18px] text-text-secondary">
          查询登录、访问和配置变更事件，并通过 requestId 回溯链路。
        </p>
      </section>

      <section className="rounded-[14px] border border-border bg-surface p-4">
        <Select
          aria-label="审计结果筛选"
          className="w-[150px]"
          onSelectionChange={(key) => {
            const value = String(key);
            setResult(value === "all" ? "" : value);
            setPage(1);
          }}
          selectedKey={selectedResultKey}
        >
          <Select.Trigger>
            <Select.Value />
            <Select.Indicator />
          </Select.Trigger>
          <Select.Popover>
            <ListBox>
              <ListBox.Item id="all">全部结果</ListBox.Item>
              <ListBox.Item id="success">Success</ListBox.Item>
              <ListBox.Item id="failure">Failure</ListBox.Item>
            </ListBox>
          </Select.Popover>
        </Select>
      </section>

      {auditQuery.error ? (
        <QueryErrorState error={auditQuery.error} title="审计列表加载失败" />
      ) : null}

      <section className="overflow-hidden rounded-[12px] bg-surface">
        {rows.length === 0 && !auditQuery.isLoading ? (
          <div className="px-4 py-8">
            <EmptyState description="当前没有审计记录。" title="暂无审计记录" />
          </div>
        ) : (
          <Table className="w-full">
            <Table.ScrollContainer className="overflow-x-auto">
              <Table.Content aria-label="审计记录数据表" className="w-full text-left">
                <Table.Header className="[&_tr]:border-b [&_tr]:border-border">
                  <Table.Column className="h-[36px] px-3 text-[12px] leading-[18px] font-medium text-text-secondary">
                    事件
                  </Table.Column>
                  <Table.Column className="h-[36px] px-3 text-[12px] leading-[18px] font-medium text-text-secondary">
                    摘要
                  </Table.Column>
                  <Table.Column className="h-[36px] px-3 text-[12px] leading-[18px] font-medium text-text-secondary">
                    结果
                  </Table.Column>
                  <Table.Column className="h-[36px] px-3 text-right text-[12px] leading-[18px] font-medium text-text-secondary">
                    操作
                  </Table.Column>
                </Table.Header>
                <Table.Body className="[&_tr:last-child]:border-0">
                  {rows.map((event) => (
                    <Table.Row
                      className="h-[36px] border-b border-border transition-colors"
                      key={event.id}
                    >
                      <Table.Cell className="px-3 text-[13px] leading-[20px] text-text-primary">
                        {formatAuditType(event.type)}
                      </Table.Cell>
                      <Table.Cell className="px-3 text-[13px] leading-[20px] text-text-primary">
                        {event.summary}
                      </Table.Cell>
                      <Table.Cell className="px-3 text-[13px] leading-[20px] text-text-primary">
                        <StatusBadge status={event.result} />
                      </Table.Cell>
                      <Table.Cell className="px-3 text-right text-[13px] leading-[20px] text-text-primary">
                        <Button
                          onClick={() => setSelectedEvent(event)}
                          size="sm"
                          variant="secondary"
                        >
                          查看详情
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

      <Pagination
        aria-label="审计记录分页"
        className="flex flex-wrap items-center justify-between gap-3 rounded-[12px] bg-surface px-4 py-3"
        size="sm"
      >
        <Pagination.Summary className="text-[12px] leading-[18px] text-text-secondary">
          第 {safePage} / {pageCount} 页，共 {totalAuditEvents} 项
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

      <Drawer.Root
        onOpenChange={(open) => !open && setSelectedEvent(null)}
        open={Boolean(selectedEvent)}
      >
        <Drawer.Content className="w-[min(720px,calc(100vw-24px))]">
          <Drawer.Header>
            <Drawer.Title>审计详情</Drawer.Title>
            <Drawer.Description>查看当前事件的关键追踪字段和 requestId。</Drawer.Description>
          </Drawer.Header>

          {selectedEvent ? (
            <div className="mt-4 grid gap-3 text-[13px] leading-[20px]">
              {[
                ["requestId", selectedEvent.requestId],
                ["type", selectedEvent.type],
                ["actorUserId", selectedEvent.actorUserId],
                ["targetType", selectedEvent.targetType],
                ["targetId", selectedEvent.targetId],
                ["serviceId", selectedEvent.serviceId],
                ["result", selectedEvent.result],
                ["summary", selectedEvent.summary],
              ].map(([label, value]) => (
                <div
                  className="rounded-[10px] border border-border bg-surface-2 px-3 py-2"
                  key={label}
                >
                  <div className="text-[12px] leading-[18px] text-text-secondary">{label}</div>
                  <div className="mt-1 font-medium">{value || "-"}</div>
                </div>
              ))}
            </div>
          ) : null}

          <Drawer.Footer>
            <Button onClick={() => setSelectedEvent(null)} variant="secondary">
              关闭
            </Button>
          </Drawer.Footer>
        </Drawer.Content>
      </Drawer.Root>
    </div>
  );
}
