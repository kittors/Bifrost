import { Button, Table } from "@heroui/react";

import type { AdminUser } from "../../entities/admin/types";
import { formatList, formatStatusLabel } from "../../shared/lib/format";
import { EmptyState } from "../../shared/ui/empty-state";
import { StatusBadge } from "../../shared/ui/status-badge";

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
    <section className="overflow-hidden rounded-[12px] bg-surface">
      {rows.length === 0 ? (
        <div className="px-4 py-8">
          <EmptyState
            description="当前条件下没有可展示的用户记录。"
            title={keyword || status ? "未匹配到用户" : "暂无用户"}
          />
        </div>
      ) : (
        <Table className="w-full">
          <Table.ScrollContainer className="overflow-x-auto">
            <Table.Content aria-label="用户管理数据表" className="w-full text-left">
              <Table.Header className="[&_tr]:border-b [&_tr]:border-border">
                <Table.Column className="h-[36px] px-3 text-[12px] leading-[18px] font-medium text-text-secondary">
                  用户
                </Table.Column>
                <Table.Column className="h-[36px] px-3 text-[12px] leading-[18px] font-medium text-text-secondary">
                  邮箱
                </Table.Column>
                <Table.Column className="h-[36px] px-3 text-[12px] leading-[18px] font-medium text-text-secondary">
                  角色
                </Table.Column>
                <Table.Column className="h-[36px] px-3 text-[12px] leading-[18px] font-medium text-text-secondary">
                  状态
                </Table.Column>
                <Table.Column className="h-[36px] px-3 text-right text-[12px] leading-[18px] font-medium text-text-secondary">
                  操作
                </Table.Column>
              </Table.Header>
              <Table.Body className="[&_tr:last-child]:border-0">
                {rows.map((user) => (
                  <Table.Row
                    className="h-[36px] border-b border-border transition-colors"
                    key={user.id}
                  >
                    <Table.Cell className="px-3 text-[13px] leading-[20px] text-text-primary">
                      <div className="space-y-0.5">
                        <div className="font-medium">{user.displayName}</div>
                        <div className="text-[12px] leading-[18px] text-text-secondary">
                          {user.username}
                        </div>
                      </div>
                    </Table.Cell>
                    <Table.Cell className="px-3 text-[13px] leading-[20px] text-text-primary">
                      {user.email}
                    </Table.Cell>
                    <Table.Cell className="px-3 text-[13px] leading-[20px] text-text-primary">
                      {formatList(user.roles)}
                    </Table.Cell>
                    <Table.Cell className="px-3 text-[13px] leading-[20px] text-text-primary">
                      <StatusBadge status={formatStatusLabel(user.status).toLowerCase()} />
                    </Table.Cell>
                    <Table.Cell className="px-3 text-right text-[13px] leading-[20px] text-text-primary">
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
            </Table.Content>
          </Table.ScrollContainer>
          <Table.Footer className="px-3 pb-3 pt-2 text-[12px] leading-[18px] text-text-muted">
            {resultCaption}
          </Table.Footer>
        </Table>
      )}
    </section>
  );
}
