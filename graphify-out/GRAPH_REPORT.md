# Graph Report - .  (2026-06-22)

## Corpus Check
- 174 files · ~329,283 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary
- 992 nodes · 1678 edges · 73 communities (55 shown, 18 thin omitted)
- Extraction: 91% EXTRACTED · 9% INFERRED · 0% AMBIGUOUS · INFERRED: 155 edges (avg confidence: 0.81)
- Token cost: 65,000 input · 8,602 output

## Community Hubs (Navigation)
- [[_COMMUNITY_Domain Model & Test Fakes|Domain Model & Test Fakes]]
- [[_COMMUNITY_API Key HTTP Handlers|API Key HTTP Handlers]]
- [[_COMMUNITY_Element Versioning Worker|Element Versioning Worker]]
- [[_COMMUNITY_Access & Attachment Handlers|Access & Attachment Handlers]]
- [[_COMMUNITY_OIDC Auth & Sessions|OIDC Auth & Sessions]]
- [[_COMMUNITY_Architecture & Design Concepts|Architecture & Design Concepts]]
- [[_COMMUNITY_Element HTTP Handlers|Element HTTP Handlers]]
- [[_COMMUNITY_Group Access Store|Group Access Store]]
- [[_COMMUNITY_CLI Client (main.go)|CLI Client (main.go)]]
- [[_COMMUNITY_Group Service Layer|Group Service Layer]]
- [[_COMMUNITY_Project Service Layer|Project Service Layer]]
- [[_COMMUNITY_Group HTTP Handlers|Group HTTP Handlers]]
- [[_COMMUNITY_Ristretto Cache|Ristretto Cache]]
- [[_COMMUNITY_Version Service Layer|Version Service Layer]]
- [[_COMMUNITY_Version HTTP Handlers|Version HTTP Handlers]]
- [[_COMMUNITY_Custom Field Service|Custom Field Service]]
- [[_COMMUNITY_Element Service Layer|Element Service Layer]]
- [[_COMMUNITY_Custom Field Handlers|Custom Field Handlers]]
- [[_COMMUNITY_Project HTTP Handlers|Project HTTP Handlers]]
- [[_COMMUNITY_Attachment Service Layer|Attachment Service Layer]]
- [[_COMMUNITY_Element Link Service|Element Link Service]]
- [[_COMMUNITY_Phase Service Layer|Phase Service Layer]]
- [[_COMMUNITY_Router & Service Wiring|Router & Service Wiring]]
- [[_COMMUNITY_Element Core Types|Element Core Types]]
- [[_COMMUNITY_Link Service Tests|Link Service Tests]]
- [[_COMMUNITY_User Service Layer|User Service Layer]]
- [[_COMMUNITY_Config Loading|Config Loading]]
- [[_COMMUNITY_Phase HTTP Handlers|Phase HTTP Handlers]]
- [[_COMMUNITY_S3 File Storage|S3 File Storage]]
- [[_COMMUNITY_Element Link Store|Element Link Store]]
- [[_COMMUNITY_Element Store|Element Store]]
- [[_COMMUNITY_Project Store|Project Store]]
- [[_COMMUNITY_Auth HTTP Handlers|Auth HTTP Handlers]]
- [[_COMMUNITY_Link HTTP Handlers|Link HTTP Handlers]]
- [[_COMMUNITY_User HTTP Handlers|User HTTP Handlers]]
- [[_COMMUNITY_Task Status Enum|Task Status Enum]]
- [[_COMMUNITY_API Key Store|API Key Store]]
- [[_COMMUNITY_Attachment Store|Attachment Store]]
- [[_COMMUNITY_User Store|User Store]]
- [[_COMMUNITY_Element Pending-Change Store|Element Pending-Change Store]]
- [[_COMMUNITY_Element Version Store|Element Version Store]]
- [[_COMMUNITY_Phase Store|Phase Store]]
- [[_COMMUNITY_Element Type Enum|Element Type Enum]]
- [[_COMMUNITY_Link Type Enum|Link Type Enum]]
- [[_COMMUNITY_User & Group Model|User & Group Model]]
- [[_COMMUNITY_Version Model|Version Model]]
- [[_COMMUNITY_UA Layer-Assign Script|UA Layer-Assign Script]]
- [[_COMMUNITY_Element Store Interface|Element Store Interface]]
- [[_COMMUNITY_User Store Interface|User Store Interface]]
- [[_COMMUNITY_Authentication Flow|Authentication Flow]]
- [[_COMMUNITY_Element Link Model|Element Link Model]]
- [[_COMMUNITY_Custom Field Handler Wiring|Custom Field Handler Wiring]]
- [[_COMMUNITY_Link Handler Wiring|Link Handler Wiring]]
- [[_COMMUNITY_Huma OpenAPI Framework|Huma OpenAPI Framework]]
- [[_COMMUNITY_Custom Field Model|Custom Field Model]]
- [[_COMMUNITY_Custom Field Store Interface|Custom Field Store Interface]]
- [[_COMMUNITY_Element Link Store Interface|Element Link Store Interface]]
- [[_COMMUNITY_Version Store Interface|Version Store Interface]]
- [[_COMMUNITY_UA Arch-Analyze Script|UA Arch-Analyze Script]]
- [[_COMMUNITY_UA Tour-Analyze Script|UA Tour-Analyze Script]]
- [[_COMMUNITY_Attachment Model|Attachment Model]]
- [[_COMMUNITY_Phase Model|Phase Model]]
- [[_COMMUNITY_Project Model|Project Model]]
- [[_COMMUNITY_Attachment Store Interface|Attachment Store Interface]]
- [[_COMMUNITY_Cache Store Interface|Cache Store Interface]]
- [[_COMMUNITY_File Storage Interface|File Storage Interface]]
- [[_COMMUNITY_Phase Store Interface|Phase Store Interface]]
- [[_COMMUNITY_Project Store Interface|Project Store Interface]]
- [[_COMMUNITY_Test Results Parser Script|Test Results Parser Script]]
- [[_COMMUNITY_Gantt View Concept|Gantt View Concept]]
- [[_COMMUNITY_Plugin System Concept|Plugin System Concept]]
- [[_COMMUNITY_Backend Module Root|Backend Module Root]]

