import { defineConfig } from '@hey-api/openapi-ts'

// Generates the typed API client into src/api from the backend's committed spec.
// NOTE: the spec's `servers` URL is absolute/cross-origin and its paths omit the
// /api/v1 prefix. We deliberately ignore `servers` and set baseUrl: '/api/v1' at
// runtime (see src/lib/api-client.ts) so requests are relative and flow through
// the Vite dev proxy.
export default defineConfig({
  input: '../backend/openapi.yaml',
  output: {
    path: './src/api',
  },
  plugins: ['@hey-api/client-fetch', '@hey-api/sdk', '@hey-api/typescript'],
})
