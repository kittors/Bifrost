import { Button } from "@heroui/react";
import { useMutation } from "@tanstack/react-query";
import { useEffect, useState } from "react";
import { toast } from "sonner";

import { replaceRoleServices } from "../../entities/admin/api";
import type { AdminRole, AdminService } from "../../entities/admin/types";
import { QueryErrorState } from "../../shared/ui/query-error-state";
import { Drawer } from "../../shared/ui/drawer";

type RoleServicesDrawerProps = {
  accessToken: string;
  onOpenChange: (open: boolean) => void;
  onSaved: () => Promise<void> | void;
  open: boolean;
  role: AdminRole | null;
  services: AdminService[];
};

// 授权抽屉只关心“角色选择哪些服务”，不耦合角色列表查询本身。
export function RoleServicesDrawer({
  accessToken,
  onOpenChange,
  onSaved,
  open,
  role,
  services,
}: RoleServicesDrawerProps) {
  const [selectedServiceIDs, setSelectedServiceIDs] = useState<string[]>([]);

  useEffect(() => {
    if (!open) {
      setSelectedServiceIDs([]);
    }
  }, [open]);

  const replaceRoleServicesMutation = useMutation({
    mutationFn: () =>
      replaceRoleServices({
        accessToken,
        roleID: role?.id ?? "",
        serviceIDs: selectedServiceIDs,
      }),
    onSuccess: async () => {
      toast.success("角色服务授权已提交");
      onOpenChange(false);
      await onSaved();
    },
  });

  return (
    <Drawer.Root onOpenChange={onOpenChange} open={open}>
      <Drawer.Content className="w-[min(640px,calc(100vw-24px))]">
        <Drawer.Header>
          <Drawer.Title>{role?.displayName ?? "角色服务授权"}</Drawer.Title>
          <Drawer.Description>
            当前后端仍未提供角色服务摘要查询接口，这里会以你勾选的服务集合执行覆盖写入。
          </Drawer.Description>
        </Drawer.Header>

        {replaceRoleServicesMutation.error ? (
          <div className="mt-4">
            <QueryErrorState error={replaceRoleServicesMutation.error} title="角色授权提交失败" />
          </div>
        ) : null}

        <div className="mt-4 grid gap-2 overflow-y-auto">
          {services.map((service) => {
            const checked = selectedServiceIDs.includes(service.id);

            return (
              <label
                className="flex items-start gap-3 rounded-[10px] border border-border bg-surface-2 px-3 py-2"
                key={service.id}
              >
                <input
                  checked={checked}
                  className="mt-1 h-4 w-4 accent-[var(--bifrost-brand)]"
                  onChange={(event) => {
                    setSelectedServiceIDs((current) =>
                      event.target.checked
                        ? [...current, service.id]
                        : current.filter((item) => item !== service.id),
                    );
                  }}
                  type="checkbox"
                />
                <div className="min-w-0">
                  <div className="text-[13px] leading-[20px] font-medium">{service.name}</div>
                  <div className="text-[12px] leading-[18px] text-text-secondary">
                    {service.key} · {service.group}
                  </div>
                </div>
              </label>
            );
          })}
        </div>

        <Drawer.Footer>
          <Button
            onClick={() => {
              onOpenChange(false);
            }}
            variant="secondary"
          >
            取消
          </Button>
          <Button
            isDisabled={replaceRoleServicesMutation.isPending || !role}
            onClick={async () => {
              await replaceRoleServicesMutation.mutateAsync();
            }}
          >
            {replaceRoleServicesMutation.isPending ? "提交中..." : "保存授权"}
          </Button>
        </Drawer.Footer>
      </Drawer.Content>
    </Drawer.Root>
  );
}
