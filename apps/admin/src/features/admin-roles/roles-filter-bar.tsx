import { Button, Input } from "@heroui/react";

type RolesFilterBarProps = {
  keyword: string;
  onKeywordChange: (value: string) => void;
  onReset: () => void;
};

export function RolesFilterBar({ keyword, onKeywordChange, onReset }: RolesFilterBarProps) {
  return (
    <section className="rounded-[14px] border border-border bg-surface p-4">
      <div className="flex flex-wrap items-center gap-2">
        <Input
          className="max-w-[280px]"
          onChange={(event) => {
            onKeywordChange(event.target.value);
          }}
          placeholder="搜索角色名或显示名"
          value={keyword}
        />
        {keyword ? (
          <Button onClick={onReset} size="sm" variant="ghost">
            清空筛选
          </Button>
        ) : null}
      </div>
    </section>
  );
}
