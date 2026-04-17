import { Button, Input } from "@bifrost/ui";

type UsersFilterBarProps = {
  keyword: string;
  onKeywordChange: (value: string) => void;
  onReset: () => void;
  onStatusChange: (value: string) => void;
  status: string;
};

// 筛选条只处理输入与筛选条件，不感知任何列表查询实现。
export function UsersFilterBar({
  keyword,
  onKeywordChange,
  onReset,
  onStatusChange,
  status,
}: UsersFilterBarProps) {
  return (
    <section className="rounded-[14px] border border-border bg-surface p-4">
      <div className="flex flex-wrap items-center gap-2">
        <Input
          className="max-w-[280px]"
          onChange={(event) => {
            onKeywordChange(event.target.value);
          }}
          placeholder="搜索用户名、显示名或邮箱"
          value={keyword}
        />
        <select
          className="h-[32px] rounded-[6px] border border-border bg-surface px-3 text-[13px] leading-[20px] text-text-primary"
          onChange={(event) => {
            onStatusChange(event.target.value);
          }}
          value={status}
        >
          <option value="">全部状态</option>
          <option value="enabled">Enabled</option>
          <option value="disabled">Disabled</option>
        </select>
        {(keyword || status) && (
          <Button onClick={onReset} size="sm" variant="ghost">
            清空筛选
          </Button>
        )}
      </div>
    </section>
  );
}
