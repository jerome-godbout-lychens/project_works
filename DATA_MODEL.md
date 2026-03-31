# Data Model Design

## Design Goals

1. **Fast current-state reads** — the "hot" tables hold only the latest version of each element. No joins to version tables needed for normal browsing.
2. **Git-like versioning** — history stored as JSON Patch diffs (RFC 6902). Reconstructing a past version applies patches backward from current state. Browsing history is infrequent and does not need to be fast.
3. **Auto-commit on inactivity** — pending changes accumulate; after a configurable inactivity window (e.g., 5 minutes with no edits), the system creates a version snapshot diff.
4. **Decoupled** — domain types have zero database dependencies. Storage is behind interfaces.

---

## Table Layout Overview

```
HOT PATH (current state — optimized for reads)
├── projects
├── project_folders
├── elements                     ← current state of every element
├── element_custom_field_values  ← current custom field values
├── element_links                ← current links between elements
├── element_attachments          ← current attachment references
├── custom_field_definitions     ← schema of custom fields per element type
├── phases
├── users
├── groups
├── group_memberships
├── group_project_access
└── api_keys

COLD PATH (version history — append-only)
├── element_versions             ← version metadata (sha, author, timestamp)
├── element_version_patches      ← JSON Patch diffs per version
└── element_pending_changes      ← staging area before auto-commit
```

---

## Hot Path Tables

### projects

| Column | Type | Notes |
|---|---|---|
| project_identifier | UUID | PK |
| project_name | TEXT | NOT NULL |
| project_description | TEXT | |
| folder_path | LTREE | e.g. `engineering.firmware.sensors` |
| creation_time | TIMESTAMPTZ | NOT NULL, DEFAULT now() |
| modification_time | TIMESTAMPTZ | NOT NULL, auto-updated |

Index: `folder_path` using GiST for ltree queries.

### elements

The **single table** for current state of all element types. Type-specific columns are nullable — only the columns relevant to that `element_type` are populated.

| Column | Type | Notes |
|---|---|---|
| element_identifier | UUID | PK |
| project_identifier | UUID | FK → projects |
| element_type | TEXT | `requirement`, `feature`, `task`, `bug`, `evaluation`, `risk` |
| title | TEXT | NOT NULL |
| description | TEXT | Markdown + Mermaid |
| content_sha | TEXT | SHA-256 of serialized element content (current version) |
| creation_time | TIMESTAMPTZ | NOT NULL, DEFAULT now() |
| modification_time | TIMESTAMPTZ | NOT NULL, auto-updated |
| — **requirement-specific** | | |
| interest_level | SMALLINT | 1–10, NULL for non-requirements |
| — **task-specific** | | |
| assignee_identifier | UUID | FK → users, NULL for non-tasks |
| task_status | TEXT | `backlog`, `todo`, `in_progress`, `in_review`, `testing`, `blocked`, `done`, `rejected` |
| task_progress | SMALLINT | 0–100 percent, NULL for non-tasks |
| close_time | TIMESTAMPTZ | Set when status → closed |
| parent_feature_identifier | UUID | FK → elements (self-ref), NULL for non-tasks |
| start_phase_identifier | UUID | FK → phases |
| delivery_phase_identifier | UUID | FK → phases |

**Why single table instead of polymorphic tables?**
Querying "all elements in a project" is one indexed scan, no unions/joins. Type-specific columns are sparse but PostgreSQL handles NULLs efficiently (bitmap). For custom fields (which are user-defined and open-ended), we use a separate table.

Indexes:

- `(project_identifier, element_type)` — filter by project + type
- `(project_identifier)` — all elements in a project
- `(parent_feature_identifier)` — tasks under a feature
- `(element_type, task_status)` — filter by status across project
- GIN on `title || ' ' || description` using `tsvector` — full-text search

### element_custom_field_values

