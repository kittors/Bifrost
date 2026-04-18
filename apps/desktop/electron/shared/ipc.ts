export const desktopIPC = {
  appInfo: "bifrost:app-info",
  deviceAttach: "bifrost:device:attach",
  deviceClear: "bifrost:device:clear",
  deviceEnsure: "bifrost:device:ensure",
  deviceLoad: "bifrost:device:load",
  deviceSignChallenge: "bifrost:device:sign-challenge",
  diagnosticsSnapshot: "bifrost:diagnostics:snapshot",
  openExternal: "bifrost:open-external",
  sessionClear: "bifrost:session:clear",
  sessionLoad: "bifrost:session:load",
  sessionSave: "bifrost:session:save",
} as const;
