import { QueryClient, QueryClientProvider } from "@tanstack/react-query";

import { DesktopApp } from "./app";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 1,
      staleTime: 15_000,
    },
  },
});

export function AppProviders() {
  return (
    <QueryClientProvider client={queryClient}>
      <DesktopApp />
    </QueryClientProvider>
  );
}