| Column | Type | Notes |
|---|---|---|
| element_identifier | UUID | FK → elements, part of composite PK |
| field_definition_identifier | UUID | FK → custom_field_definitions, part of composite PK |
| field_value | JSONB | Stores the value; type validation happens in domain layer |

Index: GIN on `field_value` for search.

### element_links

| Column | Type | Notes |
|---|---|---|
| link_identifier | UUID | PK |
| source_element_identifier | UUID | FK → elements, NOT NULL |
| destination_element_identifier | UUID | FK → elements, NOT NULL |
| link_type | TEXT | `related`, `child`, `implement` |
| creation_time | TIMESTAMPTZ | NOT NULL |

Indexes:

- `(source_element_identifier, link_type)`
- `(destination_element_identifier, link_type)`
- UNIQUE `(source_element_identifier, destination_element_identifier, link_type)` — prevent duplicate links

### element_attachments

| Column | Type | Notes |
|---|---|---|
| attachment_identifier | UUID | PK |
| element_identifier | UUID | FK → elements |
| file_storage_key | TEXT | S3/SeaweedFS object key |
| file_name | TEXT | Original filename |
| file_size_bytes | BIGINT | |
| content_type | TEXT | MIME type |
| upload_time | TIMESTAMPTZ | NOT NULL |
| uploaded_by_identifier | UUID | FK → users |

### custom_field_definitions

| Column | Type | Notes |
|---|---|---|
| field_definition_identifier | UUID | PK |
| project_identifier | UUID | FK → projects |
| applicable_element_type | TEXT | Which element type this field applies to, or `*` for all |
| field_name | TEXT | NOT NULL |
| field_type | TEXT | `string`, `textarea`, `integer`, `real`, `choice` |
| field_options | JSONB | e.g. choices list, min/max, default value |
| display_order | INTEGER | Ordering in UI |

UNIQUE: `(project_identifier, applicable_element_type, field_name)`

### phases

| Column | Type | Notes |
|---|---|---|
| phase_identifier | UUID | PK |
| project_identifier | UUID | FK → projects |
| phase_name | TEXT | NOT NULL |
| phase_order | INTEGER | Sequencing |
| planned_start_date | DATE | |
| planned_end_date | DATE | |

### Supervisors & Clients (many-to-many)

| Table | Columns |
|---|---|
| element_supervisors | `element_identifier`, `user_identifier` |
| requirement_clients | `element_identifier`, `contact_name`, `contact_email`, `contact_notes` |

### Auth tables

| Table | Key Columns |
|---|---|
| users | `user_identifier`, `email`, `display_name`, `external_identity_provider`, `external_identity_subject` |
| groups | `group_identifier`, `group_name` |
| group_memberships | `group_identifier`, `user_identifier` |
| group_project_access | `group_identifier`, `project_identifier`, `access_level` (read/write/admin) |
| api_keys | `api_key_identifier`, `user_identifier`, `hashed_key`, `label`, `created_time`, `last_used_time` |

---

## Cold Path — Versioning

### How it works

```
User edits element
        ↓
Write change to `elements` table (hot) + insert into `element_pending_changes` (staging)
        ↓
Inactivity timer fires (configurable, e.g. 5 min no edits on that element)
        ↓
Auto-commit job:
  1. Compute JSON Patch diff between previous committed state and current state
  2. Compute new content_sha = SHA-256(serialized current state)
  3. Insert version metadata into `element_versions`
  4. Insert patch into `element_version_patches`
  5. Update `elements.content_sha`
  6. Clear `element_pending_changes` for that element
```

### Reconstructing a past version

```
Start from current state (hot table)
        ↓
Walk `element_version_patches` backward (newest → target version)
        ↓
Apply reverse JSON Patches to reconstruct target state
```

This is slow for very old versions — acceptable per requirements. If needed later, periodic full snapshots can be added (like git packfiles) without schema changes.

### element_versions

