import { Button, Input } from "@bifrost/ui";
import { useQuery } from "@tanstack/react-query";
import { useState } from "react";

import { listClientServices } from "../../entities/services/api";
import { useDesktopSessionStore } from "../../entities/session/store";
import { resolveApiErrorMessage } from "../../shared/lib/http";

export function ServicesCard() {
  const { localProxyStatus, session, setErrorMessage } = useDesktopSessionStore();
  const [keyword, setKeyword] = useState("");

  const servicesQuery = useQuery({
    enabled: Boolean(session),
    queryFn: async () => {
      if (!session) {
        return [];
      }
      return listClientServices({
        accessToken: session.accessToken,
        baseURL: session.gatewayBaseURL,
        keyword,
      });
    },
    queryKey: ["desktop-services", session?.accessToken, keyword],
  });
  if (!session) {
    return null;
  }

  return (
    <section className="flex flex-1 flex-col gap-3 rounded-[14px] border border-border bg-surface p-3">
      <div className="flex items-center justify-between gap-2">
        <div>
          <div className="text-[14px] leading-[22px] font-semibold">可访问服务</div>
          <div className="text-[12px] leading-[18px] text-text-secondary">
            单列紧凑列表，不接管本机网络。
          </div>
        </div>
        <div className="rounded-full border border-border bg-surface-2 px-2 py-1 text-[11px] leading-[16px] text-text-secondary">
          {servicesQuery.data?.length ?? 0} 个服务
        </div>
      </div>

      <Input
        value={keyword}
        onChange={(event) => setKeyword(event.target.value)}
        placeholder="搜索服务"
      />

      <div className="flex flex-1 flex-col gap-2">
        {(servicesQuery.data ?? []).map((service) => (
          <div className="rounded-[10px] border border-border px-3 py-2" key={service.id}>
            <div className="flex min-h-[40px] items-center justify-between gap-3">
              <div className="min-w-0">
                <div className="truncate text-[13px] leading-[20px] font-medium">
                  {service.name}
                </div>
                <div className="truncate text-[12px] leading-[18px] text-text-secondary">
                  {service.group} · {service.accessSource}
                </div>
              </div>
              <Button
                onClick={async () => {
                  try {
                    await window.bifrostDesktop.localProxy.openService(
                      service.publicPath ?? `/s/${service.key}/`,
                    );
                  } catch (error) {
                    setErrorMessage(resolveApiErrorMessage(error, "打开服务失败"));
                  }
                }}
                size="sm"
              >
                打开
              </Button>
            </div>
            <div className="mt-2 truncate text-[11px] leading-[16px] text-text-secondary">
              本地入口：
              {localProxyStatus.running
                ? `${localProxyStatus.baseURL}${service.publicPath ?? `/s/${service.key}`}/`.replace(
                    /([^:]\/)\/+/g,
                    "$1",
                  )
                : "本地代理未启动"}
            </div>
          </div>
        ))}

        {servicesQuery.isLoading ? (
          <div className="rounded-[10px] border border-dashed border-border px-3 py-6 text-center text-[12px] leading-[18px] text-text-secondary">
            正在加载服务列表...
          </div>
        ) : null}

        {!servicesQuery.isLoading && (servicesQuery.data?.length ?? 0) === 0 ? (
          <div className="rounded-[10px] border border-dashed border-border px-3 py-6 text-center text-[12px] leading-[18px] text-text-secondary">
            当前没有可访问服务。
          </div>
        ) : null}
      </div>
    </section>
  );
}
