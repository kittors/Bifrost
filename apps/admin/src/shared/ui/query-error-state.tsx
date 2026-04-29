import { normalizeUnknownError } from "../lib/http";
import { ErrorState } from "./error-state";

export function QueryErrorState({ error, title }: { error: unknown; title: string }) {
  const normalized = normalizeUnknownError(error);

  return (
    <ErrorState
      description={normalized.userMessage}
      requestId={normalized.requestId || undefined}
      title={title}
    />
  );
}
