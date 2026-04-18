import { Button, Drawer, ErrorState } from "@bifrost/ui";
import { useMutation, useQuery } from "@tanstack/react-query";
import { useEffect, useMemo, useState } from "react";
import { toast } from "sonner";

import { listUserServiceOverrides, replaceUserServiceOverrides } from "../../entities/admin/api";
import type { AdminService, AdminUser } from "../../entities/admin/types";
import { normalizeUnknownError } from "../../shared/lib/http";

type UserServiceOverridesDrawerProps = {
  accessToken: string;
  onOpenChange: (open: boolean) => void;
  onSaved: () => Promise<void> | void;
  open: boolean;
  services: AdminService[];
  user: AdminUser | null;
};

// 用户级覆盖抽屉负责 allow/deny 双集合编辑，不让页面层直接维护复杂选择状态。
export function UserServiceOverridesDrawer({
  accessToken,
  onOpenChange,
  onSaved,
  open,
  services,
  user,
}: UserServiceOverridesDrawerProps) {
  const [allowServiceIDs, setAllowServiceIDs] = useState<string[]>([]);
  const [denyServiceIDs, setDenyServiceIDs] = useState<string[]>([]);

  const overridesQuery = useQuery({
    enabled: open && Boolean(user),
    queryFn: () =>
      listUserServiceOverrides({
        accessToken,
        userID: user?.id ?? "",
      }),
    queryKey: ["admin-user-service-overrides", accessToken, user?.id],
  });

  useEffect(() => {
    if (overridesQuery.data) {
      setAllowServiceIDs(
        overridesQuery.data.filter((item) => item.effect === "allow").map((item) => item.serviceId),
      );
      setDenyServiceIDs(
        overridesQuery.data.filter((item) => item.effect === "deny").map((item) => item.serviceId),
      );
    }
  }, [overridesQuery.data]);

  useEffect(() => {
    if (!open) {
      setAllowServiceIDs([]);
      setDenyServiceIDs([]);
    }
  }, [open]);

  const replaceOverridesMutation = useMutation({
    mutationFn: () =>
      replaceUserServiceOverrides({
        accessToken,
        allowServiceIDs,
        denyServiceIDs,
        userID: user?.id ?? "",
      }),
    onSuccess: async () => {
      toast.success("用户级服务覆盖已更新");
      onOpenChange(false);
      await onSaved();
    },
  });

  const queryError = overridesQuery.error ? normalizeUnknownError(overridesQuery.error) : null;
  const mutationError = replaceOverridesMutation.error
    ? normalizeUnknownError(replaceOverridesMutation.error)
    : null;

  const serviceRows = useMemo(
    () =>
      services.map((service) => ({
        ...service,
        mode: denyServiceIDs.includes(service.id)
          ? "deny"
          : allowServiceIDs.includes(service.id)
            ? "allow"
            : "inherit",
      })),
    [allowServiceIDs, denyServiceIDs, services],
  );

  return (
    <Drawer.Root onOpenChange={onOpenChange} open={open}>
      <Drawer.Content className="w-[min(680px,calc(100vw-24px))]">
        <Drawer.Header>
          <Drawer.Title>{user?.displayName ?? "用户级服务覆盖"}</Drawer.Title>
          <Drawer.Description>
            为单个用户配置 allow / deny 例外策略，deny 优先级最高。
          </Drawer.Description>
        </Drawer.Header>

        {queryError ? (
          <div className="mt-4">
            <ErrorState
              description={queryError.userMessage}
              requestId={queryError.requestId || undefined}
              title="覆盖配置加载失败"
            />
          </div>
        ) : null}

        {mutationError ? (
          <div className="mt-4">
            <ErrorState
              description={mutationError.userMessage}
              requestId={mutationError.requestId || undefined}
              title="覆盖配置提交失败"
            />
          </div>
        ) : null}

        <div className="mt-4 space-y-2 overflow-y-auto">
          {serviceRows.map((service) => (
            <div
              className="grid grid-cols-[minmax(0,1fr)_auto] items-center gap-3 rounded-[12px] border border-border bg-surface-2 px-3 py-2"
              key={service.id}
            >
              <div className="min-w-0">
                <div className="text-[13px] leading-[20px] font-medium">{service.name}</div>
                <div className="text-[12px] leading-[18px] text-text-secondary">
                  {service.key} · {service.group} · {service.mode}
                </div>
              </div>

              <div className="flex gap-2">
                <Button
                  onClick={() => {
                    setAllowServiceIDs((current) =>
                      current.includes(service.id)
                        ? current.filter((item) => item !== service.id)
                        : [...current, service.id],
                    );
                    setDenyServiceIDs((current) => current.filter((item) => item !== service.id));
                  }}
                  size="sm"
                  variant={service.mode === "allow" ? "primary" : "secondary"}
                >
                  Allow
                </Button>
                <Button
                  onClick={() => {
                    setDenyServiceIDs((current) =>
                      current.includes(service.id)
                        ? current.filter((item) => item !== service.id)
                        : [...current, service.id],
                    );
                    setAllowServiceIDs((current) => current.filter((item) => item !== service.id));
                  }}
                  size="sm"
                  variant={service.mode === "deny" ? "danger" : "secondary"}
                >
                  Deny
                </Button>
                <Button
                  onClick={() => {
                    setAllowServiceIDs((current) => current.filter((item) => item !== service.id));
                    setDenyServiceIDs((current) => current.filter((item) => item !== service.id));
                  }}
                  size="sm"
                  variant="ghost"
                >
                  继承
                </Button>
              </div>
            </div>
          ))}
        </div>

        <Drawer.Footer>
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
            disabled={replaceOverridesMutation.isPending || !user}
            onClick={async () => {
              await replaceOverridesMutation.mutateAsync();
            }}
            size="sm"
          >
            {replaceOverridesMutation.isPending ? "提交中..." : "保存覆盖"}
          </Button>
        </Drawer.Footer>
      </Drawer.Content>
    </Drawer.Root>
  );
}
