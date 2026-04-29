import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { WindowShell } from "./window-shell";

describe("WindowShell", () => {
  it("renders only the login content surface and leaves window chrome to Electron", () => {
    const { container } = render(
      <WindowShell mode="login">
        <div>Login content</div>
      </WindowShell>,
    );

    expect(screen.queryByText("Login content")).not.toBeNull();
    expect(container.querySelector('[data-testid="desktop-login-surface"]')).not.toBeNull();
    expect(container.querySelectorAll('[data-testid^="traffic-light-"]')).toHaveLength(0);
  });

  it("does not render a fixed-size backdrop fill in login mode", () => {
    const { container } = render(
      <WindowShell mode="login">
        <div>Login content</div>
      </WindowShell>,
    );

    expect(container.querySelector('[data-testid="login-network-backdrop"]')).not.toBeNull();
    expect(container.querySelector('[data-testid="login-network-backdrop"] rect')).toBeNull();
  });
});
