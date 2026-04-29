const userAgent = process.env.npm_config_user_agent ?? "";

if (userAgent.startsWith("pnpm/")) {
  process.exit(0);
}

const detectedPackageManager = userAgent.split(" ")[0] || "unknown package manager";

console.error(
  [
    "Please install dependencies with pnpm.",
    "",
    `Detected: ${detectedPackageManager}`,
    "Expected: pnpm@10.33.0",
    "",
    "Run:",
    "  corepack enable",
    "  corepack prepare pnpm@10.33.0 --activate",
    "  pnpm install --frozen-lockfile",
  ].join("\n"),
);

process.exit(1);
