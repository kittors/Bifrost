export type DesktopUserSummary = {
  displayName: string;
  id: string;
  roles: string[];
  username: string;
};

export type DesktopDeviceIdentity = {
  deviceId?: string;
  fingerprint: string;
  privateKeyPkcs8: string;
  publicKey: string;
};

export type DesktopSessionSnapshot = {
  accessToken: string;
  deviceId: string;
  expiresAt: string;
  gatewayBaseURL: string;
  refreshToken: string;
  user: DesktopUserSummary;
};

export type DesktopAppInfo = {
  name: string;
  platform: NodeJS.Platform;
  version: string;
};

export type DesktopDiagnosticsSnapshot = {
  encryptionAvailable: boolean;
  nodeIntegration: false;
  platform: NodeJS.Platform;
  proxyManagedByBifrost: false;
  routeManagedByBifrost: false;
  dnsManagedByBifrost: false;
};