| Column | Type | Notes |
|---|---|---|
| version_identifier | UUID | PK |
| element_identifier | UUID | FK → elements |
| version_number | INTEGER | Monotonically increasing per element |
| content_sha | TEXT | SHA-256 at this version |
| committed_time | TIMESTAMPTZ | When the auto-commit happened |
| committed_by_identifier | UUID | FK → users (who made the last edit before commit) |
| commit_message | TEXT | Optional, auto-generated or user-provided |

Index: `(element_identifier, version_number DESC)` — walk history backward.

### element_version_patches

| Column | Type | Notes |
|---|---|---|
| version_identifier | UUID | FK → element_versions, PK |
| forward_patch | JSONB | JSON Patch to go from version N-1 → N |
| reverse_patch | JSONB | JSON Patch to go from version N → N-1 (pre-computed for fast backward walk) |

Storing both forward and reverse patches avoids recomputing the inverse at read time.

### element_pending_changes

| Column | Type | Notes |
|---|---|---|
| element_identifier | UUID | FK → elements, UNIQUE |
| last_edit_time | TIMESTAMPTZ | Reset on every edit — inactivity timer checks this |
| snapshot_before_edits | JSONB | State of the element at last commit (to compute diff later) |

This table is transient — rows exist only while an element has uncommitted edits. The auto-commit job polls or is triggered by a timer.

---

## Go Domain Types

These types live in `domain/` and have **zero external dependencies**. Storage and API layers convert to/from these types.

