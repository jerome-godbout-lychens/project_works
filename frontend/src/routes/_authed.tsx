import { createFileRoute, Navigate, Outlet } from '@tanstack/react-router'

import { NavBar } from '@/components/nav-bar'
import { useCurrentUser } from '@/features/auth/use-current-user'

export const Route = createFileRoute('/_authed')({
  component: AuthedLayout,
})

/**
 * Pathless layout that gates every authenticated route: while /auth/me loads we
 * show a placeholder; a 401 (or any failure) redirects to /login; otherwise the
 * nav bar + page render.
 */
function AuthedLayout() {
  const { isLoading, isError, data: user } = useCurrentUser()

  if (isLoading) {
    return (
      <div className="flex min-h-svh items-center justify-center">
        <p className="text-muted-foreground text-sm">Loading…</p>
      </div>
    )
  }

  if (isError || !user) {
    return <Navigate to="/login" />
  }

  return (
    <div className="min-h-svh">
      <NavBar />
      <main>
        <Outlet />
      </main>
    </div>
  )
}
