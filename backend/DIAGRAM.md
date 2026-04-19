# Backend Architecture Diagrams

## Package Structure

```mermaid
graph TD
    CMD["cmd/server\nmain.go"] --> API["internal/api\nHTTP handlers (Huma v2)"]
    CMD --> SERVICE["internal/service\nBusiness logic"]
    CMD --> STORE["internal/store/postgres\nPostgreSQL implementations"]
    CMD --> FILESTORE["internal/filestore/s3\nS3-compatible file storage"]
    CMD --> CACHE["internal/cache/memory\nRistretto in-process cache"]
    CMD --> AUTH["internal/auth\nOIDC / Session / API key"]
    CMD --> CONFIG["internal/config\nYAML configuration"]

    API --> SERVICE
    SERVICE --> DOMAIN["internal/domain\nInterfaces + domain types"]
    STORE --> DOMAIN
    FILESTORE --> DOMAIN
    CACHE --> DOMAIN
```

## Domain Interfaces (Dependency Inversion)

```mermaid
classDiagram
    class ElementStore {
        +GetElementById(ctx, elementId) Element
        +ListElementsByProject(ctx, projectId, filter) []Element
        +CreateElement(ctx, element) error
        +UpdateElement(ctx, element) error
        +DeleteElement(ctx, elementId) error
        +SearchElements(ctx, projectId, query, limit, offset) []Element
    }

    class ElementLinkStore {
        +CreateLink(ctx, link) error
        +GetLinkById(ctx, linkId) ElementLink
        +UpdateLinkType(ctx, linkId, linkType) error
        +DeleteLink(ctx, linkId) error
        +ListLinksByElement(ctx, elementId, direction) []ElementLink
    }

    class ElementVersionStore {
        +CreateVersion(ctx, version, patch) error
        +ListVersionsByElement(ctx, elementId, limit, offset) []ElementVersion
        +GetPatchesInRange(ctx, elementId, from, to) []ElementVersionPatch
    }

    class ElementPendingChangeStore {
        +UpsertPendingChange(ctx, elementId, snapshot) error
        +GetPendingChanges(ctx, batchSize) []ElementPendingChange
        +DeletePendingChange(ctx, elementId) error
    }

    class AttachmentStore {
        +CreateAttachment(ctx, attachment) error
        +GetAttachmentByIdentifier(ctx, attachmentIdentifier) Attachment
        +ListAttachmentsByElement(ctx, elementId) []Attachment
        +DeleteAttachment(ctx, attachmentId) error
    }

    class FileStorage {
        +Upload(ctx, key, reader, size, contentType) error
        +GetPresignedURL(ctx, key, expiration) string
        +Delete(ctx, key) error
    }

    class CacheStore {
        +Get(ctx, key) any
        +Set(ctx, key, value, ttl) error
        +Invalidate(ctx, key) error
    }

    PostgresElementStore ..|> ElementStore
    PostgresElementLinkStore ..|> ElementLinkStore
    PostgresElementVersionStore ..|> ElementVersionStore
    PostgresElementPendingChangeStore ..|> ElementPendingChangeStore
    PostgresAttachmentStore ..|> AttachmentStore
    S3FileStorage ..|> FileStorage
    RistrettoCache ..|> CacheStore
```

## Request Flow — Element Read (Cache Path)

```mermaid
sequenceDiagram
    participant Client
    participant API as api/element_handler
    participant SVC as service/ElementService
    participant Cache as cache/RistrettoCache
    participant DB as store/postgres/ElementStore

    Client->>API: GET /api/v1/projects/{id}/elements/{element_id}
    API->>SVC: GetElementById(ctx, elementId)
    SVC->>Cache: Get("element:{id}")
    alt Cache hit
        Cache-->>SVC: *domain.Element
    else Cache miss
        SVC->>DB: GetElementById(ctx, elementId)
        DB-->>SVC: *domain.Element
        SVC->>DB: GetFieldValues(ctx, elementId)
        DB-->>SVC: []CustomFieldValue
        SVC->>DB: ListLinksByElement(ctx, elementId, outgoing)
        DB-->>SVC: []ElementLink
        SVC->>DB: ListLinksByElement(ctx, elementId, incoming)
        DB-->>SVC: []ElementLink
        SVC->>DB: ListAttachmentsByElement(ctx, elementId)
        DB-->>SVC: []Attachment
        SVC->>Cache: Set("element:{id}", element)
    end
    SVC-->>API: *domain.Element
    API-->>Client: 200 ElementResponse (JSON)
```

