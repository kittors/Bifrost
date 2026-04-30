import assert from "node:assert/strict";
import { existsSync, readFileSync, readdirSync, statSync } from "node:fs";
import { relative, resolve } from "node:path";
import test from "node:test";

const repositoryRoot = resolve(new URL("../..", import.meta.url).pathname);
const adminSourceRoot = resolve(repositoryRoot, "apps/admin/src");

function collectSourceFiles(directory) {
  const entries = readdirSync(directory);
  const files = [];

  for (const entry of entries) {
    const path = resolve(directory, entry);
    const stat = statSync(path);

    if (stat.isDirectory()) {
      files.push(...collectSourceFiles(path));
      continue;
    }

    if ((path.endsWith(".ts") || path.endsWith(".tsx")) && !path.endsWith(".test.tsx")) {
      files.push(path);
    }
  }

  return files;
}

function readSource(path) {
  return readFileSync(path, "utf8");
}

function relativeSourcePath(path) {
  return relative(repositoryRoot, path);
}

const adminSourceFiles = collectSourceFiles(adminSourceRoot);
const adminSource = adminSourceFiles.map(readSource).join("\n");

test("admin data tables consume HeroUI Table directly", () => {
  assert.equal(
    existsSync(resolve(adminSourceRoot, "shared/ui/table.tsx")),
    false,
    "后台不应保留 shared/ui/table.tsx 这种 HeroUI Table 二次封装",
  );
  assert.doesNotMatch(adminSource, /shared\/ui\/table/, "后台源码不应继续引用 shared/ui/table");

  const tableFiles = [
    "features/admin-devices/devices-table.tsx",
    "features/admin-roles/roles-table.tsx",
    "features/admin-services/services-table.tsx",
    "features/admin-users/users-table.tsx",
    "pages/audit-events-page.tsx",
  ];

  for (const file of tableFiles) {
    const source = readSource(resolve(adminSourceRoot, file));

    assert.match(
      source,
      /import\s+\{[^}]*\bTable\b[^}]*\}\s+from\s+"@heroui\/react"/,
      `${file} 应直接从 @heroui/react 引入 Table`,
    );
    assert.doesNotMatch(
      source,
      /overflow-hidden\s+rounded-\[14px\]\s+border\s+border-border\s+bg-surface/,
      `${file} 的表格外层不应继续使用整块边框容器`,
    );
  }
});

test("admin pagination consumes HeroUI Pagination directly", () => {
  assert.equal(
    existsSync(resolve(adminSourceRoot, "shared/ui/pagination-bar.tsx")),
    false,
    "后台不应保留 shared/ui/pagination-bar.tsx 这种 HeroUI Pagination 二次封装",
  );
  assert.doesNotMatch(
    adminSource,
    /shared\/ui\/pagination-bar/,
    "后台源码不应继续引用 shared/ui/pagination-bar",
  );

  const paginatedPages = [
    "pages/audit-events-page.tsx",
    "pages/devices-page.tsx",
    "pages/roles-page.tsx",
    "pages/services-page.tsx",
    "pages/users-page.tsx",
  ];

  for (const file of paginatedPages) {
    const source = readSource(resolve(adminSourceRoot, file));

    assert.match(
      source,
      /import\s+\{[^}]*\bPagination\b[^}]*\}\s+from\s+"@heroui\/react"/,
      `${file} 应直接从 @heroui/react 引入 Pagination`,
    );
  }
});

test("admin source uses HeroUI controls instead of native form controls", () => {
  const offenders = adminSourceFiles
    .map((path) => [path, readSource(path)])
    .filter(([, source]) => /<(input|select|textarea)\b/.test(source))
    .map(([path]) => relativeSourcePath(path));

  assert.deepEqual(
    offenders,
    [],
    "后台源码里不应直接使用原生 input/select/textarea；HeroUI 已提供对应组件",
  );
});
