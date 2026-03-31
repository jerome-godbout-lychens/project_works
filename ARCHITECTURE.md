# Architecture & Technology Design

## Technology Stack

### Backend — Go

| Concern | Choice | Rationale |
|---|---|---|
| Web framework | [Huma v2](https://huma.rocks) | OpenAPI-first — generates spec from Go code, rising fast, CLI/curl friendly out of the box |
| Router | Chi (via Huma adapter) | stdlib-compatible, minimal, composable middleware |
| Database (default) | PostgreSQL 16 | JSONB for custom fields, `tsvector` for full-text search, battle-tested |
| DB access layer | [sqlc](https://sqlc.dev) | Type-safe SQL → Go, no runtime reflection, swap-friendly via generated interfaces |
| DB migrations | [goose](https://github.com/pressly/goose) | Embedded SQL migrations, supports multiple DB dialects for future swap |
| File storage | AWS SDK S3 client ([aws-sdk-go-v2](https://github.com/aws/aws-sdk-go-v2)) | Single client works against SeaweedFS (local/self-hosted) and AWS S3 (prod) — endpoint is config-driven |
| Read cache | [ristretto](https://github.com/dgraph-io/ristretto) | In-process LRU; swap point exists for Redis/ElastiCache if multi-instance needed |
| Auth / OIDC | [coreos/go-oidc](https://github.com/coreos/go-oidc) + Office 365 | Standards-based OIDC, easy to swap provider |
| API key auth | Custom middleware (HMAC SHA-256) | Stored as hashed keys in DB, for CLI/automation access |
| Config | [viper](https://github.com/spf13/viper) | File + env var config, YAML/TOML support |

### Frontend — TypeScript + React

| Concern | Choice | Rationale |
|---|---|---|
| Build tool | Vite | Fast HMR, first-class TS |
| UI framework | React 19 | Largest ecosystem, hooks-first |
| Component lib | [shadcn/ui](https://ui.shadcn.com) + Radix UI | Headless, copy-owned components, no lock-in |
| Server state | [TanStack Query v5](https://tanstack.com/query) | Cache, background sync, optimistic updates |
| Tables | [TanStack Table v8](https://tanstack.com/table) | Headless, handles hierarchical/matrix views |
| Routing | [TanStack Router](https://tanstack.com/router) | Type-safe routes, file-based |
| Forms | React Hook Form + Zod | Lightweight, schema-validated |
| Diagrams | Mermaid.js + [Gantt: dhtmlx-gantt OSS](https://github.com/DHTMLX/gantt) | Mermaid in descriptions; dedicated Gantt for schedule view |
| PERT / DFP matrices | D3.js | Custom network graph and matrix renders |
| API client | [@hey-api/openapi-ts](https://github.com/hey-api/openapi-ts) | Auto-generated from Huma's OpenAPI spec |
| Markdown | [react-markdown](https://github.com/remarkjs/react-markdown) + remark-mermaid | Render descriptions with embedded diagrams |

---

## Repository Structure

```
/
├── backend/
│   ├── cmd/
│   │   └── server/          # main entrypoint
│   ├── internal/
│   │   ├── config/          # viper config loader
│   │   ├── domain/          # pure domain types & interfaces (no deps)
│   │   ├── store/           # DB interface + implementations
│   │   │   ├── interface.go # Store interface (swap point)
│   │   │   ├── postgres/    # PostgreSQL implementation via sqlc
│   │   │   └── sqlite/      # (future) SQLite implementation
│   │   ├── service/         # business logic, reads domain + store
│   │   ├── api/             # Huma handlers, OpenAPI wiring
│   │   ├── auth/            # OIDC + API key middleware
│   │   ├── filestore/
│   │   │   ├── interface.go # FileStore interface (Put, Get, Delete, URL)
│   │   │   └── s3/          # AWS SDK v2 impl — works with SeaweedFS + AWS S3
│   │   ├── cache/           # ristretto-backed read cache
│   │   └── plugin/          # plugin loader & registry
│   ├── migrations/          # goose SQL files
│   └── openapi.yaml         # generated, committed for client codegen
│
├── frontend/
│   ├── src/
│   │   ├── api/             # auto-generated from openapi.yaml
│   │   ├── components/      # shared UI components
│   │   ├── features/        # feature-scoped modules (elements, matrix, gantt…)
│   │   ├── plugins/         # frontend plugin registry
│   │   └── routes/          # TanStack Router file-based routes
│   └── vite.config.ts
│
├── plugins/                 # optional: community/custom plugins
├── docker-compose.yml       # postgres + seaweedfs + app for local dev
└── deploy/
    ├── terraform/           # AWS infra (RDS, S3, ECS/EKS, ALB)
    └── k8s/                 # optional Kubernetes manifests
```

---

## Architectural Principles

### 1. Layered / Clean Architecture (backend)

```
HTTP Handler (api/)
    ↓
Service (service/)          ← business rules, no HTTP/DB knowledge
    ↓
Store Interface (domain/)   ← pure Go interfaces
    ↓
Implementation (store/postgres/)
```

Swapping the database = providing a new `store/` implementation. No service code changes.

### 2. CQRS-lite for Read Performance

Write path: domain object → validate → write to PostgreSQL.

Read path: service → check ristretto cache → if miss, run read-optimized query → cache result.

Cache invalidation is **targeted**: modifying a Feature's child tasks invalidates only that Feature's aggregated view, not the entire cache.

### 3. Plugin System

**Backend plugins** are Go packages that implement a `Plugin` interface and are registered at startup. Active plugins are listed in config (`plugins.enabled: [gantt, pert, ...]`). This avoids dynamic `.so` loading complexity.

**Frontend plugins** are React modules lazy-loaded via dynamic `import()`. The plugin registry maps plugin IDs to their component/route contributions.

### 4. OpenAPI-First API

Huma generates the OpenAPI 3.1 spec from Go handler signatures. The spec is committed to `/backend/openapi.yaml` and `@hey-api/openapi-ts` generates the TypeScript client on every build. The CLI uses `curl` + API keys against the same spec.

### 5. Custom Fields

Stored as `JSONB` in PostgreSQL with a `custom_field_definitions` table per element type (name, type, validation). The domain layer merges them at read time. Full-text search indexes the JSONB values.

### 6. Authentication Flow

```
Browser → /auth/login → OIDC redirect (Office 365) → callback → JWT session cookie
CLI     → Authorization: Bearer <api-key>
```

Groups map to projects via a `group_project_access` join table. No per-user project ACL — only group-level.

---

## Data Model Overview

```
projects (folder-tree via ltree)
elements (polymorphic: type discriminator + type-specific table per element)
  ├── requirements
  ├── features
  ├── tasks
  ├── bugs
  └── evals
element_links (src_id, dest_id, link_type)
custom_field_definitions (element_type, name, field_type)
attachments (element_id → SeaweedFS fid)
users / groups / group_project_access
api_keys (hashed)
phases (for start_phase / delivery_phase on tasks)
```

PostgreSQL `ltree` extension handles the multi-depth project folder tree efficiently.

---

## Key Views / Matrices

| View | Implementation |
|---|---|
| Hierarchical element list | TanStack Table with row expansion, server-side pagination |
| DFP matrix | D3 custom matrix — rows=requirements, cols=features/tasks |
| Coupling matrix | D3 half-matrix — tasks × tasks |
| Gantt | dhtmlx-gantt OSS, fed from task `start_phase`/`delivery_phase` |
| PERT / critical path | D3 DAG layout, computed server-side from element links |
| Card stack | React context stack — card opens on top, back/forward navigation |

---

## Deployment Targets

All infrastructure is abstracted behind interfaces. Swapping a backend is a config change + connection string, no application code changes.

| Concern | Local / Self-hosted | AWS (prod) | Swap mechanism |
|---|---|---|---|
| Database | PostgreSQL (Docker) | AWS RDS PostgreSQL | Connection string in config |
| File storage | SeaweedFS (Docker) | AWS S3 | `filestore.endpoint`, `filestore.bucket` in config; same AWS SDK client |
| Read cache | ristretto (in-process) | AWS ElastiCache (Redis) | `cache.driver: redis` → swap ristretto for go-redis impl behind `Cache` interface |
| Compute | Local process | AWS ECS (Fargate) or EKS | Docker image, no code change |
| CDN / static | Vite dev server | CloudFront + S3 | Frontend build output |
| Secrets | Config file / env vars | AWS Secrets Manager | Viper supports env var injection; add SM loader for prod |

### Config example (environment-driven)

```yaml
# config.yaml (local dev)
database:
  driver: postgres
  dsn: "postgres://user:pass@localhost:5432/tracker"

filestore:
  driver: s3
  endpoint: "http://localhost:8333"   # SeaweedFS local
  bucket: attachments
  region: us-east-1                   # ignored by SeaweedFS but required by SDK

cache:
  driver: memory                      # ristretto in-process

server:
  domain: localhost
  port: 8080
```

```yaml
# config.yaml (AWS prod — or inject via env vars / Secrets Manager)
database:
  driver: postgres
  dsn: "${RDS_DSN}"                   # env var injected at runtime

filestore:
  driver: s3
  endpoint: ""                        # empty = native AWS S3
  bucket: my-tracker-attachments
  region: us-east-1

cache:
  driver: redis
  addr: "${ELASTICACHE_ADDR}"

server:
  domain: tracker.mycompany.com
  port: 443
```

---

## Local Dev Setup

```bash
docker compose up          # postgres + seaweedfs
cd backend && go run ./cmd/server
cd frontend && npm run dev
```

OpenAPI spec regenerated with: `cd backend && go generate ./...`
Frontend client regenerated with: `cd frontend && npm run codegen`