## Versioning — Pending Change → Commit Flow

```mermaid
sequenceDiagram
    participant Client
    participant API as api/element_handler
    participant SVC as service/ElementService
    participant PendingStore as ElementPendingChangeStore
    participant Worker as service/VersionCommitWorker
    participant VersionStore as ElementVersionStore

    Client->>API: PUT /api/v1/elements/{element_id}
    API->>SVC: UpdateElement(ctx, elementId, element)
    SVC->>DB: UpdateElement(ctx, element)
    SVC->>SVC: BuildSnapshot(element, links, attachments)
    SVC->>SVC: SerializeSnapshot → JSON
    SVC->>PendingStore: UpsertPendingChange(ctx, elementId, snapshotMap)
    SVC->>Cache: Invalidate("element:{id}")
    SVC-->>API: *domain.Element
    API-->>Client: 200 ElementResponse

    Note over Worker: Polls every N seconds
    Worker->>PendingStore: GetPendingChanges(ctx, batchSize)
    PendingStore-->>Worker: []ElementPendingChange
    loop For each pending change
        Worker->>VersionStore: GetLatestVersion(ctx, elementId)
        Worker->>Worker: Compute forward/reverse JSON patches
        Worker->>Worker: Compute content SHA-256
        Worker->>VersionStore: CreateVersion(ctx, version, patch)
        Worker->>PendingStore: DeletePendingChange(ctx, elementId)
    end
```

## Version Reconstruction — Past State

```mermaid
sequenceDiagram
    participant Client
    participant API as api/version_handler
    participant SVC as service/VersionService
    participant ElementStore
    participant VersionStore as ElementVersionStore

    Client->>API: GET /api/v1/elements/{id}/versions/{version_number}
    API->>SVC: GetElementAtVersion(ctx, elementId, targetVersion)
    SVC->>ElementStore: GetElementById + links + attachments + customFields
    SVC->>SVC: BuildSnapshot → serialize to JSON (current state)
    SVC->>VersionStore: ListVersionsByElement(ctx, elementId, 1, 0)
    VersionStore-->>SVC: latestVersionNumber
    SVC->>VersionStore: GetPatchesInRange(ctx, elementId, target+1, latest)
    VersionStore-->>SVC: []ElementVersionPatch (reverse patches)
    loop Apply reverse patches newest-first
        SVC->>SVC: jsonpatch.DecodePatch(reversePatch)
        SVC->>SVC: patch.Apply(currentJSON)
    end
    SVC->>SVC: DeserializeSnapshot → domain.Element
    SVC-->>API: *domain.Element (reconstructed at target version)
    API-->>Client: 200 ElementResponse
```

## Authentication Flow

```mermaid
sequenceDiagram
    participant Client
    participant Middleware as auth/AuthMiddleware
    participant Session as auth/SessionManager
    participant APIKey as store/postgres/APIKeyStore
    participant OIDC as auth/OIDCProvider

    Client->>Middleware: Any request (with cookie or X-API-Key header)
    alt Session cookie present
        Middleware->>Session: ValidateSession(cookie)
        Session-->>Middleware: userId (or error)
    else X-API-Key header present
        Middleware->>APIKey: GetAPIKeyByHash(ctx, hash(key))
        APIKey-->>Middleware: *domain.APIKey (with userId)
    else No credentials
        Middleware-->>Client: 401 Unauthorized
    end
    Middleware->>Handler: ctx with userId injected

    Note over Client, OIDC: OIDC Login Flow
    Client->>OIDC: GET /api/v1/auth/login
    OIDC-->>Client: 302 Redirect to provider
    Client->>OIDC: GET /api/v1/auth/callback?code=...
    OIDC->>OIDC: Exchange code → ID token
    OIDC->>UserService: GetOrCreateUserFromOIDC(claims)
    OIDC->>Session: CreateSession(userId)
    OIDC-->>Client: 302 Redirect + Set-Cookie session
```

