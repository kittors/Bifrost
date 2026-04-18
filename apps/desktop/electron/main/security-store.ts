import { mkdir, readFile, writeFile } from "node:fs/promises";
import { dirname, join } from "node:path";
import { app, safeStorage } from "electron";

import type { DesktopDeviceIdentity, DesktopSessionSnapshot } from "../shared/types";

type SecureStoreShape = {
  device?: string;
  session?: string;
};

const storePath = () => join(app.getPath("userData"), "secure-store.json");

async function readStore(): Promise<SecureStoreShape> {
  try {
    return JSON.parse(await readFile(storePath(), "utf8")) as SecureStoreShape;
  } catch {
    return {};
  }
}

async function writeStore(payload: SecureStoreShape) {
  await mkdir(dirname(storePath()), { recursive: true });
  await writeFile(storePath(), JSON.stringify(payload, null, 2), "utf8");
}

function encrypt(value: unknown) {
  if (!safeStorage.isEncryptionAvailable()) {
    throw new Error("system secure storage is unavailable");
  }
  return safeStorage.encryptString(JSON.stringify(value)).toString("base64");
}

function decrypt<T>(value?: string): T | null {
  if (!value || !safeStorage.isEncryptionAvailable()) {
    return null;
  }
  return JSON.parse(safeStorage.decryptString(Buffer.from(value, "base64"))) as T;
}

// 安全存储只保存加密 blob，Renderer 无法直接读取该文件中的敏感明文。
export async function loadSessionSnapshot() {
  return decrypt<DesktopSessionSnapshot>((await readStore()).session);
}

export async function saveSessionSnapshot(session: DesktopSessionSnapshot) {
  const store = await readStore();
  store.session = encrypt(session);
  await writeStore(store);
}

export async function clearSessionSnapshot() {
  const store = await readStore();
  delete store.session;
  await writeStore(store);
}

export async function loadDeviceIdentity() {
  return decrypt<DesktopDeviceIdentity>((await readStore()).device);
}

export async function saveDeviceIdentity(device: DesktopDeviceIdentity) {
  const store = await readStore();
  store.device = encrypt(device);
  await writeStore(store);
}

export async function clearDeviceIdentity() {
  const store = await readStore();
  delete store.device;
  await writeStore(store);
}

export function encryptionAvailable() {
  return safeStorage.isEncryptionAvailable();
}
