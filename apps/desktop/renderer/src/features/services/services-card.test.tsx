import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { useDesktopSessionStore } from "../../entities/session/store";
import { ApiClientError } from "../../shared/lib/http";
import { renderWithQueryClient } from "../../test/render";
import { ServicesCard } from "./services-card";

const { listClientServicesMock, openServiceMock } = vi.hoisted(() => ({
  listClientServicesMock: vi.fn(),
  openServiceMock: vi.fn().mockResolvedValue("http://127.0.0.1:18080/s/gitlab/"),
}));

vi.mock("../../entities/services/api", () => ({
  listClientServices: listClientServicesMock,
}));

const session = {
  accessToken: "access_alice",
  deviceId: "device_alice",
  expiresAt: "2026-04-18T12:00:00.000Z",
  gatewayBaseURL: "http://127.0.0.1:8080",
  refreshToken: "refresh_alice",
  user: {
    displayName: "Alice",
    id: "user_alice",
    roles: ["role_developer"],
    username: "alice",
  },
};

describe("ServicesCard", () => {
  beforeEach(() => {
    listClientServicesMock.mockReset();
    openServiceMock.mockReset();
    openServiceMock.mockResolvedValue("http://127.0.0.1:18080/s/gitlab/");

    window.bifrostDesktop = {
      app: {
        getInfo: vi.fn(),
      },
      device: {
        attach: vi.fn(),
        clear: vi.fn().mockResolvedValue(undefined),
        ensure: vi.fn(),
        load: vi.fn(),
        signChallenge: vi.fn(),
      },
      diagnostics: {
        snapshot: vi.fn(),
      },
      localProxy: {
        openService: openServiceMock,
        start: vi.fn(),
        status: vi.fn(),
        stop: vi.fn(),
      },
      openExternal: vi.fn(),
      session: {
        clear: vi.fn().mockResolvedValue(undefined),
        load: vi.fn().mockResolvedValue(null),
        save: vi.fn().mockResolvedValue(undefined),
      },
    };

    useDesktopSessionStore.setState({
      errorMessage: null,
      gatewayBaseURL: session.gatewayBaseURL,
      localProxyStatus: {
        baseURL: "http://127.0.0.1:18080",
        host: "127.0.0.1",
        port: 18080,
        running: true,
      },
      session: null,
      view: "services",
    });
  });

  it("renders nothing before a desktop session is available", () => {
    const { container } = renderWithQueryClient(<ServicesCard />);

    expect(container.innerHTML).toBe("");
    expect(listClientServicesMock).not.toHaveBeenCalled();
  });

  it("loads services and opens the local proxy service URL without changing local network settings", async () => {
    listClientServicesMock.mockResolvedValue([
      {
        accessSource: "role",
        group: "engineering",
        id: "service_gitlab",
        key: "gitlab",
        name: "GitLab",
        status: "enabled",
      },
    ]);
    useDesktopSessionStore.setState({ session });

    renderWithQueryClient(<ServicesCard />);

    expect(await screen.findByText("GitLab")).not.toBeNull();
    expect(screen.queryByText("engineering · role")).not.toBeNull();
    expect(screen.queryByText("本地入口：http://127.0.0.1:18080/s/gitlab/")).not.toBeNull();

    await userEvent.click(screen.getByRole("button", { name: "打开" }));

    await waitFor(() => {
      expect(openServiceMock).toHaveBeenCalledWith("/s/gitlab/");
    });
  });

  it("renders the shared error state with requestId when the service list cannot load", async () => {
    listClientServicesMock.mockRejectedValue(
      new ApiClientError({
        code: "POLICY_ACCESS_DENIED",
        message: "policy denied",
        requestId: "req_services_denied",
        statusCode: 403,
        userMessage: "当前账号没有可访问服务",
      }),
    );
    useDesktopSessionStore.setState({ session });

    renderWithQueryClient(<ServicesCard />);

    expect(await screen.findByText("服务列表不可用")).not.toBeNull();
    expect(screen.queryByText("当前账号没有可访问服务")).not.toBeNull();
    expect(screen.queryByText("Request ID: req_services_denied")).not.toBeNull();
  });
});
