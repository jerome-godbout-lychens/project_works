# Project Works — Frontend

TypeScript + React single-page app for Project Works. This first pass ships the
landing **dashboard** (project list with a live filter) and an **SSO login** page.

## Stack

- **Vite** + **React 19** + **TypeScript**
- **TanStack Router** (file-based routes) + **TanStack Query** (server state)
- **Tailwind CSS v4** + **shadcn/ui** components
- Typed API client generated from `../backend/openapi.yaml` via **@hey-api/openapi-ts**

## Getting started

```bash
npm install
npm run codegen   # generate src/api from ../backend/openapi.yaml
npm run dev       # http://localhost:5173 (or next free port)
```

The dev server proxies `/api/v1` to the backend on `http://localhost:8088`, so the
browser stays same-origin (no CORS) and the `session` cookie flows normally.

### Local auth during development

The backend's SSO (OIDC) is unconfigured by default, and the seeded super-admin
authenticates with a Bearer API key. For local development, copy `.env.example`
to `.env.development` and set:

```
VITE_DEV_API_KEY=super-admin-dev-key-do-not-use-in-production
```

The API client injects this key **only** in dev builds (`import.meta.env.DEV`).
Without it, `/auth/me` returns 401 and you get the SSO login screen.

## Scripts

| Script | Purpose |
| --- | --- |
| `npm run dev` | Start the Vite dev server |
| `npm run build` | Generate routes, type-check, and build for production |
| `npm run preview` | Preview the production build |
| `npm run codegen` | Regenerate the typed API client from the backend spec |
| `npm run routes:gen` | Regenerate the TanStack Router route tree |

## SSO configuration

`public/config.json` decides which SSO providers are offered and the allowed
email domain — no rebuild required:

```json
{
  "sso": {
    "providers": [
      { "id": "microsoft", "label": "Sign in with Microsoft", "enabled": true },
      { "id": "google", "label": "Sign in with Google", "enabled": false },
      { "id": "apple", "label": "Sign in with Apple", "enabled": false }
    ],
    "allowedDomain": "lichens.ai",
    "enterpriseRestriction": true
  }
}
```

Each enabled provider links to `GET /api/v1/auth/login`. The backend currently
has a single OIDC issuer, so `provider`/`domain_hint` are forward-compatible
hints; multi-provider wiring and domain enforcement live on the backend/IdP side.

## Structure

```
src/
├── api/            # generated typed client (do not edit)
├── components/     # shared UI (nav-bar) + ui/ (shadcn primitives)
├── features/
│   ├── auth/       # login page, current-user hook, route guard
│   └── projects/   # project list, card, query hook
├── lib/            # api-client config, query client, runtime config, cn()
└── routes/         # file-based routes (login, _authed layout, dashboard, stub)
```

Clicking a project navigates to `/projects/$projectId`, currently a placeholder —
the detail view is intentionally not implemented yet.
