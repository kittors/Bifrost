import { beforeEach, describe, expect, it } from "vitest";

import { adminThemeStorageKey, applyAdminTheme, readAdminTheme } from "./admin-shell";

describe("AdminShell", () => {
  beforeEach(() => {
    window.localStorage.clear();
    document.documentElement.removeAttribute("data-theme");
  });

  it("applies and persists the selected theme", () => {
    expect(readAdminTheme()).toBe("light");

    applyAdminTheme("dark");

    expect(document.documentElement.getAttribute("data-theme")).toBe("dark");
    expect(window.localStorage.getItem(adminThemeStorageKey)).toBe("dark");
    expect(readAdminTheme()).toBe("dark");
  });
});
