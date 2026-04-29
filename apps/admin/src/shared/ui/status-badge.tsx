import { Chip } from "@heroui/react";

export function StatusBadge({ status }: { status: string }) {
  if (status === "enabled" || status === "trusted" || status === "success") {
    return (
      <Chip color="success" size="sm" variant="soft">
        {status}
      </Chip>
    );
  }

  if (status === "disabled" || status === "failure" || status === "denied") {
    return (
      <Chip color="danger" size="sm" variant="soft">
        {status}
      </Chip>
    );
  }

  return (
    <Chip color="default" size="sm" variant="soft">
      {status || "-"}
    </Chip>
  );
}
