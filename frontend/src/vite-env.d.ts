/// <reference types="vite/client" />

interface ImportMetaEnv {
  /** Dev-only Bearer API key injected by the API client when import.meta.env.DEV. */
  readonly VITE_DEV_API_KEY?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
