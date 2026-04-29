import { cn as heroCn } from "@heroui/react";

export function cn(...values: Array<string | false | null | undefined>): string {
  return heroCn(...values) ?? "";
}
