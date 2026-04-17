import { cva, type VariantProps } from "class-variance-authority";
import * as React from "react";

import { cn } from "../lib/utils";

const inputVariants = cva(
  "flex w-full rounded-[6px] border border-border bg-surface px-3 text-[14px] leading-[22px] text-text-primary shadow-none transition-colors outline-none placeholder:text-text-muted focus:border-brand focus:ring-0 disabled:cursor-not-allowed disabled:opacity-60",
  {
    variants: {
      size: {
        sm: "h-[32px] text-[13px] leading-[20px]",
        md: "h-[36px]",
        lg: "h-[40px]",
      },
    },
    defaultVariants: {
      size: "md",
    },
  },
);

export interface InputProps
  extends Omit<React.InputHTMLAttributes<HTMLInputElement>, "size">,
    VariantProps<typeof inputVariants> {}

export const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, size, ...props }, ref) => (
    <input
      className={cn(inputVariants({ size }), className)}
      data-slot="input"
      ref={ref}
      {...props}
    />
  ),
);

Input.displayName = "Input";
