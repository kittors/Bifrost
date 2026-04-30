import { screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { renderWithQueryClient } from "../test/render";
import { LoginPage } from "./login-page";

const { navigateMock } = vi.hoisted(() => ({
  navigateMock: vi.fn(),
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

vi.mock("../features/auth/api", () => ({
  adminLogin: vi.fn(),
}));

describe("LoginPage", () => {
  it("renders the screenshot-inspired admin login composition", () => {
    renderWithQueryClient(<LoginPage />);

    expect(screen.getByRole("region", { name: "后台登录产品说明" })).not.toBeNull();
    expect(screen.getByRole("region", { name: "管理员登录表单" })).not.toBeNull();
    expect(
      screen.getByRole("heading", {
        name: "为私有服务访问配置一个清爽、可追溯的控制面。",
      }),
    ).not.toBeNull();

    for (const cardTitle of ["Users", "Services", "Audit"]) {
      expect(screen.getByText(cardTitle)).not.toBeNull();
    }
  });
});
