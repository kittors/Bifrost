import { Button, Table } from "@heroui/react";

import type { AdminRole } from "../../entities/admin/types";
import { EmptyState } from "../../shared/ui/empty-state";

type RolesTableProps = {
  keyword: string;
  onEdit: (role: AdminRole) => void;
  onOpenPermissions: (role: AdminRole) => void;
  rows: AdminRole[];
  totalRoles: number;
};

export function RolesTable({
  keyword,
  onEdit,
  onOpenPermissions,
  rows,
  totalRoles,
}: RolesTableProps) {
  const caption = `当前共有 ${totalRoles} 个角色`;

  return (
    <section className="overflow-hidden rounded-[12px] bg-surface">
      {rows.length === 0 ? (
        <div className="px-4 py-8">
          <EmptyState
            description="当前没有可展示的角色记录。"
            title={keyword ? "未匹配到角色" : "暂无角色"}
          />
        </div>
      ) : (
        <Table className="w-full">
          <Table.ScrollContainer className="overflow-x-auto">
            <Table.Content aria-label="角色管理数据表" className="w-full text-left">
              <Table.Header className="[&_tr]:border-b [&_tr]:border-border">
                <Table.Column className="h-[36px] px-3 text-[12px] leading-[18px] font-medium text-text-secondary">
                  角色
                </Table.Column>
                <Table.Column className="h-[36px] px-3 text-[12px] leading-[18px] font-medium text-text-secondary">
                  描述
                </Table.Column>
                <Table.Column className="h-[36px] px-3 text-right text-[12px] leading-[18px] font-medium text-text-secondary">
                  操作
                </Table.Column>
              </Table.Header>
              <Table.Body className="[&_tr:last-child]:border-0">
                {rows.map((role) => (
                  <Table.Row
                    className="h-[36px] border-b border-border transition-colors"
                    key={role.id}
                  >
                    <Table.Cell className="px-3 text-[13px] leading-[20px] text-text-primary">
                      <div className="space-y-0.5">
                        <div className="font-medium">{role.displayName}</div>
                        <div className="text-[12px] leading-[18px] text-text-secondary">
                          {role.name}
                        </div>
                      </div>
                    </Table.Cell>
                    <Table.Cell className="px-3 text-[13px] leading-[20px] text-text-primary">
                      {role.description}
                    </Table.Cell>
                    <Table.Cell className="px-3 text-right text-[13px] leading-[20px] text-text-primary">
                      <div className="flex justify-end gap-2">
                        <Button
                          onClick={() => {
                            onEdit(role);
                          }}
                          size="sm"
                          variant="ghost"
                        >
                          编辑
                        </Button>
                        <Button
                          onClick={() => {
                            onOpenPermissions(role);
                          }}
                          size="sm"
                          variant="secondary"
                        >
                          授权服务
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
