import * as TabsPrimitive from "@radix-ui/react-tabs";
import * as React from "react";

import { cn } from "../lib/utils";

function TabsRoot(props: React.ComponentPropsWithoutRef<typeof TabsPrimitive.Root>) {
  return <TabsPrimitive.Root {...props} />;
}

const TabsList = React.forwardRef<
  React.ElementRef<typeof TabsPrimitive.List>,
  React.ComponentPropsWithoutRef<typeof TabsPrimitive.List>
>(({ className, ...props }, ref) => (
  <TabsPrimitive.List
    className={cn("inline-flex h-[32px] items-center rounded-[10px] bg-surface-2 p-1", className)}
    ref={ref}
    {...props}
  />
));

TabsList.displayName = "TabsList";

const TabsTrigger = React.forwardRef<
  React.ElementRef<typeof TabsPrimitive.Trigger>,
  React.ComponentPropsWithoutRef<typeof TabsPrimitive.Trigger>
>(({ className, ...props }, ref) => (
  <TabsPrimitive.Trigger
    className={cn(
      "inline-flex min-w-[72px] items-center justify-center rounded-[6px] px-2.5 text-[13px] leading-[20px] font-medium text-text-secondary outline-none transition-colors data-[state=active]:bg-surface data-[state=active]:text-text-primary",
      className,
    )}
    ref={ref}
    {...props}
  />
));

TabsTrigger.displayName = "TabsTrigger";

const TabsContent = React.forwardRef<
  React.ElementRef<typeof TabsPrimitive.Content>,
  React.ComponentPropsWithoutRef<typeof TabsPrimitive.Content>
>(({ className, ...props }, ref) => (
  <TabsPrimitive.Content className={cn("mt-3 outline-none", className)} ref={ref} {...props} />
));

TabsContent.displayName = "TabsContent";

export const Tabs = {
  Root: TabsRoot,
  List: TabsList,
  Trigger: TabsTrigger,
  Content: TabsContent,
};

export { TabsContent, TabsList, TabsRoot, TabsTrigger };
