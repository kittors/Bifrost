import { EmptyState as HeroEmptyState } from "@heroui/react";
import type { ReactNode } from "react";

type EmptyStateProps = {
  action?: ReactNode;
  className?: string;
  description: string;
  title: string;
};

export function EmptyState({ action, className, description, title }: EmptyStateProps) {
  return (
    <HeroEmptyState className={className}>
      <div className="text-[16px] leading-[24px] font-semibold text-text-primary">{title}</div>
      <p className="text-[13px] leading-[20px] text-text-secondary">{description}</p>
      {action ? <div className="pt-1">{action}</div> : null}
    </HeroEmptyState>
  );
}
