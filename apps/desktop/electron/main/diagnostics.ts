import type { DesktopDiagnosticsSnapshot } from "../shared/types";
import { encryptionAvailable } from "./security-store";

// 诊断快照明确声明 Bifrost 没有接管代理、DNS 或系统路由。
export function getDiagnosticsSnapshot(): DesktopDiagnosticsSnapshot {
  return {
    dnsManagedByBifrost: false,
    encryptionAvailable: encryptionAvailable(),
    nodeIntegration: false,
    platform: process.platform,
    proxyManagedByBifrost: false,
    routeManagedByBifrost: false,
  };
}
