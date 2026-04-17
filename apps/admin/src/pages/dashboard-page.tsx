import { EmptyState } from "@bifrost/ui";
import { useQuery } from "@tanstack/react-query";
import {
  listAdminAuditEvents,
  listAdminDevices,
  listAdminServices,
  listAdminUsers,
  requireAccessToken,
} from "../entities/admin/api";
import { getCurrentAdminSession } from "../features/auth/store";
import { formatAuditType } from "../shared/lib/format";
import { StatusBadge } from "../shared/ui/status-badge";

function SummaryCard({
  description,
  title,
  value,
}: {
  description: string;
  title: string;
  value: number;
}) {
  return (
    <div className="rounded-[14px] border border-border bg-surface p-4">
      <div className="text-[12px] leading-[18px] text-text-secondary">{title}</div>
      <div className="mt-2 text-[28px] leading-[32px] font-semibold tracking-[-0.02em]">
        {value}
      </div>
      <div className="mt-1 text-[12px] leading-[18px] text-text-muted">{description}</div>
    </div>
  );
}

export function DashboardPage() {
  const session = getCurrentAdminSession();
  const accessToken = requireAccessToken(session);

  const summaryQuery = useQuery({
    queryFn: async () => {
      const [users, services, devices, failures, events] = await Promise.all([
        listAdminUsers({ accessToken, pageSize: 1 }),
        listAdminServices({ accessToken, pageSize: 1 }),
        listAdminDevices({ accessToken, pageSize: 1 }),
        listAdminAuditEvents({ accessToken, pageSize: 1, result: "failure" }),
        listAdminAuditEvents({ accessToken, pageSize: 6 }),
      ]);

      return {
        devices: devices.total,
        events: events.items,
        failureEvents: failures.total,
        services: services.total,
        users: users.total,
      };
    },
    queryKey: ["admin-dashboard-summary", accessToken],
  });

  if (summaryQuery.isLoading) {
    return (
      <div className="text-[13px] leading-[20px] text-text-secondary">正在加载系统概览...</div>
    );
  }

  if (!summaryQuery.data) {
    return <EmptyState description="概览数据暂时不可用，请稍后刷新重试。" title="系统概览不可用" />;
  }

  return (
    <div className="space-y-5">
      <section className="grid gap-4 lg:grid-cols-4">
        <SummaryCard
          description="当前已配置的账号总数"
          title="Users"
          value={summaryQuery.data.users}
        />
        <SummaryCard
          description="当前已登记的设备数量"
          title="Devices"
          value={summaryQuery.data.devices}
        />
        <SummaryCard
          description="已收录到目录中的私有服务"
          title="Services"
          value={summaryQuery.data.services}
        />
        <SummaryCard
          description="当前后端已可查询到的失败审计条目"
          title="Failure Events"
          value={summaryQuery.data.failureEvents}
        />
      </section>

      <section className="grid gap-4 xl:grid-cols-[minmax(0,1fr)_320px]">
        <div className="rounded-[14px] border border-border bg-surface">
          <div className="flex items-center justify-between border-b border-border-soft px-4 py-3">
            <div>
              <h2 className="text-[16px] leading-[24px] font-semibold">最近审计</h2>
              <p className="text-[12px] leading-[18px] text-text-secondary">
                先展示现有列表接口可提供的最新事件。
              </p>
            </div>
          </div>
          <div className="divide-y divide-border-soft">
            {summaryQuery.data.events.length === 0 ? (
              <div className="px-4 py-8">
                <EmptyState description="当前没有可展示的审计事件。" title="暂无审计记录" />
              </div>
            ) : (
              summaryQuery.data.events.map((event) => (
                <div
                  className="grid gap-2 px-4 py-3 md:grid-cols-[minmax(0,1fr)_120px] md:items-center"
                  key={event.id}
                >
                  <div className="min-w-0">
                    <div className="text-[13px] leading-[20px] font-medium">{event.summary}</div>
                    <div className="truncate text-[12px] leading-[18px] text-text-secondary">
                      {formatAuditType(event.type)} · requestId {event.requestId || "-"}
                    </div>
                  </div>
                  <div className="flex justify-start md:justify-end">
                    <StatusBadge status={event.result} />
                  </div>
                </div>
              ))
            )}
          </div>
        </div>

        <div className="rounded-[14px] border border-border bg-surface p-4">
          <h2 className="text-[16px] leading-[24px] font-semibold">风险提示</h2>
          <div className="mt-3 space-y-3 text-[13px] leading-[20px] text-text-secondary">
            <div className="rounded-[10px] border border-border-soft bg-surface-2 p-3">
              当前仪表盘使用现有管理 API 聚合统计，后续会切到专用 dashboard summary 接口。
            </div>
            <div className="rounded-[10px] border border-border-soft bg-surface-2 p-3">
              失败事件已经可见，但按时间范围过滤和更细审计详情还会在后续页面中继续补齐。
            </div>
          </div>
        </div>
      </section>
    </div>
  );
}
