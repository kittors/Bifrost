import type * as React from "react";

import { cn } from "../lib/utils";

function TableRoot({ className, ...props }: React.TableHTMLAttributes<HTMLTableElement>) {
  return (
    <table
      className={cn("w-full caption-bottom text-left", className)}
      data-slot="table"
      {...props}
    />
  );
}

function TableHeader({ className, ...props }: React.HTMLAttributes<HTMLTableSectionElement>) {
  return <thead className={cn("[&_tr]:border-b [&_tr]:border-border", className)} {...props} />;
}

function TableBody({ className, ...props }: React.HTMLAttributes<HTMLTableSectionElement>) {
  return <tbody className={cn("[&_tr:last-child]:border-0", className)} {...props} />;
}

function TableRow({ className, ...props }: React.HTMLAttributes<HTMLTableRowElement>) {
  return (
    <tr className={cn("h-[36px] border-b border-border transition-colors", className)} {...props} />
  );
}

function TableHead({ className, ...props }: React.ThHTMLAttributes<HTMLTableCellElement>) {
  return (
    <th
      className={cn(
        "h-[36px] px-3 text-[12px] leading-[18px] font-medium text-text-secondary",
        className,
      )}
      {...props}
    />
  );
}

function TableCell({ className, ...props }: React.TdHTMLAttributes<HTMLTableCellElement>) {
  return (
    <td className={cn("px-3 text-[13px] leading-[20px] text-text-primary", className)} {...props} />
  );
}

function TableCaption({ className, ...props }: React.HTMLAttributes<HTMLTableCaptionElement>) {
  return (
    <caption
      className={cn("mt-3 text-[12px] leading-[18px] text-text-muted", className)}
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

export { TableBody, TableCaption, TableCell, TableHead, TableHeader, TableRoot, TableRow };
