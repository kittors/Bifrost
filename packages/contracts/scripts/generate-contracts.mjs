import { execFileSync } from "node:child_process";
import { mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

import YAML from "yaml";

const __dirname = dirname(fileURLToPath(import.meta.url));
const packageRoot = dirname(__dirname);
const openapiPath = join(packageRoot, "openapi", "bifrost.v1.yaml");
const tsOutputPath = join(packageRoot, "src", "generated.ts");
const goOutputPath = join(
  packageRoot,
  "..",
  "..",
  "apps",
  "gateway",
  "internal",
  "contracts",
  "generated.go",
);
const checkOnly = process.argv.includes("--check");

const openapiDocument = YAML.parse(readFileSync(openapiPath, "utf8"));
const schemas = openapiDocument.components?.schemas ?? {};

const errorCodes = schemas.ErrorCode?.enum;
const auditEventTypes = schemas.AuditEventType?.enum;

if (!Array.isArray(errorCodes) || errorCodes.length === 0) {
  throw new Error("components.schemas.ErrorCode.enum is required");
}

if (!Array.isArray(auditEventTypes) || auditEventTypes.length === 0) {
  throw new Error("components.schemas.AuditEventType.enum is required");
}

const formatStringArray = (values) => `[
${values.map((value) => `  ${JSON.stringify(value)},`).join("\n")}
]`;

const tsContent = `export const ERROR_CODES = ${formatStringArray(errorCodes)} as const;

export type ErrorCode = (typeof ERROR_CODES)[number];

export const AUDIT_EVENT_TYPES = ${formatStringArray(auditEventTypes)} as const;

export type AuditEventType = (typeof AUDIT_EVENT_TYPES)[number];
`;

const toPascalCase = (value) =>
  value
    .split(/[^a-zA-Z0-9]+/)
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1).toLowerCase())
    .join("");

const goErrorConsts = errorCodes
  .map((value) => `\tErrorCode${toPascalCase(value)} ErrorCode = "${value}"`)
  .join("\n");
const goAuditConsts = auditEventTypes
  .map((value) => `\tAuditEventType${toPascalCase(value)} AuditEventType = "${value}"`)
  .join("\n");

const goContent = `package contracts

type ErrorCode string

const (
${goErrorConsts}
)

type AuditEventType string

const (
${goAuditConsts}
)

type Pagination struct {
\tPage       int64 \`json:"page"\`
\tPageSize   int64 \`json:"pageSize"\`
\tTotal      int64 \`json:"total"\`
\tTotalPages int64 \`json:"totalPages"\`
}

type Meta struct {
\tRequestID  string      \`json:"requestId"\`
\tTimestamp  string      \`json:"timestamp"\`
\tPagination *Pagination \`json:"pagination,omitempty"\`
}

type APIError struct {
\tCode        ErrorCode          \`json:"code"\`
\tMessage     string             \`json:"message"\`
\tUserMessage string             \`json:"userMessage"\`
\tDetails     map[string]any     \`json:"details"\`
}
`;

function formatGoSource(source) {
  try {
    return execFileSync("gofmt", { encoding: "utf8", input: source });
  } catch {
    return source;
  }
}

function writeGeneratedFile(outputPath, nextContent) {
  const currentContent = (() => {
    try {
      return readFileSync(outputPath, "utf8");
    } catch {
      return null;
    }
  })();

  if (currentContent === nextContent) {
    return;
  }

  if (checkOnly) {
    throw new Error(`Generated contracts are stale: ${outputPath}`);
  }

  mkdirSync(dirname(outputPath), { recursive: true });
  writeFileSync(outputPath, nextContent);
}

writeGeneratedFile(tsOutputPath, tsContent);
writeGeneratedFile(goOutputPath, formatGoSource(goContent));
