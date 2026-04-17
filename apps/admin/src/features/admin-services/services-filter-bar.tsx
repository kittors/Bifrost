import { Input } from "@bifrost/ui";

type ServicesFilterBarProps = {
  keyword: string;
  onKeywordChange: (value: string) => void;
  onStatusChange: (value: string) => void;
  status: string;
};

export function ServicesFilterBar({
  keyword,
  onKeywordChange,
  onStatusChange,
  status,
}: ServicesFilterBarProps) {
  return (
    <section className="rounded-[14px] border border-border bg-surface p-4">
      <div className="flex flex-wrap items-center gap-2">
        <Input
          className="max-w-[280px]"
          onChange={(event) => {
            onKeywordChange(event.target.value);
          }}
          placeholder="搜索服务名或标识"
          value={keyword}
        />
        <select
          className="h-[32px] rounded-[6px] border border-border bg-surface px-3 text-[13px] leading-[20px]"
          onChange={(event) => {
            onStatusChange(event.target.value);
          }}
          value={status}
        >
          <option value="">全部状态</option>
          <option value="enabled">Enabled</option>
          <option value="disabled">Disabled</option>
        </select>
      </div>
    </section>
  );
}
