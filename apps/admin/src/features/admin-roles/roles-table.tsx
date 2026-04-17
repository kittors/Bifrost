import { Button, EmptyState, Table } from "@bifrost/ui";

import type { AdminRole } from "../../entities/admin/types";

type RolesTableProps = {
  keyword: string;
  onOpenPermissions: (role: AdminRole) => void;
  rows: AdminRole[];
  totalRoles: number;
};

export function RolesTable({ keyword, onOpenPermissions, rows, totalRoles }: RolesTableProps) {
  const caption = `当前共有 ${totalRoles} 个角色`;

  return (
    <section className="overflow-hidden rounded-[14px] border border-border bg-surface">
      {rows.length === 0 ? (
        <div className="px-4 py-8">
          <EmptyState
            description="当前没有可展示的角色记录。"
            title={keyword ? "未匹配到角色" : "暂无角色"}
          />
        </div>
      ) : (
        <Table.Root>
          <Table.Caption>{caption}</Table.Caption>
          <Table.Header>
            <Table.Row>
              <Table.Head>角色</Table.Head>
              <Table.Head>描述</Table.Head>
              <Table.Head className="text-right">操作</Table.Head>
            </Table.Row>
          </Table.Header>
          <Table.Body>
            {rows.map((role) => (
              <Table.Row key={role.id}>
                <Table.Cell>
                  <div className="space-y-0.5">
                    <div className="font-medium">{role.displayName}</div>
                    <div className="text-[12px] leading-[18px] text-text-secondary">
                      {role.name}
                    </div>
                  </div>
                </Table.Cell>
                <Table.Cell>{role.description}</Table.Cell>
                <Table.Cell className="text-right">
                  <Button
                    onClick={() => {
                      onOpenPermissions(role);
                    }}
                    size="sm"
                    variant="secondary"
                  >
                    授权服务
                  </Button>
                </Table.Cell>
              </Table.Row>
            ))}
          </Table.Body>
        </Table.Root>
      )}
    </section>
  );
}
