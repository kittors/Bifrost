import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { useDesktopSessionStore } from "../../entities/session/store";
import { renderWithQueryClient } from "../../test/render";
import { ServicesCard } from "./services-card";

const { createServiceAccessURLMock, listClientServicesMock, openExternalMock } = vi.hoisted(() => ({
  createServiceAccessURLMock: vi.fn(),
  listClientServicesMock: vi.fn(),
  openExternalMock: vi.fn().mockResolvedValue(undefined),
}));

vi.mock("../../entities/services/api", () => ({
  createServiceAccessURL: createServiceAccessURLMock,
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
    createServiceAccessURLMock.mockReset();
    openExternalMock.mockReset();
    openExternalMock.mockResolvedValue(undefined);

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
      openExternal: openExternalMock,
      session: {
        clear: vi.fn().mockResolvedValue(undefined),
        load: vi.fn().mockResolvedValue(null),
        save: vi.fn().mockResolvedValue(undefined),
      },
    };

    useDesktopSessionStore.setState({
      errorMessage: null,
      gatewayBaseURL: session.gatewayBaseURL,
      session: null,
      view: "services",
    });
  });

  it("renders nothing before a desktop session is available", () => {
    const { container } = renderWithQueryClient(<ServicesCard />);

    expect(container.innerHTML).toBe("");
    expect(listClientServicesMock).not.toHaveBeenCalled();
  });

  it("loads services and opens the gateway access URL without changing local network settings", async () => {
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
    createServiceAccessURLMock.mockResolvedValue({
      url: "http://127.0.0.1:8080/s/gitlab/",
    });
    useDesktopSessionStore.setState({ session });

    renderWithQueryClient(<ServicesCard />);

    expect(await screen.findByText("GitLab")).not.toBeNull();
    expect(screen.queryByText("engineering · role")).not.toBeNull();

    await userEvent.click(screen.getByRole("button", { name: "打开" }));

    await waitFor(() => {
      expect(openExternalMock).toHaveBeenCalledWith("http://127.0.0.1:8080/s/gitlab/");
    });
    expect(createServiceAccessURLMock).toHaveBeenCalledWith({
      accessToken: session.accessToken,
      baseURL: session.gatewayBaseURL,
      serviceId: "service_gitlab",
    });
  });
});
