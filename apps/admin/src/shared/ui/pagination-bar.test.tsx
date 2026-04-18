import { screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, expect, it, vi } from "vitest";

import { renderWithQueryClient } from "../../test/render";
import { PaginationBar } from "./pagination-bar";

describe("PaginationBar", () => {
  it("renders summary and triggers page changes", async () => {
    const user = userEvent.setup();
    const onPageChange = vi.fn();

    renderWithQueryClient(
      <PaginationBar onPageChange={onPageChange} page={2} pageSize={20} total={55} />,
    );

    expect(screen.getByText("第 2 / 3 页，共 55 项")).not.toBeNull();

    await user.click(screen.getByRole("button", { name: "上一页" }));
    await user.click(screen.getByRole("button", { name: "下一页" }));

    expect(onPageChange).toHaveBeenNthCalledWith(1, 1);
    expect(onPageChange).toHaveBeenNthCalledWith(2, 3);
  });
});
