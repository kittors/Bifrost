import { cva, type VariantProps } from "class-variance-authority";
import type * as React from "react";

import { cn } from "../lib/utils";

const badgeVariants = cva(
  "inline-flex items-center rounded-[6px] border px-2 py-0.5 text-[12px] leading-[18px] font-medium transition-colors",
  {
    variants: {
      variant: {
        neutral: "border-border bg-surface text-text-secondary",
        brand: "border-transparent bg-brand-soft text-brand",
        success:
          "border-transparent bg-[color-mix(in_oklab,var(--bifrost-success)_18%,white)] text-success",
        warning:
          "border-transparent bg-[color-mix(in_oklab,var(--bifrost-warning)_18%,white)] text-warning",
        danger:
          "border-transparent bg-[color-mix(in_oklab,var(--bifrost-danger)_18%,white)] text-danger",
      },
    },
    defaultVariants: {
      variant: "neutral",
    },
  },
);

export interface BadgeProps
  extends React.HTMLAttributes<HTMLSpanElement>,
    VariantProps<typeof badgeVariants> {}

export function Badge({ className, variant, ...props }: BadgeProps) {
  return (
    <span className={cn(badgeVariants({ variant }), className)} data-slot="badge" {...props} />
  );
}
