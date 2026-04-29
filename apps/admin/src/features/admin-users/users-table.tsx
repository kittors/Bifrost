import { Button } from "@heroui/react";

import type { AdminUser } from "../../entities/admin/types";
import { formatList, formatStatusLabel } from "../../shared/lib/format";
import { EmptyState } from "../../shared/ui/empty-state";
import { StatusBadge } from "../../shared/ui/status-badge";
import { Table } from "../../shared/ui/table";

type UsersTableProps = {
  keyword: string;
  onOpenDetails: (userID: string) => void;
  onOpenOverrides: (user: AdminUser) => void;
  rows: AdminUser[];
  status: string;
  totalUsers: number;
};

export function UsersTable({
  keyword,
  onOpenDetails,
  onOpenOverrides,
  rows,
  status,
  totalUsers,
}: UsersTableProps) {
  const resultCaption =
    keyword || status ? `当前筛选命中 ${totalUsers} 个用户` : `当前共有 ${totalUsers} 个用户`;

  return (
    <section className="overflow-hidden rounded-[14px] border border-border bg-surface">
      {rows.length === 0 ? (
        <div className="px-4 py-8">
          <EmptyState
            description="当前条件下没有可展示的用户记录。"
            title={keyword || status ? "未匹配到用户" : "暂无用户"}
          />
        </div>
      ) : (
        <Table.Root>
          <Table.Caption>{resultCaption}</Table.Caption>
          <Table.Header>
            <Table.Head>用户</Table.Head>
            <Table.Head>邮箱</Table.Head>
            <Table.Head>角色</Table.Head>
            <Table.Head>状态</Table.Head>
            <Table.Head className="text-right">操作</Table.Head>
          </Table.Header>
          <Table.Body>
            {rows.map((user) => (
              <Table.Row key={user.id}>
                <Table.Cell>
                  <div className="space-y-0.5">
                    <div className="font-medium">{user.displayName}</div>
                    <div className="text-[12px] leading-[18px] text-text-secondary">
                      {user.username}
                    </div>
                  </div>
                </Table.Cell>
                <Table.Cell>{user.email}</Table.Cell>
                <Table.Cell>{formatList(user.roles)}</Table.Cell>
                <Table.Cell>
                  <StatusBadge status={formatStatusLabel(user.status).toLowerCase()} />
                </Table.Cell>
                <Table.Cell className="text-right">
                  <div className="flex justify-end gap-2">
                    <Button
                      onClick={() => {
                        onOpenOverrides(user);
                      }}
                      size="sm"
                      variant="secondary"
                    >
                      覆盖
                    </Button>
                    <Button
                      onClick={() => {
                        onOpenDetails(user.id);
                      }}
                      size="sm"
                      variant="ghost"
                    >
                      详情
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
