import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { render, screen, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { useDesktopSessionStore } from "../entities/session/store";
import { DesktopApp } from "./app";

function renderDesktopApp() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  });

  return render(
    <QueryClientProvider client={queryClient}>
      <DesktopApp />
    </QueryClientProvider>,
  );
}

describe("DesktopApp", () => {
  beforeEach(() => {
    useDesktopSessionStore.setState({
      device: null,
      errorMessage: null,
      gatewayBaseURL: "http://127.0.0.1:8080",
      isHydrating: true,
      session: null,
      view: "services",
    });

    window.bifrostDesktop = {
      app: {
        getInfo: vi.fn(),
      },
      device: {
        attach: vi.fn(),
        clear: vi.fn(),
        ensure: vi.fn(),
        load: vi.fn().mockRejectedValue(new Error("missing local identity")),
        signChallenge: vi.fn(),
      },
      diagnostics: {
        snapshot: vi.fn().mockResolvedValue({
          dnsManagedByBifrost: false,
          encryptionAvailable: true,
          nodeIntegration: false,
          platform: "darwin",
          proxyManagedByBifrost: false,
          routeManagedByBifrost: false,
        }),
      },
      localProxy: {
        openService: vi.fn(),
        start: vi.fn(),
        status: vi.fn(),
        stop: vi.fn().mockResolvedValue(undefined),
      },
      openExternal: vi.fn(),
      session: {
        clear: vi.fn().mockResolvedValue(undefined),
        load: vi.fn().mockResolvedValue(null),
        save: vi.fn(),
      },
    };
  });

  it("does not show a local device identity error before first login", async () => {
    renderDesktopApp();

    await waitFor(() => {
      expect(window.bifrostDesktop.device.load).toHaveBeenCalledTimes(1);
    });

    await expect(screen.findByText("本地设备身份读取失败", {}, { timeout: 100 })).rejects.toThrow();
    expect(screen.getByRole("button", { name: "登录" })).not.toBeNull();
  });
});
