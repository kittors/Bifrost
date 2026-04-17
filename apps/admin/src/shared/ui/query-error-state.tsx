import { ErrorState } from "@bifrost/ui";

import { normalizeUnknownError } from "../lib/http";

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