## Service Layer Dependencies

```mermaid
graph LR
    ElementSVC["ElementService"] --> ElementStore
    ElementSVC --> ElementLinkStore
    ElementSVC --> ElementPendingChangeStore
    ElementSVC --> CustomFieldValueStore
    ElementSVC --> AttachmentStore
    ElementSVC --> CacheStore

    LinkSVC["LinkService"] --> ElementLinkStore
    LinkSVC --> ElementPendingChangeStore
    LinkSVC --> ElementStore
    LinkSVC --> CustomFieldValueStore
    LinkSVC --> AttachmentStore
    LinkSVC --> CacheStore

    AttachmentSVC["AttachmentService"] --> AttachmentStore
    AttachmentSVC --> FileStorage
    AttachmentSVC --> ElementPendingChangeStore
    AttachmentSVC --> ElementStore
    AttachmentSVC --> ElementLinkStore
    AttachmentSVC --> CustomFieldValueStore
    AttachmentSVC --> CacheStore

    VersionSVC["VersionService"] --> ElementVersionStore
    VersionSVC --> ElementStore
    VersionSVC --> ElementLinkStore
    VersionSVC --> CustomFieldValueStore
    VersionSVC --> AttachmentStore

    VersionWorker["VersionCommitWorker"] --> ElementPendingChangeStore
    VersionWorker --> ElementVersionStore
    VersionWorker --> ElementStore
    VersionWorker --> ElementLinkStore
    VersionWorker --> CustomFieldValueStore
    VersionWorker --> AttachmentStore
```

## API Endpoints Overview

```mermaid
graph LR
    subgraph Projects
        P1["GET /api/v1/projects"]
        P2["POST /api/v1/projects"]
        P3["GET /api/v1/projects/{id}"]
        P4["PUT /api/v1/projects/{id}"]
        P5["DELETE /api/v1/projects/{id}"]
    end

    subgraph Elements
        E1["GET /api/v1/projects/{id}/elements"]
        E2["POST /api/v1/projects/{id}/elements"]
        E3["GET /api/v1/elements/{id}"]
        E4["PUT /api/v1/elements/{id}"]
        E5["DELETE /api/v1/elements/{id}"]
        E6["GET /api/v1/elements/search"]
    end

    subgraph Links
        L1["GET /api/v1/elements/{id}/links"]
        L2["POST /api/v1/elements/{id}/links"]
        L3["PATCH /api/v1/links/{id}"]
        L4["DELETE /api/v1/links/{id}"]
    end

    subgraph Versions
        V1["GET /api/v1/elements/{id}/versions"]
        V2["GET /api/v1/elements/{id}/versions/{n}"]
        V3["GET /api/v1/elements/{id}/versions/{n}/diff"]
    end

    subgraph Attachments
        A1["GET /api/v1/elements/{id}/attachments"]
        A2["POST /api/v1/elements/{id}/attachments"]
        A3["GET /api/v1/attachments/{id}"]
        A4["DELETE /api/v1/attachments/{id}"]
    end

    subgraph Access
        AC1["GET /api/v1/projects/{id}/access"]
        AC2["PUT /api/v1/projects/{id}/access"]
        AC3["DELETE /api/v1/projects/{id}/access/{groupId}"]
        AC4["GET /api/v1/groups/{id}/access"]
        AC5["GET /api/v1/projects/{id}/access/check"]
    end
```