## God Nodes (most connected - your core abstractions)
1. `Context` - 40 edges
2. `main()` - 35 edges
3. `NewRouter()` - 28 edges
4. `newLinkServiceWithFakes()` - 18 edges
5. `testContext` - 17 edges
6. `CustomFieldService` - 17 edges
7. `GroupService` - 17 edges
8. `BuildSnapshot()` - 16 edges
9. `Config` - 15 edges
10. `ElementService` - 15 edges

## Surprising Connections (you probably didn't know these)
- `main()` --calls--> `Load()`  [INFERRED]
  backend/cmd/server/main.go → backend/internal/config/config.go
- `just unit-test-ci (CI command)` --references--> `Project Works (README)`  [INFERRED]
  .github/workflows/tests.yml → README.md
- `CacheStore interface` --conceptually_related_to--> `CQRS-lite Read Performance`  [INFERRED]
  DATA_MODEL.md → ARCHITECTURE.md
- `Element Read Cache-Path Flow` --references--> `CQRS-lite Read Performance`  [INFERRED]
  backend/DIAGRAM.md → ARCHITECTURE.md
- `Project Works API (OpenAPI 3.1)` --references--> `OpenAPI-First API`  [INFERRED]
  backend/openapi.yaml → ARCHITECTURE.md

## Import Cycles
- None detected.

## Hyperedges (group relationships)
- **Auto-commit Versioning Pipeline** — data_model_auto_commit, diagram_versioncommitworker, data_model_git_like_versioning, config_yaml [INFERRED 0.85]
- **Pluggable Storage Swap Pattern** — architecture_clean_architecture, architecture_store_interface, data_model_filestorage, data_model_cachestore [INFERRED 0.75]
- **Local Dev Service Stack** — docker_compose_backend, docker_compose_database, docker_compose_file_storage, config_yaml [EXTRACTED 1.00]

