import { Input, ListBox, Select } from "@heroui/react";

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
  const selectedStatusKey = status || "all";

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
        <Select
          aria-label="服务状态筛选"
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
      </div>
    </section>
  );
}
