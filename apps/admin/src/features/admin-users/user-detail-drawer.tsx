import { Button, toast } from "@heroui/react";
import { useMutation, useQuery } from "@tanstack/react-query";
import { useState } from "react";

import { getAdminUser, setAdminUserStatus } from "../../entities/admin/api";
import { formatList } from "../../shared/lib/format";
import { QueryErrorState } from "../../shared/ui/query-error-state";
import { StatusBadge } from "../../shared/ui/status-badge";
import { Drawer } from "../../shared/ui/drawer";
import { ResetPasswordDialog } from "./reset-password-dialog";

type UserDetailDrawerProps = {
  accessToken: string;
  onOpenChange: (open: boolean) => void;
  onUpdated: () => Promise<void> | void;
  open: boolean;
  userID: string | null;
};

// 用户详情抽屉只处理详情读取和账号控制，列表页保持轻量编排。
export function UserDetailDrawer({
  accessToken,
  onOpenChange,
  onUpdated,
  open,
  userID,
}: UserDetailDrawerProps) {
  const [resetOpen, setResetOpen] = useState(false);

  const userQuery = useQuery({
    enabled: open && Boolean(userID),
    queryFn: () =>
      getAdminUser({
        accessToken,
        userID: userID ?? "",
      }),
    queryKey: ["admin-user-detail", accessToken, userID],
  });

  const user = userQuery.data ?? null;
  const nextStatus = user?.status === "enabled" ? "disabled" : "enabled";
  const statusLabel = nextStatus === "disabled" ? "禁用账号" : "启用账号";

  const statusMutation = useMutation({
    mutationFn: () =>
      setAdminUserStatus({
        accessToken,
        status: nextStatus,
        userID: user?.id ?? "",
      }),
    onSuccess: async () => {
      toast.success("用户状态已更新");
      await userQuery.refetch();
      await onUpdated();
    },
  });

  return (
    <Drawer.Root onOpenChange={onOpenChange} open={open}>
      <Drawer.Content>
        <Drawer.Header>
          <Drawer.Title>{user?.displayName ?? "用户详情"}</Drawer.Title>
          <Drawer.Description>查看账号状态、角色和邮箱，并执行必要的账号控制。</Drawer.Description>
        </Drawer.Header>

        {userQuery.error ? (
          <div className="mt-4">
            <QueryErrorState error={userQuery.error} title="用户详情加载失败" />
          </div>
        ) : null}

        {userQuery.isLoading ? (
          <div className="mt-4 rounded-[12px] border border-border bg-surface-2 px-3 py-4 text-[13px] leading-[20px] text-text-secondary">
            正在加载用户详情...
          </div>
        ) : null}

        {user ? (
          <div className="mt-4 space-y-3">
            <DetailRow label="用户名" value={user.username} />
            <DetailRow label="邮箱" value={user.email || "-"} />
            <DetailRow label="角色" value={formatList(user.roles)} />
            <div className="flex items-center justify-between rounded-[12px] border border-border bg-surface-2 px-3 py-2">
              <span className="text-[12px] leading-[18px] text-text-secondary">状态</span>
              <StatusBadge status={user.status} />
            </div>
          </div>
        ) : null}

        {statusMutation.error ? (
          <div className="mt-4">
            <QueryErrorState error={statusMutation.error} title="用户状态更新失败" />
          </div>
        ) : null}

        <Drawer.Footer className="justify-between">
          <Button
            isDisabled={!user}
            onClick={() => {
              setResetOpen(true);
            }}
            size="sm"
            variant="secondary"
          >
            重置密码
          </Button>
          <div className="flex items-center gap-2">
            <Button
              onClick={() => {
                onOpenChange(false);
              }}
              size="sm"
              variant="ghost"
            >
              关闭
            </Button>
            <Button
              isDisabled={!user || statusMutation.isPending}
              onClick={async () => {
                await statusMutation.mutateAsync();
              }}
              size="sm"
              variant={nextStatus === "disabled" ? "danger" : "primary"}
            >
              {statusMutation.isPending ? "处理中..." : statusLabel}
            </Button>
          </div>
        </Drawer.Footer>

        <ResetPasswordDialog
          accessToken={accessToken}
          onOpenChange={setResetOpen}
          onReset={async () => {
            await onUpdated();
          }}
          open={resetOpen}
          user={user}
        />
      </Drawer.Content>
    </Drawer.Root>
  );
}

function DetailRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between gap-3 rounded-[12px] border border-border bg-surface-2 px-3 py-2">
      <span className="text-[12px] leading-[18px] text-text-secondary">{label}</span>
      <span className="truncate text-[13px] leading-[20px] font-medium">{value}</span>
    </div>
  );
}