## Communities (73 total, 18 thin omitted)

### Community 0 - "Domain Model & Test Fakes"
Cohesion: 0.06
Nodes (29): Attachment, Context, CustomFieldValue, Duration, Element, ElementFilter, ElementLink, LinkDirection (+21 more)

### Community 1 - "API Key HTTP Handlers"
Cohesion: 0.08
Nodes (39): RegisterAPIKeyHandlers(), APIKeyGeneratedResponse, APIKeyResponse, DeleteAPIKeyInput, DeleteAPIKeyOutput, GenerateAPIKeyInput, GenerateAPIKeyOutput, ListAPIKeysInput (+31 more)

### Community 2 - "Element Versioning Worker"
Cohesion: 0.06
Nodes (46): Attachment, AttachmentStore, Context, CustomFieldValueStore, Duration, Element, ElementLink, ElementLinkStore (+38 more)

### Community 3 - "Access & Attachment Handlers"
Cohesion: 0.06
Nodes (38): RegisterAccessHandlers(), RegisterAttachmentHandlers(), AttachmentResponse, CheckProjectAccessInput, CheckProjectAccessOutput, DeleteAttachmentInput, DeleteAttachmentOutput, GetAttachmentPresignedURLInput (+30 more)

### Community 4 - "OIDC Auth & Sessions"
Cohesion: 0.08
Nodes (22): NewOIDCProvider(), OIDCProvider, NewSessionManager(), SessionManager, Context, DB, Duration, Context (+14 more)

### Community 5 - "Architecture & Design Concepts"
Cohesion: 0.06
Nodes (35): Layered / Clean Architecture, Coupling Matrix View, CQRS-lite Read Performance, Custom Fields (JSONB), DFP Matrix View, Folder-path Project Tree, PERT / Critical Path View, PostgreSQL 16 (default DB) (+27 more)

### Community 6 - "Element HTTP Handlers"
Cohesion: 0.12
Nodes (24): CreateElementInput, CreateElementOutput, DeleteElementInput, DeleteElementOutput, mapElementToResponse(), parseCommaSeparated(), parseCommaSeparatedElementTypes(), parseCommaSeparatedTaskStatuses() (+16 more)

### Community 7 - "Group Access Store"
Cohesion: 0.16
Nodes (10): AccessLevel, Context, DB, Group, GroupProjectAccess, User, NewGroupProjectAccessStore(), NewGroupStore() (+2 more)

### Community 8 - "CLI Client (main.go)"
Cohesion: 0.38
Nodes (17): Client, getItems(), getString(), main(), testAttachmentLifecycle(), testAuthMeEndpoint(), testCustomFieldLifecycle(), testElementLifecycle() (+9 more)

### Community 9 - "Group Service Layer"
Cohesion: 0.17
Nodes (9): AccessLevel, Context, Group, GroupProjectAccess, User, GroupProjectAccessStore, GroupStore, NewGroupService() (+1 more)

### Community 10 - "Project Service Layer"
Cohesion: 0.21
Nodes (15): CacheStore, Context, Project, T, ProjectStore, NewProjectService(), TestProjectService_CreateProject_PrimesCache(), TestProjectService_DeleteProject_RemovesAndInvalidatesCache() (+7 more)

### Community 11 - "Group HTTP Handlers"
Cohesion: 0.11
Nodes (21): AddGroupMemberInput, AddGroupMemberOutput, CreateGroupInput, CreateGroupOutput, DeleteGroupInput, DeleteGroupOutput, GetGroupInput, GetGroupOutput (+13 more)

### Community 12 - "Ristretto Cache"
Cohesion: 0.17
Nodes (16): CacheStore, Context, Duration, T, Cache, NewRistrettoCache(), newTestCache(), TestRistrettoCache_GetMissReturnsFalse() (+8 more)

