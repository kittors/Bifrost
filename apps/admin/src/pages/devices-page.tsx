import { EmptyState, Input, Table } from "@bifrost/ui";
import { useQuery } from "@tanstack/react-query";
import { useMemo, useState } from "react";

import { listAdminDevices, requireAccessToken } from "../entities/admin/api";
import { getCurrentAdminSession } from "../features/auth/store";
import { QueryErrorState } from "../shared/ui/query-error-state";
import { StatusBadge } from "../shared/ui/status-badge";

export function DevicesPage() {
  const session = getCurrentAdminSession();
  const accessToken = requireAccessToken(session);
  const [keyword, setKeyword] = useState("");
  const [status, setStatus] = useState("");

  const devicesQuery = useQuery({
    queryFn: () => listAdminDevices({ accessToken, keyword, status }),
    queryKey: ["admin-devices", accessToken, keyword, status],
  });

  const rows = devicesQuery.data?.items ?? [];
  const caption = useMemo(
    () => `当前共有 ${devicesQuery.data?.total ?? 0} 台设备`,
    [devicesQuery.data?.total],
  );

  return (
    <div className="space-y-4">
      <section>
        <h1 className="text-[16px] leading-[24px] font-semibold">设备管理</h1>
        <p className="text-[12px] leading-[18px] text-text-secondary">
          查看已绑定设备、系统版本和当前信任状态。
        </p>
      </section>

      <section className="rounded-[14px] border border-border bg-surface p-4">
        <div className="flex flex-wrap items-center gap-2">
          <Input
            className="max-w-[280px]"
            onChange={(event) => setKeyword(event.target.value)}
            placeholder="搜索设备名、用户名或指纹"
            value={keyword}
          />
          <select
            className="h-[32px] rounded-[6px] border border-border bg-surface px-3 text-[13px] leading-[20px]"
            onChange={(event) => setStatus(event.target.value)}
            value={status}
          >
            <option value="">全部状态</option>
            <option value="trusted">Trusted</option>
            <option value="disabled">Disabled</option>
          </select>
        </div>
      </section>

      {devicesQuery.error ? (
        <QueryErrorState error={devicesQuery.error} title="设备列表加载失败" />
      ) : null}

      <section className="overflow-hidden rounded-[14px] border border-border bg-surface">
        {rows.length === 0 && !devicesQuery.isLoading ? (
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
                </Table.Row>
              ))}
            </Table.Body>
          </Table.Root>
        )}
      </section>
    </div>
  );
}
