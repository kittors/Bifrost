import * as DialogPrimitive from "@radix-ui/react-dialog";
import * as React from "react";

import { cn } from "../lib/utils";

const DialogRoot = DialogPrimitive.Root;
const DialogTrigger = DialogPrimitive.Trigger;
const DialogClose = DialogPrimitive.Close;
const DialogPortal = DialogPrimitive.Portal;

const DialogOverlay = React.forwardRef<
  React.ElementRef<typeof DialogPrimitive.Overlay>,
  React.ComponentPropsWithoutRef<typeof DialogPrimitive.Overlay>
>(({ className, ...props }, ref) => (
  <DialogPrimitive.Overlay
    className={cn(
      "fixed inset-0 bg-[color-mix(in_oklab,var(--bifrost-text-primary)_18%,transparent)]",
      className,
    )}
    ref={ref}
    {...props}
  />
));

DialogOverlay.displayName = "DialogOverlay";

const DialogContent = React.forwardRef<
  React.ElementRef<typeof DialogPrimitive.Content>,
  React.ComponentPropsWithoutRef<typeof DialogPrimitive.Content>
>(({ className, ...props }, ref) => (
  <DialogPortal>
    <DialogOverlay />
    <DialogPrimitive.Content
      className={cn(
        "fixed top-1/2 left-1/2 w-[min(560px,calc(100vw-32px))] -translate-x-1/2 -translate-y-1/2 rounded-[14px] border border-border bg-surface p-4 text-text-primary shadow-[0_18px_60px_-24px_rgba(15,23,42,0.35)] outline-none",
        className,
      )}
      ref={ref}
      {...props}
    />
  </DialogPortal>
));

DialogContent.displayName = "DialogContent";

function DialogHeader({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn("flex flex-col gap-1.5", className)} {...props} />;
}

function DialogFooter({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn("flex items-center justify-end gap-2 pt-4", className)} {...props} />;
}

const DialogTitle = React.forwardRef<
  React.ElementRef<typeof DialogPrimitive.Title>,
  React.ComponentPropsWithoutRef<typeof DialogPrimitive.Title>
>(({ className, ...props }, ref) => (
  <DialogPrimitive.Title
    className={cn("text-[16px] leading-[24px] font-semibold", className)}
    ref={ref}
    {...props}
  />
));

DialogTitle.displayName = "DialogTitle";

const DialogDescription = React.forwardRef<
  React.ElementRef<typeof DialogPrimitive.Description>,
  React.ComponentPropsWithoutRef<typeof DialogPrimitive.Description>
>(({ className, ...props }, ref) => (
  <DialogPrimitive.Description
    className={cn("text-[13px] leading-[20px] text-text-secondary", className)}
    ref={ref}
    {...props}
  />
));

DialogDescription.displayName = "DialogDescription";

export const Dialog = {
  Root: DialogRoot,
  Trigger: DialogTrigger,
  Close: DialogClose,
  Overlay: DialogOverlay,
  Content: DialogContent,
  Header: DialogHeader,
  Footer: DialogFooter,
  Title: DialogTitle,
  Description: DialogDescription,
};

export {
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogOverlay,
  DialogTitle,
  DialogTrigger,
};
