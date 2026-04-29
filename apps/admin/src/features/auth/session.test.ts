import { describe, expect, it } from "vitest";

import {
  clearStoredAdminSession,
  loadStoredAdminSession,
  type StoredAdminSession,
  saveStoredAdminSession,
} from "./session";

function createStorage() {
  const values = new Map<string, string>();

  return {
    getItem(key: string) {
      return values.get(key) ?? null;
    },
    removeItem(key: string) {
      values.delete(key);
    },
    setItem(key: string, value: string) {
      values.set(key, value);
    },
  };
}

describe("admin session persistence", () => {
  it("loads null when storage is empty", () => {
    expect(loadStoredAdminSession(createStorage())).toBeNull();
  });

  it("saves and loads a serialized session", () => {
    const storage = createStorage();
    const session: StoredAdminSession = {
      accessToken: "access_01",
      expiresAt: "2026-04-17T16:00:00Z",
      refreshToken: "refresh_01",
      user: {
        displayName: "Administrator",
        id: "user_admin",
        roles: ["role_admin"],
        username: "admin",
      },
    };

    saveStoredAdminSession(storage, session);

    expect(loadStoredAdminSession(storage)).toEqual(session);
  });

  it("clears broken payloads and invalid sessions", () => {
    const storage = createStorage();
    storage.setItem("bifrost.admin.session", "{bad json");
    expect(loadStoredAdminSession(storage)).toBeNull();
    expect(storage.getItem("bifrost.admin.session")).toBeNull();

    storage.setItem(
      "bifrost.admin.session",
      JSON.stringify({
        accessToken: "access_01",
      }),
    );
    expect(loadStoredAdminSession(storage)).toBeNull();
    expect(storage.getItem("bifrost.admin.session")).toBeNull();
  });

  it("removes a stored session", () => {
    const storage = createStorage();
    saveStoredAdminSession(storage, {
      accessToken: "access_01",
      expiresAt: "2026-04-17T16:00:00Z",
      refreshToken: "refresh_01",
      user: {
        displayName: "Administrator",
        id: "user_admin",
        roles: ["role_admin"],
        username: "admin",
      },
    });

    clearStoredAdminSession(storage);

    expect(loadStoredAdminSession(storage)).toBeNull();
  });
});
