import type { QueryClient } from "@tanstack/react-query";
import {
  createRootRouteWithContext,
  createRoute,
  createRouter,
  Outlet,
  redirect,
} from "@tanstack/react-router";

import { getCurrentAdminSession } from "../features/auth/store";
import { DashboardPage } from "../pages/dashboard-page";
import { LoginPage } from "../pages/login-page";
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

const routeTree = rootRoute.addChildren([
  loginRoute,
  appRoute.addChildren([dashboardRoute, usersRoute]),
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
