import { create } from "zustand";

import {
  browserStorage,
  clearStoredAdminSession,
  loadStoredAdminSession,
  type StoredAdminSession,
  saveStoredAdminSession,
} from "./session";

type AdminSessionState = {
  clearSession: () => void;
  session: StoredAdminSession | null;
  setSession: (session: StoredAdminSession) => void;
};

function initialSession() {
  const storage = browserStorage();

  if (!storage) {
    return null;
  }

  return loadStoredAdminSession(storage);
}

export const useAdminSessionStore = create<AdminSessionState>((set) => ({
  clearSession: () => {
    const storage = browserStorage();
    if (storage) {
      clearStoredAdminSession(storage);
    }
    set({ session: null });
  },
  session: initialSession(),
  setSession: (session) => {
    const storage = browserStorage();
    if (storage) {
      saveStoredAdminSession(storage, session);
    }
    set({ session });
  },
}));

export function getCurrentAdminSession() {
  return useAdminSessionStore.getState().session;
}
