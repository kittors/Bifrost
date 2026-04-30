import { Input, ListBox, Select } from "@heroui/react";

type DevicesFilterBarProps = {
  keyword: string;
  onKeywordChange: (value: string) => void;
  onStatusChange: (value: string) => void;
  status: string;
};

export function DevicesFilterBar({
  keyword,
  onKeywordChange,
  onStatusChange,
  status,
}: DevicesFilterBarProps) {
  const selectedStatusKey = status || "all";

  return (
    <section className="rounded-[14px] border border-border bg-surface p-4">
      <div className="flex flex-wrap items-center gap-2">
        <Input
          className="max-w-[280px]"
          onChange={(event) => {
            onKeywordChange(event.target.value);
          }}
          placeholder="搜索设备名、用户名或指纹"
          value={keyword}
        />
        <Select
          aria-label="设备状态筛选"
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
              <ListBox.Item id="trusted">Trusted</ListBox.Item>
              <ListBox.Item id="disabled">Disabled</ListBox.Item>
            </ListBox>
          </Select.Popover>
        </Select>
      </div>
    </section>
  );
}