```go
package domain

import "time"

// ElementType enumerates the kinds of trackable items.
type ElementType string

const (
    ElementTypeRequirement ElementType = "requirement"
    ElementTypeFeature     ElementType = "feature"
    ElementTypeTask        ElementType = "task"
    ElementTypeBug         ElementType = "bug"
    ElementTypeEvaluation  ElementType = "evaluation"
    ElementTypeRisk        ElementType = "risk"
)

// TaskStatus represents the lifecycle state of a task.
type TaskStatus string

const (
    TaskStatusBacklog    TaskStatus = "backlog"
    TaskStatusTodo       TaskStatus = "todo"
    TaskStatusInProgress TaskStatus = "in_progress"
    TaskStatusInReview   TaskStatus = "in_review"
    TaskStatusTesting    TaskStatus = "testing"
    TaskStatusBlocked    TaskStatus = "blocked"
    TaskStatusDone       TaskStatus = "done"
    TaskStatusRejected   TaskStatus = "rejected"
)

// LinkType defines how two elements are related.
type LinkType string

const (
    LinkTypeRelated   LinkType = "related"
    LinkTypeChild     LinkType = "child"
    LinkTypeImplement LinkType = "implement"
)

// Element is the core domain entity. Type-specific fields are non-nil
// only when relevant to the ElementType.
type Element struct {
    ElementIdentifier  string
    ProjectIdentifier  string
    ElementType        ElementType
    Title              string
    Description        string
    ContentSha         string
    CreationTime       time.Time
    ModificationTime   time.Time

    // Requirement-specific
    InterestLevel *int

    // Task-specific
    AssigneeIdentifier         *string
    TaskStatus                 *TaskStatus
    TaskProgress               *int
    CloseTime                  *time.Time
    ParentFeatureIdentifier    *string
    StartPhaseIdentifier       *string
    DeliveryPhaseIdentifier    *string

    // Populated via joins / separate queries (not stored inline)
    CustomFieldValues []CustomFieldValue
    Supervisors       []string   // user identifiers
    Clients           []ContactInfo
}

// ComputeFeatureProgress calculates a feature's progress from its child tasks.
// This is a derived value — never stored in the database.
// Returns 0 if the feature has no child tasks.
func ComputeFeatureProgress(childTasks []Element) int {
    if len(childTasks) == 0 {
        return 0
    }
    totalProgress := 0
    for _, task := range childTasks {
        if task.TaskProgress != nil {
            totalProgress += *task.TaskProgress
        }
    }
    return totalProgress / len(childTasks)
}

// CustomFieldValue holds a single custom field's current value.
type CustomFieldValue struct {
    FieldDefinitionIdentifier string
    FieldName                 string
    FieldValue                interface{}
}

// ContactInfo represents a client contact on a requirement.
type ContactInfo struct {
    ContactName  string
    ContactEmail string
    ContactNotes string
}

// ElementLink represents a directional relationship between two elements.
type ElementLink struct {
    LinkIdentifier               string
    SourceElementIdentifier      string
    DestinationElementIdentifier string
    LinkType                     LinkType
    CreationTime                 time.Time
}

// ElementVersion is the metadata for one committed version.
type ElementVersion struct {
    VersionIdentifier       string
    ElementIdentifier       string
    VersionNumber           int
    ContentSha              string
    CommittedTime           time.Time
    CommittedByIdentifier   string
    CommitMessage           string
}

// ElementVersionPatch holds the diff between two consecutive versions.
type ElementVersionPatch struct {
    VersionIdentifier string
    ForwardPatch      []PatchOperation // version N-1 → N
    ReversePatch      []PatchOperation // version N → N-1
}

// PatchOperation is one step in a JSON Patch (RFC 6902).
type PatchOperation struct {
    Operation string      // "add", "remove", "replace", "move", "copy", "test"
    Path      string      // JSON Pointer (RFC 6901)
    Value     interface{} // new value (for add/replace)
    From      string      // source path (for move/copy)
}

// Project represents a trackable project.
type Project struct {
    ProjectIdentifier  string
    ProjectName        string
    ProjectDescription string
    FolderPath         string // ltree path e.g. "engineering.firmware"
    CreationTime       time.Time
    ModificationTime   time.Time
}

// Phase represents a project phase for task scheduling.
type Phase struct {
    PhaseIdentifier    string
    ProjectIdentifier  string
    PhaseName          string
    PhaseOrder         int
    PlannedStartDate   *time.Time
    PlannedEndDate     *time.Time
}

// CustomFieldDefinition describes a user-defined field on an element type.
type CustomFieldDefinition struct {
    FieldDefinitionIdentifier string
    ProjectIdentifier         string
    ApplicableElementType     string // element type or "*" for all
    FieldName                 string
    FieldType                 string // "string", "textarea", "integer", "real", "choice"
    FieldOptions              map[string]interface{}
    DisplayOrder              int
}

// Attachment references a file stored in the file storage backend.
type Attachment struct {
    AttachmentIdentifier   string
    ElementIdentifier      string
    FileStorageKey         string
    FileName               string
    FileSizeBytes          int64
    ContentType            string
    UploadTime             time.Time
    UploadedByIdentifier   string
}
```

---

## Store Interfaces

Each interface is small and focused. No God-interface. Implementations live in `store/postgres/`.

