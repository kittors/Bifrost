import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { useDesktopSessionStore } from "../../entities/session/store";
import { LoginCard } from "./login-card";

const { bootstrapClientDeviceMock, clientLoginMock, ensureLocalDeviceIdentityMock } = vi.hoisted(
  () => ({
    bootstrapClientDeviceMock: vi.fn(),
    clientLoginMock: vi.fn(),
    ensureLocalDeviceIdentityMock: vi.fn(),
  }),
);

vi.mock("../../entities/device/api", () => ({
  bootstrapClientDevice: bootstrapClientDeviceMock,
  ensureLocalDeviceIdentity: ensureLocalDeviceIdentityMock,
}));

vi.mock("../../entities/session/api", () => ({
  clientLogin: clientLoginMock,
}));

describe("LoginCard", () => {
  beforeEach(() => {
    useDesktopSessionStore.setState({
      errorMessage: null,
      gatewayBaseURL: "http://127.0.0.1:8080",
      session: null,
    });
  });

  it("renders the branded HeroUI login panel from the desktop mockup", () => {
    const { container } = render(<LoginCard />);

    expect(screen.queryByText("Bifrost Desktop")).not.toBeNull();
    expect(screen.queryByText("安全 · 稳定 · 高效")).not.toBeNull();
    expect(screen.queryByLabelText("服务端地址")).not.toBeNull();
    expect(screen.queryByLabelText("用户名")).not.toBeNull();
    expect(screen.queryByLabelText("密码")).not.toBeNull();
    expect(screen.queryByRole("button", { name: "登录" })).not.toBeNull();
    expect(screen.queryByRole("button", { name: "显示密码" })).not.toBeNull();
    expect(screen.queryByLabelText("Bifrost logo")).not.toBeNull();
    expect(container.querySelector('[data-slot="card"]')).not.toBeNull();
    expect(container.querySelector(".card--default")).not.toBeNull();
    expect(container.querySelectorAll('[data-slot="input-group"]')).toHaveLength(3);
    expect(container.querySelectorAll('[data-slot="input-group-prefix"]')).toHaveLength(3);
    expect(container.querySelectorAll('[data-slot="input-group-suffix"]')).toHaveLength(1);
    expect(container.querySelector(".bifrost-login-input-group")).toBeNull();
    expect(container.querySelector(".button--sm")).not.toBeNull();
    expect(container.querySelectorAll("svg").length).toBeGreaterThanOrEqual(5);
    expect(screen.queryByText("私有服务访问客户端")).toBeNull();
  });

  it("uses a simplified refined SVG logo mark", () => {
    render(<LoginCard />);

    const logo = screen.getByLabelText("Bifrost logo");

    expect(logo.querySelectorAll("circle").length).toBeLessThanOrEqual(3);
    expect(logo.querySelectorAll("path").length).toBeLessThanOrEqual(4);
  });

  it("toggles password visibility from the icon button", async () => {
    render(<LoginCard />);

    const passwordInput = screen.getByLabelText("密码");
    expect(passwordInput.getAttribute("type")).toBe("password");

    await userEvent.click(screen.getByRole("button", { name: "显示密码" }));

    expect(passwordInput.getAttribute("type")).toBe("text");
    expect(screen.queryByRole("button", { name: "隐藏密码" })).not.toBeNull();
  });

  it("shows a readable error when the bootstrap response has no device binding", async () => {
    ensureLocalDeviceIdentityMock.mockResolvedValue({
      deviceId: undefined,
      fingerprint: "fp_public_key_01",
      publicKey: "public-key",
    });
    bootstrapClientDeviceMock.mockResolvedValue({
      accessToken: "token",
      expiresIn: 900,
      refreshToken: "refresh",
      user: {
        displayName: "Alice",
        id: "user_alice",
        roles: ["role_developer"],
        username: "alice",
      },
    });

    render(<LoginCard />);
    await userEvent.click(screen.getByRole("button", { name: "登录" }));

    expect(await screen.findByText("服务端未返回设备绑定结果，请稍后重试")).not.toBeNull();
    expect(screen.queryByText(/Cannot read properties/)).toBeNull();
  });
});
