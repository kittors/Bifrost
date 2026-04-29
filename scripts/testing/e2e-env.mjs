export const e2ePorts = {
  admin: process.env.BIFROST_DEV_ADMIN_PORT?.trim() || "15173",
  gateway: process.env.BIFROST_DEV_GATEWAY_PORT?.trim() || "18080",
  postgres: process.env.BIFROST_DEV_POSTGRES_PORT?.trim() || "15432",
};

const e2ePortEnvironmentKeys = [
  "BIFROST_DEV_ADMIN_PORT",
  "BIFROST_DEV_GATEWAY_PORT",
  "BIFROST_DEV_POSTGRES_PORT",
];

export function e2eEnvironment(overrides = {}) {
  return {
    ...process.env,
    BIFROST_DEV_ADMIN_PORT: e2ePorts.admin,
    BIFROST_DEV_GATEWAY_PORT: e2ePorts.gateway,
    BIFROST_DEV_POSTGRES_PORT: e2ePorts.postgres,
    ...overrides,
  };
}

export function withoutE2EPortOverrides(environment = process.env) {
  const nextEnvironment = { ...environment };
  for (const key of e2ePortEnvironmentKeys) {
    delete nextEnvironment[key];
  }
  return nextEnvironment;
}
