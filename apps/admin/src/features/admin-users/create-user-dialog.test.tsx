import { fireEvent, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { renderWithQueryClient } from "../../test/render";
import { CreateUserDialog } from "./create-user-dialog";

const { createAdminUserMock, toastSuccessMock } = vi.hoisted(() => ({
  createAdminUserMock: vi.fn(),
  toastSuccessMock: vi.fn(),
}));

vi.mock("@heroui/react", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@heroui/react")>();

  return {
    ...actual,
    toast: {
      ...actual.toast,
      success: toastSuccessMock,
    },
  };
});

vi.mock("../../entities/admin/api", () => ({
  createAdminUser: createAdminUserMock,
}));

const roleOptions = [
  {
    description: "Developer access",
    displayName: "Developer",
    id: "role_developer",
    name: "developer",
  },
];

describe("CreateUserDialog", () => {
  beforeEach(() => {
    createAdminUserMock.mockReset();
    toastSuccessMock.mockReset();
  });

  it("closes the dialog when the operator cancels creation", () => {
    const onOpenChange = vi.fn();
    renderWithQueryClient(
      <CreateUserDialog
        accessToken="access_admin"
        onCreated={vi.fn()}
        onOpenChange={onOpenChange}
        open
        roleOptions={roleOptions}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "取消" }));

    expect(onOpenChange).toHaveBeenCalledWith(false);
    expect(createAdminUserMock).not.toHaveBeenCalled();
  });

  it("submits selected role and closes after successful creation", async () => {
    const onCreated = vi.fn();
    const onOpenChange = vi.fn();
    createAdminUserMock.mockResolvedValue({
      id: "user_alice",
      username: "alice",
    });

    renderWithQueryClient(
      <CreateUserDialog
        accessToken="access_admin"
        onCreated={onCreated}
        onOpenChange={onOpenChange}
        open
        roleOptions={roleOptions}
      />,
    );

    fireEvent.change(screen.getByLabelText("用户名"), { target: { value: "alice" } });
    fireEvent.change(screen.getByLabelText("显示名"), { target: { value: "Alice" } });
    fireEvent.change(screen.getByLabelText("邮箱"), { target: { value: "alice@example.com" } });
    fireEvent.change(screen.getByLabelText("初始密码"), { target: { value: "ChangeMe123!" } });
    fireEvent.click(screen.getByRole("checkbox"));

    const form = screen.getByRole("dialog", { name: "创建用户" }).querySelector("form");
    if (!form) {
      throw new Error("create user form was not rendered");
    }
    fireEvent.submit(form);

    await waitFor(() => {
      expect(createAdminUserMock).toHaveBeenCalledWith({
        accessToken: "access_admin",
        displayName: "Alice",
        email: "alice@example.com",
        password: "ChangeMe123!",
        roleIds: ["role_developer"],
        username: "alice",
      });
    });

    expect(toastSuccessMock).toHaveBeenCalledWith("用户已创建");
    expect(onOpenChange).toHaveBeenCalledWith(false);
    expect(onCreated).toHaveBeenCalledTimes(1);
  });
});
