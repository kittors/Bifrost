import { Button, Input, ListBox, Select } from "@heroui/react";

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
  const selectedStatusKey = status || "all";

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
        <Select
          aria-label="用户状态筛选"
          className="w-[150px]"
          onSelectionChange={(key) => {
            const value = String(key);
            onStatusChange(value === "all" ? "" : value);
          }}
          selectedKey={selectedStatusKey}
        >
          <Select.Trigger>
            <Select.Value />
            <Select.Indicator />
          </Select.Trigger>
          <Select.Popover>
            <ListBox>
              <ListBox.Item id="all">全部状态</ListBox.Item>
              <ListBox.Item id="enabled">Enabled</ListBox.Item>
              <ListBox.Item id="disabled">Disabled</ListBox.Item>
            </ListBox>
          </Select.Popover>
        </Select>
        {(keyword || status) && (
          <Button onClick={onReset} size="sm" variant="ghost">
            清空筛选
          </Button>
        )}
      </div>
    </section>
  );
}
