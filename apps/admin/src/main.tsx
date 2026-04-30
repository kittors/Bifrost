import "@bifrost/design-tokens/app.css";

import { ToastProvider } from "@heroui/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { RouterProvider } from "@tanstack/react-router";
import React from "react";
import ReactDOM from "react-dom/client";

import { router } from "./app/router";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      staleTime: 30_000,
    },
  },
});

const rootElement = document.getElementById("app");

if (!rootElement) {
  throw new Error("Bifrost admin root element was not found.");
}

ReactDOM.createRoot(rootElement).render(
  <React.StrictMode>
    <QueryClientProvider client={queryClient}>
      <RouterProvider context={{ queryClient }} router={router} />
      <ToastProvider placement="top end" />
    </QueryClientProvider>
  </React.StrictMode>,
);
