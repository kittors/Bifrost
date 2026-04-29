import { generateKeyPairSync, sign } from "node:crypto";

import type { DesktopDeviceIdentity } from "../shared/types";
import { loadDeviceIdentity, saveDeviceIdentity } from "./security-store";

function encodeBase64URL(buffer: Buffer) {
  return buffer.toString("base64").replaceAll("+", "-").replaceAll("/", "_").replaceAll("=", "");
}

function generateDeviceIdentity(): DesktopDeviceIdentity {
  const { privateKey, publicKey } = generateKeyPairSync("ed25519");
  const publicKeyDER = publicKey.export({ format: "der", type: "spki" });
  const privateKeyPKCS8 = privateKey.export({ format: "der", type: "pkcs8" });
  const publicKeyRaw = Buffer.from(publicKeyDER).subarray(-32);
  const encodedPublicKey = encodeBase64URL(publicKeyRaw);

  return {
    fingerprint: `fp_${encodedPublicKey.slice(0, 16)}`,
    privateKeyPkcs8: Buffer.from(privateKeyPKCS8).toString("base64"),
    publicKey: encodedPublicKey,
  };
}

// 设备密钥只在 Main 进程生成并保存，Renderer 只能拿到公钥和指纹摘要。
export async function ensureDeviceIdentity() {
  const existing = await loadDeviceIdentity();
  if (existing) {
    return existing;
  }

  const identity = generateDeviceIdentity();
  await saveDeviceIdentity(identity);
  return identity;
}

export async function attachDeviceID(deviceID: string) {
  const identity = await ensureDeviceIdentity();
  const next = { ...identity, deviceId: deviceID };
  await saveDeviceIdentity(next);
  return next;
}

export async function signDeviceChallenge(challenge: string) {
  const identity = await loadDeviceIdentity();
  if (!identity) {
    throw new Error("device identity is missing");
  }

  const signature = sign(null, Buffer.from(challenge, "base64url"), {
    key: Buffer.from(identity.privateKeyPkcs8, "base64"),
    format: "der",
    type: "pkcs8",
  });

  return encodeBase64URL(signature);
}
