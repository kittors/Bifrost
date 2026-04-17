import { Badge } from "@bifrost/ui";

export function StatusBadge({ status }: { status: string }) {
  if (status === "enabled" || status === "trusted" || status === "success") {
    return <Badge variant="success">{status}</Badge>;
  }

  if (status === "disabled" || status === "failure" || status === "denied") {
    return <Badge variant="danger">{status}</Badge>;
  }

  return <Badge variant="neutral">{status || "-"}</Badge>;
}
