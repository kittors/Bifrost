import * as DialogPrimitive from "@radix-ui/react-dialog";
import * as React from "react";

import { cn } from "../lib/utils";

type DrawerSide = "left" | "right";

const DrawerRoot = DialogPrimitive.Root;
const DrawerTrigger = DialogPrimitive.Trigger;
const DrawerClose = DialogPrimitive.Close;
const DrawerPortal = DialogPrimitive.Portal;

const DrawerOverlay = React.forwardRef<
  React.ElementRef<typeof DialogPrimitive.Overlay>,
  React.ComponentPropsWithoutRef<typeof DialogPrimitive.Overlay>
>(({ className, ...props }, ref) => (
  <DialogPrimitive.Overlay
    className={cn(
      "fixed inset-0 bg-[color-mix(in_oklab,var(--bifrost-text-primary)_14%,transparent)]",
      className,
    )}
    ref={ref}
    {...props}
  />
));

DrawerOverlay.displayName = "DrawerOverlay";

const DrawerContent = React.forwardRef<
  React.ElementRef<typeof DialogPrimitive.Content>,
  React.ComponentPropsWithoutRef<typeof DialogPrimitive.Content> & { side?: DrawerSide }
>(({ className, side = "right", ...props }, ref) => (
  <DrawerPortal>
    <DrawerOverlay />
    <DialogPrimitive.Content
      className={cn(
        "fixed top-0 z-50 flex h-full w-[min(420px,calc(100vw-24px))] flex-col border-border bg-surface p-4 text-text-primary shadow-[0_18px_60px_-24px_rgba(15,23,42,0.35)] outline-none",
        side === "right" ? "right-0 border-l" : "left-0 border-r",
        className,
      )}
      ref={ref}
      {...props}
    />
  </DrawerPortal>
));

DrawerContent.displayName = "DrawerContent";

function DrawerHeader({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn("flex flex-col gap-1.5", className)} {...props} />;
}

function DrawerFooter({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return (
    <div className={cn("mt-auto flex items-center justify-end gap-2 pt-4", className)} {...props} />
  );
}

const DrawerTitle = React.forwardRef<
  React.ElementRef<typeof DialogPrimitive.Title>,
  React.ComponentPropsWithoutRef<typeof DialogPrimitive.Title>
>(({ className, ...props }, ref) => (
  <DialogPrimitive.Title
    className={cn("text-[16px] leading-[24px] font-semibold", className)}
    ref={ref}
    {...props}
  />
));

DrawerTitle.displayName = "DrawerTitle";

const DrawerDescription = React.forwardRef<
  React.ElementRef<typeof DialogPrimitive.Description>,
  React.ComponentPropsWithoutRef<typeof DialogPrimitive.Description>
>(({ className, ...props }, ref) => (
  <DialogPrimitive.Description
    className={cn("text-[13px] leading-[20px] text-text-secondary", className)}
    ref={ref}
    {...props}
  />
));

DrawerDescription.displayName = "DrawerDescription";

export const Drawer = {
  Root: DrawerRoot,
  Trigger: DrawerTrigger,
  Close: DrawerClose,
  Overlay: DrawerOverlay,
  Content: DrawerContent,
  Header: DrawerHeader,
  Footer: DrawerFooter,
  Title: DrawerTitle,
  Description: DrawerDescription,
};

export {
  DrawerClose,
  DrawerContent,
  DrawerDescription,
  DrawerFooter,
  DrawerHeader,
  DrawerOverlay,
  DrawerTitle,
  DrawerTrigger,
};
