import type { QueryClient } from "@tanstack/react-query";
import {
  createRootRouteWithContext,
  createRoute,
  createRouter,
  Outlet,
  redirect,
} from "@tanstack/react-router";

import { getCurrentAdminSession } from "../features/auth/store";
import { AuditEventsPage } from "../pages/audit-events-page";
import { DashboardPage } from "../pages/dashboard-page";
import { DevicesPage } from "../pages/devices-page";
import { LoginPage } from "../pages/login-page";
import { RolesPage } from "../pages/roles-page";
import { ServicesPage } from "../pages/services-page";
import { UsersPage } from "../pages/users-page";
import { AdminShell } from "./layout/admin-shell";

type RouterContext = {
  queryClient: QueryClient;
};

const rootRoute = createRootRouteWithContext<RouterContext>()({
  component: () => <Outlet />,
});

const loginRoute = createRoute({
  beforeLoad: () => {
    if (getCurrentAdminSession()) {
      throw redirect({ to: "/" });
    }
  },
  getParentRoute: () => rootRoute,
  path: "/login",
  component: LoginPage,
});

const appRoute = createRoute({
  beforeLoad: () => {
    if (!getCurrentAdminSession()) {
      throw redirect({ to: "/login" });
    }
  },
  getParentRoute: () => rootRoute,
  id: "admin-app",
  component: AdminShell,
});

const dashboardRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/",
  component: DashboardPage,
});

const usersRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/users",
  component: UsersPage,
});

const rolesRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/roles",
  component: RolesPage,
});

const devicesRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/devices",
  component: DevicesPage,
});

const servicesRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/services",
  component: ServicesPage,
});

const auditEventsRoute = createRoute({
  getParentRoute: () => appRoute,
  path: "/audit-events",
  component: AuditEventsPage,
});

const routeTree = rootRoute.addChildren([
  loginRoute,
  appRoute.addChildren([
    dashboardRoute,
    usersRoute,
    rolesRoute,
    devicesRoute,
    servicesRoute,
    auditEventsRoute,
  ]),
]);

export const router = createRouter({
  context: {
    queryClient: undefined as unknown as QueryClient,
  },
  defaultPreload: "intent",
  routeTree,
});

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router;
  }
}
