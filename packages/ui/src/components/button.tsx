import { cva, type VariantProps } from "class-variance-authority";
import * as React from "react";

import { cn } from "../lib/utils";

const buttonVariants = cva(
  "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-[10px] border border-transparent px-3 text-[14px] leading-[22px] font-medium transition-colors outline-none disabled:pointer-events-none disabled:opacity-50",
  {
    variants: {
      variant: {
        primary: "bg-brand text-white hover:bg-brand-hover",
        secondary: "bg-surface text-text-primary border-border hover:bg-surface-2",
        ghost: "bg-transparent text-text-secondary hover:bg-surface-2 hover:text-text-primary",
        danger: "bg-danger text-white hover:opacity-90",
      },
      size: {
        xs: "h-[28px] px-[10px] text-[12px] leading-[18px]",
        sm: "h-[32px] px-3 text-[13px] leading-[20px]",
        md: "h-[36px] px-[14px] text-[14px] leading-[22px]",
        lg: "h-[40px] px-4 text-[14px] leading-[22px]",
      },
    },
    defaultVariants: {
      variant: "primary",
      size: "md",
    },
  },
);

export interface ButtonProps
  extends React.ButtonHTMLAttributes<HTMLButtonElement>,
    VariantProps<typeof buttonVariants> {}

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, size, type = "button", variant, ...props }, ref) => (
    <button
      className={cn(buttonVariants({ size, variant }), className)}
      data-slot="button"
      ref={ref}
      type={type}
      {...props}
    />
  ),
);

Button.displayName = "Button";