### Community 13 - "Version Service Layer"
Cohesion: 0.16
Nodes (15): Attachment, AttachmentStore, Context, CustomFieldValueStore, Element, ElementLink, ElementLinkStore, ElementStore (+7 more)

### Community 14 - "Version HTTP Handlers"
Cohesion: 0.14
Nodes (20): ElementVersionDiffResponse, ElementVersionMetaResponse, GetElementAtVersionInput, GetElementAtVersionOutput, GetElementVersionDiffInput, GetElementVersionDiffOutput, ListElementVersionsInput, ListElementVersionsOutput (+12 more)

### Community 15 - "Custom Field Service"
Cohesion: 0.19
Nodes (12): AttachmentStore, CacheStore, Context, CustomFieldDefinition, CustomFieldValue, CustomFieldValueStore, ElementLinkStore, ElementPendingChangeStore (+4 more)

### Community 16 - "Element Service Layer"
Cohesion: 0.19
Nodes (13): Attachment, AttachmentStore, CacheStore, Context, CustomFieldValueStore, Element, ElementFilter, ElementLink (+5 more)

### Community 17 - "Custom Field Handlers"
Cohesion: 0.13
Nodes (18): CreateCustomFieldDefinitionInput, CreateCustomFieldDefinitionOutput, CustomFieldDefinitionResponse, CustomFieldValueResponse, DeleteCustomFieldDefinitionInput, DeleteCustomFieldDefinitionOutput, DeleteCustomFieldValueInput, DeleteCustomFieldValueOutput (+10 more)

### Community 18 - "Project HTTP Handlers"
Cohesion: 0.14
Nodes (18): CreateProjectInput, CreateProjectOutput, DeleteProjectInput, DeleteProjectOutput, GetProjectInput, GetProjectOutput, ListFolderPathsOutput, ListProjectsInput (+10 more)

### Community 19 - "Attachment Service Layer"
Cohesion: 0.20
Nodes (13): Attachment, AttachmentStore, CacheStore, Context, CustomFieldValueStore, Duration, ElementLinkStore, ElementPendingChangeStore (+5 more)

### Community 20 - "Element Link Service"
Cohesion: 0.24
Nodes (11): AttachmentStore, CacheStore, Context, CustomFieldValueStore, ElementLink, ElementLinkStore, ElementPendingChangeStore, ElementStore (+3 more)

### Community 21 - "Phase Service Layer"
Cohesion: 0.24
Nodes (11): Context, Phase, T, PhaseStore, NewPhaseService(), TestPhaseService_CreateAndListByProject(), TestPhaseService_DeletePhase(), TestPhaseService_DeletePhase_NotFoundPropagates() (+3 more)

### Community 22 - "Router & Service Wiring"
Cohesion: 0.12
Nodes (15): NewRouter(), AuthMiddleware, APIKeyStore, AttachmentService, CustomFieldService, ElementService, GroupService, Handler (+7 more)

### Community 23 - "Element Core Types"
Cohesion: 0.23
Nodes (15): CustomFieldValue, ElementType, TaskStatus, Time, T, ContactInfo, ContactInfo, Element (+7 more)

### Community 24 - "Link Service Tests"
Cohesion: 0.27
Nodes (15): LinkService, T, fakeCacheStore, fakeElementLinkStore, fakeElementStore, fakePendingChangeStore, containsString(), newLinkServiceWithFakes() (+7 more)

### Community 25 - "User Service Layer"
Cohesion: 0.25
Nodes (9): Context, User, UserStore, T, NewUserService(), TestUserService_CreateAndGetUser(), TestUserService_GetOrCreateUserFromOIDC_CreatesNewUser(), TestUserService_GetOrCreateUserFromOIDC_ReturnsExistingUser() (+1 more)

### Community 26 - "Config Loading"
Cohesion: 0.16
Nodes (15): AuthConfig, Duration, CacheConfig, AuthConfig, CacheConfig, Config, Load(), DatabaseConfig (+7 more)

