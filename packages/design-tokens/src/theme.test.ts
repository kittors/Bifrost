import { readFileSync } from "node:fs";
import { join } from "node:path";

import { describe, expect, it } from "vitest";

const themeCss = readFileSync(join(import.meta.dirname, "theme.css"), "utf8");
const baseCss = readFileSync(join(import.meta.dirname, "base.css"), "utf8");

describe("design tokens", () => {
  it("defines light and dark semantic color variables", () => {
    expect(themeCss).toContain("--color-bg: var(--bifrost-bg);");
    expect(themeCss).toContain("--color-surface: var(--bifrost-surface);");
    expect(baseCss).toContain(":root {");
    expect(baseCss).toContain('[data-theme="dark"] {');
    expect(baseCss).toContain("--bifrost-bg:");
    expect(baseCss).toContain("--bifrost-brand:");
  });

  it("defines typography, radius, button, input, and layout tokens", () => {
    expect(themeCss).toContain("--text-1-size: 12px;");
    expect(themeCss).toContain("--text-4-size: 16px;");
    expect(themeCss).toContain("--bifrost-btn-md-height: 36px;");
    expect(themeCss).toContain("--bifrost-input-lg-height: 40px;");
    expect(themeCss).toContain("--bifrost-desktop-window-width: 420px;");
    expect(themeCss).toContain("--radius-lg: 14px;");
  });
});
