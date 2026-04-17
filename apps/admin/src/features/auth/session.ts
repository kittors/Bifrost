export type StoredAdminSession = {
  accessToken: string;
  expiresAt: string;
  refreshToken: string;
  user: {
    displayName: string;
    id: string;
    roles: string[];
    username: string;
  };
};

type StorageLike = Pick<Storage, "getItem" | "removeItem" | "setItem">;

const adminSessionStorageKey = "bifrost.admin.session";

function isValidStoredAdminSession(value: unknown): value is StoredAdminSession {
  if (!value || typeof value !== "object") {
    return false;
  }

  const candidate = value as Record<string, unknown>;
  const user = candidate.user as Record<string, unknown> | undefined;

  return (
    typeof candidate.accessToken === "string" &&
    typeof candidate.refreshToken === "string" &&
    typeof candidate.expiresAt === "string" &&
    !!user &&
    typeof user.id === "string" &&
    typeof user.username === "string" &&
    typeof user.displayName === "string" &&
    Array.isArray(user.roles)
  );
}

export function loadStoredAdminSession(storage: StorageLike) {
  const raw = storage.getItem(adminSessionStorageKey);

  if (!raw) {
    return null;
  }

  try {
    const parsed = JSON.parse(raw);
    if (isValidStoredAdminSession(parsed)) {
      return parsed;
    }
  } catch {
    // Invalid payloads are cleared below.
  }

  storage.removeItem(adminSessionStorageKey);
  return null;
}

export function saveStoredAdminSession(storage: StorageLike, session: StoredAdminSession) {
  storage.setItem(adminSessionStorageKey, JSON.stringify(session));
}

export function clearStoredAdminSession(storage: StorageLike) {
  storage.removeItem(adminSessionStorageKey);
}

export function browserStorage() {
  if (typeof window === "undefined") {
    return null;
  }

  return window.localStorage;
}