### Community 27 - "Phase HTTP Handlers"
Cohesion: 0.20
Nodes (13): CreateProjectPhaseInput, CreateProjectPhaseOutput, DeletePhaseInput, DeletePhaseOutput, ListProjectPhasesInput, ListProjectPhasesOutput, RegisterPhaseHandlers(), PhaseResponse (+5 more)

### Community 28 - "S3 File Storage"
Cohesion: 0.19
Nodes (9): Client, Context, Duration, FileStorage, Reader, PresignClient, ReadCloser, NewS3FileStorage() (+1 more)

### Community 29 - "Element Link Store"
Cohesion: 0.27
Nodes (8): Context, DB, ElementLink, LinkDirection, LinkType, NewElementLinkStore(), scanLink(), ElementLinkStore

### Community 30 - "Element Store"
Cohesion: 0.34
Nodes (7): Context, DB, Element, ElementFilter, NewElementStore(), scanElement(), ElementStore

### Community 31 - "Project Store"
Cohesion: 0.32
Nodes (5): Context, DB, Project, NewProjectStore(), ProjectStore

### Community 32 - "Auth HTTP Handlers"
Cohesion: 0.25
Nodes (10): RegisterAuthHandlers(), RegisterAuthHandlersWithHuma(), CurrentUserResponse, LoginResponse, LogoutResponse, OIDCProvider, SessionManager, UserService (+2 more)

### Community 33 - "Link HTTP Handlers"
Cohesion: 0.24
Nodes (10): CreateElementLinkInput, CreateElementLinkOutput, DeleteLinkInput, DeleteLinkOutput, LinkResponse, ListElementLinksInput, ListElementLinksOutput, UpdateLinkInput (+2 more)

### Community 34 - "User HTTP Handlers"
Cohesion: 0.24
Nodes (10): GetUserInput, GetUserOutput, ListUsersInput, ListUsersOutput, UpdateUserInput, UpdateUserOutput, RegisterUserHandlers(), UserResponse (+2 more)

### Community 35 - "Task Status Enum"
Cohesion: 0.27
Nodes (6): T, AllTaskStatuses(), TestAllTaskStatuses_ContainsAllEightValues(), TestTaskStatus_IsOpenAndClosed(), TestTaskStatus_IsValid(), TaskStatus

### Community 36 - "API Key Store"
Cohesion: 0.33
Nodes (5): APIKey, Context, DB, NewAPIKeyStore(), APIKeyStore

### Community 37 - "Attachment Store"
Cohesion: 0.36
Nodes (6): Attachment, Context, DB, NewAttachmentStore(), scanAttachment(), AttachmentStore

### Community 38 - "User Store"
Cohesion: 0.36
Nodes (5): Context, DB, User, NewUserStore(), UserStore

### Community 39 - "Element Pending-Change Store"
Cohesion: 0.29
Nodes (6): Context, DB, Duration, PendingChange, NewElementPendingChangeStore(), ElementPendingChangeStore

### Community 40 - "Element Version Store"
Cohesion: 0.33
Nodes (6): Context, DB, ElementVersion, ElementVersionPatch, NewElementVersionStore(), ElementVersionStore

### Community 41 - "Phase Store"
Cohesion: 0.36
Nodes (5): Context, DB, Phase, NewPhaseStore(), PhaseStore

### Community 42 - "Element Type Enum"
Cohesion: 0.36
Nodes (5): T, AllElementTypes(), TestAllElementTypes_ContainsAllSixValuesUnique(), TestElementType_IsValid(), ElementType

### Community 43 - "Link Type Enum"
Cohesion: 0.36
Nodes (5): T, AllLinkTypes(), TestAllLinkTypes_ContainsAllThreeValuesUnique(), TestLinkType_IsValid(), LinkType

### Community 44 - "User & Group Model"
Cohesion: 0.38
Nodes (7): Time, AccessLevel, APIKey, Group, GroupMembership, GroupProjectAccess, User

### Community 45 - "Version Model"
Cohesion: 0.33
Nodes (6): PatchOperation, Time, ElementVersion, ElementVersionPatch, PatchOperation, PendingChange

