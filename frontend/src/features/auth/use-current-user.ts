import { useQuery } from '@tanstack/react-query'
import { authGetCurrentUser } from '@/api'
import type { User } from '@/api'

export const currentUserQueryKey = ['auth', 'me'] as const

/**
 * Loads the authenticated user from GET /auth/me.
 * A 401 rejects the promise (retry is disabled globally) → treat as logged out.
 */
export function useCurrentUser() {
  return useQuery<User>({
    queryKey: currentUserQueryKey,
    queryFn: async () => {
      const { data } = await authGetCurrentUser({ throwOnError: true })
      return data
    },
  })
}
