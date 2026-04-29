import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { renderWithQueryClient } from "../../test/render";
import { LoginForm } from "./login-form";
import { useAdminSessionStore } from "./store";

const { adminLoginMock, navigateMock, toastSuccessMock } = vi.hoisted(() => ({
  adminLoginMock: vi.fn(),
  navigateMock: vi.fn(),
  toastSuccessMock: vi.fn(),
}));

vi.mock("@tanstack/react-router", async (importOriginal) => {
  const actual = await importOriginal<typeof import("@tanstack/react-router")>();

  return {
    ...actual,
    useRouter: () => ({
      navigate: navigateMock,
    }),
  };
});

vi.mock("sonner", () => ({
  toast: {
    success: toastSuccessMock,
  },
}));

vi.mock("./api", () => ({
  adminLogin: adminLoginMock,
}));

describe("LoginForm", () => {
  beforeEach(() => {
    localStorage.clear();
    navigateMock.mockReset();
    adminLoginMock.mockReset();
    toastSuccessMock.mockReset();
    useAdminSessionStore.setState({ session: null });
  });

  it("validates required credentials before calling the API", async () => {
    const user = userEvent.setup();
    renderWithQueryClient(<LoginForm />);

    await user.click(screen.getByRole("button", { name: "登录后台" }));

    expect(await screen.findByText("请输入用户名")).not.toBeNull();
    expect(screen.queryByText("请输入密码")).not.toBeNull();
    expect(adminLoginMock).not.toHaveBeenCalled();
  });

  it("persists session and navigates to dashboard after successful login", async () => {
    const user = userEvent.setup();
    const session = {
      accessToken: "access_admin",
      expiresAt: "2026-04-18T12:00:00.000Z",
      refreshToken: "refresh_admin",
      user: {
        displayName: "Administrator",
        id: "user_admin",
        roles: ["role_admin"],
        username: "admin",
      },
    };
    adminLoginMock.mockResolvedValue(session);

    renderWithQueryClient(<LoginForm />);

    await user.type(screen.getByLabelText("用户名"), "admin");
    await user.type(screen.getByLabelText("密码"), "ChangeMe123!");
    await user.click(screen.getByRole("button", { name: "登录后台" }));

    await waitFor(() => {
      expect(navigateMock).toHaveBeenCalledWith({ to: "/" });
    });

    expect(adminLoginMock).toHaveBeenCalledWith(
      {
        password: "ChangeMe123!",
        username: "admin",
      },
      expect.anything(),
    );
    expect(toastSuccessMock).toHaveBeenCalledWith("管理员会话已建立");
    expect(useAdminSessionStore.getState().session).toEqual(session);
  });
});
