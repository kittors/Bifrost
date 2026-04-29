import {
  Alert,
  Button,
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
  Chip,
  EmptyState,
  Input,
} from "@heroui/react";
import { useQuery } from "@tanstack/react-query";
import { ExternalLink, Search, Server } from "lucide-react";
import { useState } from "react";

import { listClientServices } from "../../entities/services/api";
import { useDesktopSessionStore } from "../../entities/session/store";
import { normalizeUnknownError, resolveApiErrorMessage } from "../../shared/lib/http";

function buildLocalServiceURL(input: {
  baseURL: string;
  key: string;
  publicPath?: string;
  running: boolean;
}) {
  if (!input.running) {
    return "本地代理未启动";
  }

  return `${input.baseURL}${input.publicPath ?? `/s/${input.key}/`}`.replace(/([^:]\/)\/+/g, "$1");
}

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

  const serviceCount = servicesQuery.data?.length ?? 0;
  const normalizedError = servicesQuery.error ? normalizeUnknownError(servicesQuery.error) : null;

  return (
    <Card className="flex flex-1 flex-col overflow-hidden">
      <CardHeader className="pb-2">
        <div>
          <CardTitle>可访问服务</CardTitle>
          <CardDescription>通过本地入口在系统浏览器中打开。</CardDescription>
        </div>
        <Chip color={localProxyStatus.running ? "success" : "warning"} size="sm" variant="soft">
          {localProxyStatus.running ? `${serviceCount} 个服务` : "代理未启动"}
        </Chip>
      </CardHeader>

      <CardContent className="flex min-h-0 flex-1 flex-col gap-3">
        <label className="relative block">
          <Search className="pointer-events-none absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-text-muted" />
          <Input
            className="pl-9"
            value={keyword}
            onChange={(event) => setKeyword(event.target.value)}
            placeholder="搜索服务"
          />
        </label>

        {normalizedError ? (
          <Alert status="danger">
            <Alert.Content>
              <Alert.Title>服务列表不可用</Alert.Title>
              <Alert.Description>{normalizedError.userMessage}</Alert.Description>
              {normalizedError.requestId ? (
                <div className="mt-1 font-mono text-[12px]">
                  Request ID: {normalizedError.requestId}
                </div>
              ) : null}
            </Alert.Content>
          </Alert>
        ) : (
          <div className="flex min-h-0 flex-1 flex-col gap-2 overflow-y-auto pr-1">
            {(servicesQuery.data ?? []).map((service) => (
              <div
                className="group rounded-[10px] border border-border bg-surface px-3 py-2 transition-colors hover:bg-surface-2"
                key={service.id}
              >
                <div className="flex min-h-[44px] items-center justify-between gap-3">
                  <div className="flex min-w-0 items-center gap-3">
                    <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-[10px] bg-brand-soft text-brand">
                      <Server className="h-4 w-4" />
                    </div>
                    <div className="min-w-0">
                      <div className="truncate text-[13px] leading-[20px] font-medium">
                        {service.name}
                      </div>
                      <div className="truncate text-[12px] leading-[18px] text-text-secondary">
                        {service.group} · {service.accessSource}
                      </div>
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
                    <ExternalLink className="h-4 w-4" />
                    打开
                  </Button>
                </div>
                <div className="mt-2 flex items-center justify-between gap-2 text-[11px] leading-[16px] text-text-secondary">
                  <span className="truncate">
                    本地入口：
                    {buildLocalServiceURL({
                      baseURL: localProxyStatus.baseURL,
                      key: service.key,
                      publicPath: service.publicPath,
                      running: localProxyStatus.running,
                    })}
                  </span>
                  <Chip
                    color={service.status === "enabled" ? "success" : "default"}
                    size="sm"
                    variant="soft"
                  >
                    {service.status}
                  </Chip>
                </div>
              </div>
            ))}

            {servicesQuery.isLoading ? (
              <div className="rounded-[10px] border border-dashed border-border px-3 py-6 text-center text-[12px] leading-[18px] text-text-secondary">
                正在加载服务列表...
              </div>
            ) : null}

            {!servicesQuery.isLoading && serviceCount === 0 ? (
              <EmptyState className="p-3">
                <div className="text-[14px] leading-[22px] font-semibold">暂无可访问服务</div>
                <p className="text-[12px] leading-[18px] text-text-secondary">
                  当前账号没有可打开的服务。
                </p>
              </EmptyState>
            ) : null}
          </div>
        )}
      </CardContent>
    </Card>
  );
}