```go
package domain

import (
    "context"
    "io"
    "time"
)

// ElementStore handles CRUD for elements (current state).
type ElementStore interface {
    GetElementByIdentifier(context context.Context, elementIdentifier string) (*Element, error)
    ListElementsByProject(context context.Context, projectIdentifier string, filter ElementFilter) ([]Element, error)
    CreateElement(context context.Context, element *Element) error
    UpdateElement(context context.Context, element *Element) error
    DeleteElement(context context.Context, elementIdentifier string) error
    SearchElements(context context.Context, projectIdentifier string, query string) ([]Element, error)
}

// ElementVersionStore handles version history (cold path).
type ElementVersionStore interface {
    CreateVersion(context context.Context, version *ElementVersion, patch *ElementVersionPatch) error
    ListVersionsByElement(context context.Context, elementIdentifier string, limit int, offset int) ([]ElementVersion, error)
    GetPatchesInRange(context context.Context, elementIdentifier string, fromVersion int, toVersion int) ([]ElementVersionPatch, error)
}

// ElementPendingChangeStore manages the staging area for auto-commit.
type ElementPendingChangeStore interface {
    UpsertPendingChange(context context.Context, elementIdentifier string, snapshotBeforeEdits interface{}) error
    GetStalePendingChanges(context context.Context, inactivityThreshold time.Duration) ([]PendingChange, error)
    DeletePendingChange(context context.Context, elementIdentifier string) error
}

// ElementLinkStore handles links between elements.
type ElementLinkStore interface {
    CreateLink(context context.Context, link *ElementLink) error
    DeleteLink(context context.Context, linkIdentifier string) error
    ListLinksByElement(context context.Context, elementIdentifier string, direction string) ([]ElementLink, error)
}

// ProjectStore handles project CRUD.
type ProjectStore interface {
    GetProjectByIdentifier(context context.Context, projectIdentifier string) (*Project, error)
    ListProjects(context context.Context, folderPathPrefix string) ([]Project, error)
    CreateProject(context context.Context, project *Project) error
    UpdateProject(context context.Context, project *Project) error
    DeleteProject(context context.Context, projectIdentifier string) error
}

// AttachmentStore handles attachment metadata.
type AttachmentStore interface {
    CreateAttachment(context context.Context, attachment *Attachment) error
    ListAttachmentsByElement(context context.Context, elementIdentifier string) ([]Attachment, error)
    DeleteAttachment(context context.Context, attachmentIdentifier string) error
}

// FileStorage is the interface for the file blob backend (S3/SeaweedFS).
type FileStorage interface {
    PutFile(context context.Context, storageKey string, data io.Reader, contentType string) error
    GetFile(context context.Context, storageKey string) (io.ReadCloser, error)
    DeleteFile(context context.Context, storageKey string) error
    GeneratePresignedURL(context context.Context, storageKey string, expiration time.Duration) (string, error)
}

// CacheStore is the interface for the read cache layer.
type CacheStore interface {
    Get(context context.Context, cacheKey string) (interface{}, bool)
    Set(context context.Context, cacheKey string, value interface{}, timeToLive time.Duration)
    Invalidate(context context.Context, cacheKey string)
    InvalidateByPrefix(context context.Context, prefix string)
}

// --- Supporting types ---

// ElementFilter for list queries.
type ElementFilter struct {
    ElementTypes []ElementType
    TaskStatuses []TaskStatus
    AssigneeIdentifier *string
    SearchQuery        *string
    Limit              int
    Offset             int
}

// PendingChange is a row from the staging area.
type PendingChange struct {
    ElementIdentifier    string
    LastEditTime         time.Time
    SnapshotBeforeEdits  interface{}
}
```

---

## Auto-Commit Flow (detail)

```
┌─────────────────────────────────────────────────────────┐
│ User edits element via API                              │
│   1. Update `elements` table (hot, immediate)           │
│   2. Upsert `element_pending_changes`:                  │
│      - If no row exists: save current state as snapshot  │
│      - If row exists: just update `last_edit_time`      │
└──────────────┬──────────────────────────────────────────┘
               │
┌──────────────▼──────────────────────────────────────────┐
│ Auto-commit background worker (runs every 60s)          │
│   1. Query `element_pending_changes`                    │
│      WHERE last_edit_time < now() - inactivity_window   │
│   2. For each stale pending change:                     │
│      a. Read current state from `elements`              │
│      b. Diff against `snapshot_before_edits` → patches  │
│      c. Compute content_sha of current state            │
│      d. INSERT into `element_versions` + `_patches`     │
│      e. UPDATE `elements.content_sha`                   │
│      f. DELETE from `element_pending_changes`            │
└─────────────────────────────────────────────────────────┘
```

The inactivity window and poll interval are configurable in `config.yaml`:

```yaml
versioning:
  inactivity_window: 5m
  commit_poll_interval: 60s
```
