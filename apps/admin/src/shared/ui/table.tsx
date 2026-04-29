import { Table as HeroTable } from "@heroui/react";
import { Children, isValidElement, type ComponentProps, type ReactNode } from "react";

import { cn } from "./style";

type TableRootProps = Omit<ComponentProps<typeof HeroTable>, "children"> & {
  children: ReactNode;
};

function composeHeroClassName<T>(
  baseClassName: string,
  className: ((values: T) => string) | string | undefined,
) {
  if (typeof className === "function") {
    return (values: T) => cn(baseClassName, className(values));
  }

  return cn(baseClassName, className);
}

function TableRoot({
  "aria-label": ariaLabel = "Bifrost 管理数据表",
  children,
  className,
  ...props
}: TableRootProps) {
  const tableChildren: ReactNode[] = [];
  let caption: ReactNode = null;

  Children.forEach(children, (child) => {
    if (isValidElement(child) && child.type === TableCaption) {
      caption = child;
      return;
    }

    tableChildren.push(child);
  });

  return (
    <HeroTable className={cn("w-full", className)} {...props}>
      <HeroTable.ScrollContainer className="overflow-x-auto">
        <HeroTable.Content aria-label={ariaLabel} className="w-full text-left">
          {tableChildren}
        </HeroTable.Content>
      </HeroTable.ScrollContainer>
      {caption}
    </HeroTable>
  );
}

function TableHeader({ className, ...props }: ComponentProps<typeof HeroTable.Header>) {
  return (
    <HeroTable.Header
      className={composeHeroClassName("[&_tr]:border-b [&_tr]:border-border", className)}
      {...props}
    />
  );
}

function TableBody({ className, ...props }: ComponentProps<typeof HeroTable.Body>) {
  return (
    <HeroTable.Body
      className={composeHeroClassName("[&_tr:last-child]:border-0", className)}
      {...props}
    />
  );
}

function TableRow({ className, ...props }: ComponentProps<typeof HeroTable.Row>) {
  return (
    <HeroTable.Row
      className={composeHeroClassName(
        "h-[36px] border-b border-border transition-colors",
        className,
      )}
      {...props}
    />
  );
}

function TableHead({ className, ...props }: ComponentProps<typeof HeroTable.Column>) {
  return (
    <HeroTable.Column
      className={composeHeroClassName(
        "h-[36px] px-3 text-[12px] leading-[18px] font-medium text-text-secondary",
        className,
      )}
      {...props}
    />
  );
}

function TableCell({ className, ...props }: ComponentProps<typeof HeroTable.Cell>) {
  return (
    <HeroTable.Cell
      className={composeHeroClassName(
        "px-3 text-[13px] leading-[20px] text-text-primary",
        className,
      )}
      {...props}
    />
  );
}

function TableCaption({ className, ...props }: ComponentProps<typeof HeroTable.Footer>) {
  return (
    <HeroTable.Footer
      className={cn("px-3 pb-3 pt-2 text-[12px] leading-[18px] text-text-muted", className)}
      {...props}
    />
  );
}

export const Table = {
  Root: TableRoot,
  Header: TableHeader,
  Body: TableBody,
  Row: TableRow,
  Head: TableHead,
  Cell: TableCell,
  Caption: TableCaption,
};
