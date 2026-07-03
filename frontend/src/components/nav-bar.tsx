import { Link, useNavigate, useRouter } from '@tanstack/react-router'
import { ArrowLeftIcon, LogOutIcon } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { authLogout } from '@/api'
import { queryClient } from '@/lib/query-client'
import { useCurrentUser } from '@/features/auth/use-current-user'

export function NavBar() {
  const router = useRouter()
  const navigate = useNavigate()
  const { data: user } = useCurrentUser()

  async function handleLogout() {
    try {
      await authLogout()
    } catch {
      // Ignore network/expired-session errors — we clear client state regardless.
    }
    queryClient.clear()
    await navigate({ to: '/login' })
  }

  return (
    <header className="bg-background/95 supports-[backdrop-filter]:bg-background/60 sticky top-0 z-10 border-b backdrop-blur">
      <nav className="mx-auto flex h-14 w-full max-w-5xl items-center gap-3 px-4">
        <Button
          variant="ghost"
          size="icon"
          aria-label="Go back"
          onClick={() => router.history.back()}
        >
          <ArrowLeftIcon className="size-4" />
        </Button>

        <Link to="/" className="font-semibold tracking-tight">
          Project Works
        </Link>

        <div className="ml-auto flex items-center gap-3">
          {user && (
            <span className="text-muted-foreground hidden text-sm sm:inline">
              {user.display_name || user.email}
            </span>
          )}
          <Button variant="ghost" size="sm" onClick={handleLogout}>
            <LogOutIcon className="size-4" />
            Logout
          </Button>
        </div>
      </nav>
    </header>
  )
}
