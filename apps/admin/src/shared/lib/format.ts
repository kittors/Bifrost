export function formatList(values: string[]) {
  if (values.length === 0) {
    return "-";
  }

  return values.join(", ");
}

export function formatStatusLabel(status: string) {
  switch (status) {
    case "enabled":
      return "Enabled";
    case "disabled":
      return "Disabled";
    case "trusted":
      return "Trusted";
    default:
      return status || "-";
  }
}

export function formatAuditType(type: string) {
  return type.replaceAll(".", " / ");
}
