/**
 * Runtime frontend configuration, loaded from `public/config.json` at startup.
 *
 * This is the single place that decides which SSO providers are offered on the
 * login page and which email domain is allowed. It is a plain static file so it
 * can be changed by an operator without rebuilding the app.
 */

export type SsoProviderId = 'microsoft' | 'google' | 'apple' | (string & {})

export interface SsoProvider {
  id: SsoProviderId
  label: string
  enabled: boolean
}

export interface SsoConfig {
  providers: SsoProvider[]
  /** Allowed email domain, e.g. "lichens.ai" (presentational + optional IdP hint). */
  allowedDomain?: string
  /** When true, the login page states that only allowedDomain accounts may sign in. */
  enterpriseRestriction?: boolean
}

export interface AppConfig {
  sso: SsoConfig
}

const FALLBACK_CONFIG: AppConfig = {
  sso: {
    providers: [{ id: 'microsoft', label: 'Sign in with Microsoft', enabled: true }],
  },
}

let configPromise: Promise<AppConfig> | null = null

/** Fetches and caches the runtime config. Falls back to a safe default on error. */
export function loadAppConfig(): Promise<AppConfig> {
  if (!configPromise) {
    configPromise = fetch('/config.json', { cache: 'no-cache' })
      .then((res) => {
        if (!res.ok) throw new Error(`config.json ${res.status}`)
        return res.json() as Promise<AppConfig>
      })
      .catch((err) => {
        console.warn('Failed to load /config.json, using fallback config.', err)
        return FALLBACK_CONFIG
      })
  }
  return configPromise
}
