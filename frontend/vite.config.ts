import { defineConfig } from 'vite'
import { fileURLToPath, URL } from 'node:url'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import { tanstackRouter } from '@tanstack/router-plugin/vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    // Must come before the React plugin. Generates src/routeTree.gen.ts from src/routes.
    tanstackRouter({ target: 'react' }),
    react(),
    tailwindcss(),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    port: 5173,
    // Proxy the backend so the browser stays same-origin: no CORS, and the
    // HttpOnly SameSite=Lax `session` cookie flows normally. No path rewrite —
    // the /api/v1 prefix must survive to match the live Go handlers.
    proxy: {
      '/api/v1': {
        target: 'http://localhost:8088',
        changeOrigin: true,
      },
    },
  },
})
