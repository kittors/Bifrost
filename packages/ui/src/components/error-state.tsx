import type * as React from "react";
import { cn } from "../lib/utils";
import { Badge } from "./badge";

export interface ErrorStateProps extends React.HTMLAttributes<HTMLDivElement> {
  title: string;
  description: string;
  requestId?: string;
  action?: React.ReactNode;
}

export function ErrorState({
  action,
  className,
  description,
  requestId,
  title,
  ...props
}: ErrorStateProps) {
  return (
    <div
      className={cn(
        "flex flex-col items-start gap-2 rounded-[14px] border border-danger/20 bg-surface p-4 text-left",
        className,
      )}
      data-slot="error-state"
      {...props}
    >
      <Badge variant="danger">Error</Badge>
      <div className="text-[16px] leading-[24px] font-semibold text-text-primary">{title}</div>
      <p className="text-[13px] leading-[20px] text-text-secondary">{description}</p>
      {requestId ? (
        <div className="rounded-[6px] bg-surface-2 px-2 py-1 font-mono text-[12px] leading-[18px] text-text-secondary">
          Request ID: {requestId}
        </div>
      ) : null}
      {action ? <div className="pt-1">{action}</div> : null}
    </div>
  );
}
