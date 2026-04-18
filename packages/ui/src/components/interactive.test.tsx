import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { Button, Dialog, Input } from "../index";

describe("interactive primitives", () => {
  it("keeps compact button semantics while forwarding click handlers", async () => {
    const onClick = vi.fn();

    render(
      <Button onClick={onClick} size="sm">
        打开服务
      </Button>,
    );

    await userEvent.click(screen.getByRole("button", { name: "打开服务" }));

    const button = screen.getByRole("button", { name: "打开服务" });
    expect(onClick).toHaveBeenCalledTimes(1);
    expect(button.className).toContain("h-[32px]");
  });

  it("opens dialog content through the Radix trigger composition", async () => {
    render(
      <Dialog.Root>
        <Dialog.Trigger asChild>
          <Button>创建用户</Button>
        </Dialog.Trigger>
        <Dialog.Content>
          <Dialog.Title>创建用户</Dialog.Title>
          <Dialog.Description>填写账号资料并分配角色。</Dialog.Description>
          <Input aria-label="用户名" />
        </Dialog.Content>
      </Dialog.Root>,
    );

    await userEvent.click(screen.getByRole("button", { name: "创建用户" }));

    expect(screen.queryByRole("dialog", { name: "创建用户" })).not.toBeNull();
    expect(screen.queryByLabelText("用户名")).not.toBeNull();
  });
});
