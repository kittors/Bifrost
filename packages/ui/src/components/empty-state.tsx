import type * as React from "react";

import { cn } from "../lib/utils";

export interface EmptyStateProps extends React.HTMLAttributes<HTMLDivElement> {
  title: string;
  description: string;
  action?: React.ReactNode;
}

export function EmptyState({ action, className, description, title, ...props }: EmptyStateProps) {
  return (
    <div
      className={cn(
        "flex flex-col items-start gap-2 rounded-[14px] border border-dashed border-border bg-surface p-4 text-left",
        className,
      )}
      data-slot="empty-state"
      {...props}
    >
      <div className="text-[16px] leading-[24px] font-semibold text-text-primary">{title}</div>
      <p className="text-[13px] leading-[20px] text-text-secondary">{description}</p>
      {action ? <div className="pt-1">{action}</div> : null}
    </div>
  );
}
