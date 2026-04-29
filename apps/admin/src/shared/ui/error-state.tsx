import { Alert } from "@heroui/react";
import type { ReactNode } from "react";

type ErrorStateProps = {
  action?: ReactNode;
  className?: string;
  description: string;
  requestId?: string;
  title: string;
};

export function ErrorState({ action, className, description, requestId, title }: ErrorStateProps) {
  return (
    <Alert className={className} status="danger">
      <Alert.Content>
        <Alert.Title>{title}</Alert.Title>
        <Alert.Description>{description}</Alert.Description>
        {requestId ? (
          <div className="mt-1 rounded-[6px] bg-surface-2 px-2 py-1 font-mono text-[12px] leading-[18px] text-text-secondary">
            Request ID: {requestId}
          </div>
        ) : null}
        {action ? <div className="pt-1">{action}</div> : null}
      </Alert.Content>
    </Alert>
  );
}
