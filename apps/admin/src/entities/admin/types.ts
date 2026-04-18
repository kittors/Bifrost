export type AdminUser = {
  displayName: string;
  email: string;
  id: string;
  roles: string[];
  status: string;
  username: string;
};

export type AdminRole = {
  description: string;
  displayName: string;
  id: string;
  name: string;
};

export type AdminService = {
  description: string;
  group: string;
  id: string;
  key: string;
  name: string;
  protocol: string;
  publicPath: string;
  status: string;
  upstreamUrl: string;
};

export type AdminDevice = {
  clientVersion: string;
  id: string;
  name: string;
  os: string;
  publicKeyFingerprint: string;
  status: string;
  userId: string;
  userUsername: string;
};

export type AdminAuditEvent = {
  actorUserId: string;
  id: string;
  requestId: string;
  result: string;
  serviceId: string;
  summary: string;
  targetId: string;
  targetType: string;
  type: string;
};

export type UserServiceOverride = {
  effect: "allow" | "deny";
  serviceId: string;
};

export type PaginatedResult<T> = {
  items: T[];
  total: number;
};
