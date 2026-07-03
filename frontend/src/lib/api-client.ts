import { client } from '@/api/client.gen'

/**
 * Configures the generated API client. Import this module for its side effects
 * (from main.tsx) before any query runs.
 *
 * - baseUrl '/api/v1' OVERRIDES the spec's absolute `http://localhost:8088/api/v1`
 *   server URL so requests are relative and same-origin. In dev they flow through
 *   the Vite proxy (see vite.config.ts) to the backend; no CORS, and the HttpOnly
 *   `session` cookie is sent/received automatically.
 * - credentials 'include' attaches that cookie on every request.
 * - In DEV only, when VITE_DEV_API_KEY is set, we add a Bearer header so the app
 *   can authenticate against a backend whose OIDC is unconfigured (the seeded
 *   super-admin authenticates via API key). This never runs in a production build.
 */
const devApiKey = import.meta.env.DEV ? import.meta.env.VITE_DEV_API_KEY : undefined

client.setConfig({
  baseUrl: '/api/v1',
  credentials: 'include',
  ...(devApiKey
    ? { headers: { Authorization: `Bearer ${devApiKey}` } }
    : {}),
})