### Community 46 - "UA Layer-Assign Script"
Cohesion: 0.29
Nodes (5): fn, fs, layers, out, total

### Community 47 - "Element Store Interface"
Cohesion: 0.40
Nodes (4): ElementType, TaskStatus, ElementFilter, ElementStore

### Community 48 - "User Store Interface"
Cohesion: 0.40
Nodes (4): APIKeyStore, GroupProjectAccessStore, GroupStore, UserStore

### Community 49 - "Authentication Flow"
Cohesion: 0.50
Nodes (4): Authentication Flow (OIDC + API key), OIDC + Office 365 Auth, AuthMiddleware, Security Schemes (bearerApiKey, cookieSession)

### Community 50 - "Element Link Model"
Cohesion: 0.50
Nodes (3): LinkType, Time, ElementLink

### Community 51 - "Custom Field Handler Wiring"
Cohesion: 0.67
Nodes (3): RegisterCustomFieldHandlers(), API, CustomFieldService

### Community 52 - "Link Handler Wiring"
Cohesion: 0.67
Nodes (3): RegisterLinkHandlers(), API, LinkService

### Community 53 - "Huma OpenAPI Framework"
Cohesion: 1.00
Nodes (3): Huma v2 (Go web framework), OpenAPI-First API, Project Works API (OpenAPI 3.1)

## Knowledge Gaps
- **254 isolated node(s):** `fs`, `fs`, `fn`, `layers`, `out` (+249 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **18 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `main()` connect `OIDC Auth & Sessions` to `API Key HTTP Handlers`, `Element Versioning Worker`, `Access & Attachment Handlers`, `Group Access Store`, `Group Service Layer`, `Project Service Layer`, `Ristretto Cache`, `Version Service Layer`, `Custom Field Service`, `Element Service Layer`, `Attachment Service Layer`, `Element Link Service`, `Phase Service Layer`, `Router & Service Wiring`, `User Service Layer`, `Config Loading`, `S3 File Storage`, `Element Link Store`, `Element Store`, `Project Store`, `API Key Store`, `Attachment Store`, `User Store`, `Element Pending-Change Store`, `Element Version Store`, `Phase Store`?**
  _High betweenness centrality (0.553) - this node is a cross-community bridge._
- **Why does `NewRouter()` connect `Router & Service Wiring` to `Auth HTTP Handlers`, `API Key HTTP Handlers`, `User HTTP Handlers`, `Access & Attachment Handlers`, `OIDC Auth & Sessions`, `Element HTTP Handlers`, `Group HTTP Handlers`, `Version HTTP Handlers`, `Project HTTP Handlers`, `Custom Field Handler Wiring`, `Link Handler Wiring`, `Phase HTTP Handlers`?**
  _High betweenness centrality (0.302) - this node is a cross-community bridge._
- **Why does `NewLinkService()` connect `Element Link Service` to `Link Service Tests`, `OIDC Auth & Sessions`?**
  _High betweenness centrality (0.110) - this node is a cross-community bridge._
- **Are the 33 inferred relationships involving `main()` (e.g. with `NewRouter()` and `NewAuthMiddleware()`) actually correct?**
  _`main()` has 33 INFERRED edges - model-reasoned connections that need verification._
- **Are the 13 inferred relationships involving `NewRouter()` (e.g. with `RegisterAccessHandlers()` and `RegisterAPIKeyHandlers()`) actually correct?**
  _`NewRouter()` has 13 INFERRED edges - model-reasoned connections that need verification._
- **Are the 7 inferred relationships involving `newLinkServiceWithFakes()` (e.g. with `NewLinkService()` and `newFakeAttachmentStore()`) actually correct?**
  _`newLinkServiceWithFakes()` has 7 INFERRED edges - model-reasoned connections that need verification._
- **What connects `fs`, `fs`, `fn` to the rest of the system?**
  _259 weakly-connected nodes found - possible documentation gaps or missing edges._