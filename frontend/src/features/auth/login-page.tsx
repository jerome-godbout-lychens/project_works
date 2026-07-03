import { useQuery } from '@tanstack/react-query'
import { Navigate } from '@tanstack/react-router'

import { Button } from '@/components/ui/button'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { loadAppConfig, type SsoProvider } from '@/lib/config'
import { useCurrentUser } from './use-current-user'

/**
 * Builds the backend login URL for a provider. The backend has a single OIDC
 * issuer today and ignores these query hints, so `provider` / `domain_hint` are
 * forward-compatible only. This is a full-page navigation (not fetch) because
 * login is a 302 redirect chain to the identity provider that sets cookies.
 */
function loginUrl(providerId: string, allowedDomain?: string): string {
  const params = new URLSearchParams({ provider: providerId })
  if (allowedDomain) params.set('domain_hint', allowedDomain)
  return `/api/v1/auth/login?${params.toString()}`
}

export function LoginPage() {
  // If already authenticated (e.g. dev API key), skip the login screen.
  const { data: user } = useCurrentUser()
  const { data: config, isLoading } = useQuery({
    queryKey: ['app-config'],
    queryFn: loadAppConfig,
    staleTime: Infinity,
  })

  if (user) return <Navigate to="/" />

  const providers: SsoProvider[] =
    config?.sso.providers.filter((p) => p.enabled) ?? []
  const allowedDomain = config?.sso.allowedDomain
  const restricted = config?.sso.enterpriseRestriction && allowedDomain

  return (
    <div className="flex min-h-svh items-center justify-center bg-muted/30 p-4">
      <Card className="w-full max-w-sm">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl">Project Works</CardTitle>
          <CardDescription>
            {restricted
              ? `Sign in with your @${allowedDomain} account`
              : 'Sign in to continue'}
          </CardDescription>
        </CardHeader>
        <CardContent className="flex flex-col gap-3">
          {isLoading && (
            <p className="text-muted-foreground text-center text-sm">
              Loading sign-in options…
            </p>
          )}

          {!isLoading && providers.length === 0 && (
            <p className="text-muted-foreground text-center text-sm">
              No sign-in providers are enabled. Configure them in{' '}
              <code>config.json</code>.
            </p>
          )}

          {providers.map((provider) => (
            <Button key={provider.id} asChild size="lg" className="w-full">
              <a href={loginUrl(provider.id, allowedDomain)}>{provider.label}</a>
            </Button>
          ))}
        </CardContent>
      </Card>
    </div>
  )
}
