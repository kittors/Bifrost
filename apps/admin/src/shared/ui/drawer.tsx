import { Drawer as HeroDrawer } from "@heroui/react";
import type { HTMLAttributes, ReactNode } from "react";

import { cn } from "./style";

function DrawerRoot({
  children,
  onOpenChange,
  open,
}: {
  children: ReactNode;
  onOpenChange?: (open: boolean) => void;
  open?: boolean;
}) {
  return (
    <HeroDrawer isOpen={open} onOpenChange={onOpenChange}>
      {children}
    </HeroDrawer>
  );
}

function DrawerContent({ className, ...props }: HTMLAttributes<HTMLDivElement>) {
  return (
    <HeroDrawer.Backdrop>
      <HeroDrawer.Content
        className={cn("w-[min(420px,calc(100vw-24px))]", className)}
        placement="right"
      >
        <HeroDrawer.Dialog>
          <HeroDrawer.CloseTrigger />
          <div {...props} />
        </HeroDrawer.Dialog>
      </HeroDrawer.Content>
    </HeroDrawer.Backdrop>
  );
}

function DrawerHeader({ className, ...props }: HTMLAttributes<HTMLDivElement>) {
  return <HeroDrawer.Header className={cn("flex flex-col gap-1.5", className)} {...props} />;
}

function DrawerFooter({ className, ...props }: HTMLAttributes<HTMLDivElement>) {
  return (
    <HeroDrawer.Footer
      className={cn("mt-auto flex items-center justify-end gap-2 pt-4", className)}
      {...props}
    />
  );
}

function DrawerTitle({ className, ...props }: HTMLAttributes<HTMLHeadingElement>) {
  return (
    <HeroDrawer.Heading
      className={cn("text-[16px] leading-[24px] font-semibold", className)}
      {...props}
    />
  );
}

function DrawerDescription({ className, ...props }: HTMLAttributes<HTMLParagraphElement>) {
  return (
    <p className={cn("text-[13px] leading-[20px] text-text-secondary", className)} {...props} />
  );
}

export const Drawer = {
  Root: DrawerRoot,
  Content: DrawerContent,
  Header: DrawerHeader,
  Footer: DrawerFooter,
  Title: DrawerTitle,
  Description: DrawerDescription,
};
