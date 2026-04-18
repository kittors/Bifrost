import { Button, Drawer, EmptyState, Table } from "@bifrost/ui";
import { useQuery } from "@tanstack/react-query";
import { useMemo, useState } from "react";

import { listAdminAuditEvents, requireAccessToken } from "../entities/admin/api";
import type { AdminAuditEvent } from "../entities/admin/types";
import { getCurrentAdminSession } from "../features/auth/store";
import { formatAuditType } from "../shared/lib/format";
import { PaginationBar } from "../shared/ui/pagination-bar";
import { QueryErrorState } from "../shared/ui/query-error-state";
import { StatusBadge } from "../shared/ui/status-badge";

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
  const caption = useMemo(
    () => `当前共有 ${auditQuery.data?.total ?? 0} 条审计记录`,
    [auditQuery.data?.total],
  );

  return (
    <div className="space-y-4">
      <section>
        <h1 className="text-[16px] leading-[24px] font-semibold">审计记录</h1>
        <p className="text-[12px] leading-[18px] text-text-secondary">
          查询登录、访问和配置变更事件，并通过 requestId 回溯链路。
        </p>
      </section>

      <section className="rounded-[14px] border border-border bg-surface p-4">
        <select
          className="h-[32px] rounded-[6px] border border-border bg-surface px-3 text-[13px] leading-[20px]"
          onChange={(event) => {
            setResult(event.target.value);
            setPage(1);
          }}
          value={result}
        >
          <option value="">全部结果</option>
          <option value="success">Success</option>
          <option value="failure">Failure</option>
        </select>
      </section>

      {auditQuery.error ? (
        <QueryErrorState error={auditQuery.error} title="审计列表加载失败" />
      ) : null}

      <section className="overflow-hidden rounded-[14px] border border-border bg-surface">
        {rows.length === 0 && !auditQuery.isLoading ? (
          <div className="px-4 py-8">
            <EmptyState description="当前没有审计记录。" title="暂无审计记录" />
          </div>
        ) : (
          <Table.Root>
            <Table.Caption>{caption}</Table.Caption>
            <Table.Header>
              <Table.Row>
                <Table.Head>事件</Table.Head>
                <Table.Head>摘要</Table.Head>
                <Table.Head>结果</Table.Head>
                <Table.Head className="text-right">操作</Table.Head>
              </Table.Row>
            </Table.Header>
            <Table.Body>
              {rows.map((event) => (
                <Table.Row key={event.id}>
                  <Table.Cell>{formatAuditType(event.type)}</Table.Cell>
                  <Table.Cell>{event.summary}</Table.Cell>
                  <Table.Cell>
                    <StatusBadge status={event.result} />
                  </Table.Cell>
                  <Table.Cell className="text-right">
                    <Button onClick={() => setSelectedEvent(event)} size="sm" variant="secondary">
                      查看详情
                    </Button>
                  </Table.Cell>
                </Table.Row>
              ))}
            </Table.Body>
          </Table.Root>
        )}
      </section>

      <PaginationBar
        onPageChange={setPage}
        page={page}
        pageSize={auditPageSize}
        total={auditQuery.data?.total ?? 0}
      />

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
