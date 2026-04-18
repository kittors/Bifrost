import { Button } from "@bifrost/ui";

type PaginationBarProps = {
  onPageChange: (page: number) => void;
  page: number;
  pageSize: number;
  total: number;
};

export function PaginationBar({ onPageChange, page, pageSize, total }: PaginationBarProps) {
  const pageCount = Math.max(1, Math.ceil(total / pageSize));
  const safePage = Math.min(Math.max(page, 1), pageCount);

  return (
    <div className="flex flex-wrap items-center justify-between gap-3 rounded-[14px] border border-border bg-surface px-4 py-3">
      <div className="text-[12px] leading-[18px] text-text-secondary">
        第 {safePage} / {pageCount} 页，共 {total} 项
      </div>
      <div className="flex items-center gap-2">
        <Button
          disabled={safePage <= 1}
          onClick={() => onPageChange(safePage - 1)}
          size="sm"
          variant="secondary"
        >
          上一页
        </Button>
        <Button
          disabled={safePage >= pageCount}
          onClick={() => onPageChange(safePage + 1)}
          size="sm"
          variant="secondary"
        >
          下一页
        </Button>
      </div>
    </div>
  );
}
