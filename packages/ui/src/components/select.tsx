import * as SelectPrimitive from "@radix-ui/react-select";
import * as React from "react";

import { cn } from "../lib/utils";

const SelectRoot = SelectPrimitive.Root;
const SelectValue = SelectPrimitive.Value;
const SelectGroup = SelectPrimitive.Group;
const SelectLabel = SelectPrimitive.Label;
const SelectSeparator = SelectPrimitive.Separator;

const SelectTrigger = React.forwardRef<
  React.ElementRef<typeof SelectPrimitive.Trigger>,
  React.ComponentPropsWithoutRef<typeof SelectPrimitive.Trigger>
>(({ className, children, ...props }, ref) => (
  <SelectPrimitive.Trigger
    className={cn(
      "flex h-[36px] w-full items-center justify-between rounded-[6px] border border-border bg-surface px-3 text-[14px] leading-[22px] text-text-primary outline-none placeholder:text-text-muted focus:border-brand",
      className,
    )}
    ref={ref}
    {...props}
  >
    {children}
  </SelectPrimitive.Trigger>
));

SelectTrigger.displayName = "SelectTrigger";

const SelectContent = React.forwardRef<
  React.ElementRef<typeof SelectPrimitive.Content>,
  React.ComponentPropsWithoutRef<typeof SelectPrimitive.Content>
>(({ className, children, ...props }, ref) => (
  <SelectPrimitive.Portal>
    <SelectPrimitive.Content
      className={cn(
        "overflow-hidden rounded-[10px] border border-border bg-surface p-1 shadow-sm",
        className,
      )}
      ref={ref}
      {...props}
    >
      <SelectPrimitive.Viewport className="p-0.5">{children}</SelectPrimitive.Viewport>
    </SelectPrimitive.Content>
  </SelectPrimitive.Portal>
));

SelectContent.displayName = "SelectContent";

const SelectItem = React.forwardRef<
  React.ElementRef<typeof SelectPrimitive.Item>,
  React.ComponentPropsWithoutRef<typeof SelectPrimitive.Item>
>(({ className, children, ...props }, ref) => (
  <SelectPrimitive.Item
    className={cn(
      "relative flex min-h-[32px] cursor-default select-none items-center rounded-[6px] px-2.5 text-[13px] leading-[20px] text-text-primary outline-none focus:bg-surface-2 data-[disabled]:pointer-events-none data-[disabled]:opacity-50",
      className,
    )}
    ref={ref}
    {...props}
  >
    <SelectPrimitive.ItemText>{children}</SelectPrimitive.ItemText>
  </SelectPrimitive.Item>
));

SelectItem.displayName = "SelectItem";

export const Select = {
  Root: SelectRoot,
  Trigger: SelectTrigger,
  Value: SelectValue,
  Content: SelectContent,
  Item: SelectItem,
  Group: SelectGroup,
  Label: SelectLabel,
  Separator: SelectSeparator,
};

export {
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectLabel,
  SelectRoot,
  SelectSeparator,
  SelectTrigger,
  SelectValue,
};
