import { QueryClient } from '@tanstack/react-query'

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      // A 401 from /auth/me is a normal "logged out" signal, not a transient
      // failure — don't hammer the server retrying it.
      retry: false,
      refetchOnWindowFocus: false,
      staleTime: 30_000,
    },
  },
})
