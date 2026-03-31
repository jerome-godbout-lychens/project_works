package domain

import "time"

// ElementVersion is the metadata for one committed version of an element.
type ElementVersion struct {
	VersionId     string
	ElementId     string
	VersionNumber         int
	ContentSha            string
	CommittedTime         time.Time
	CommittedById string
	CommitMessage         string
}

// ElementVersionPatch holds the diff between two consecutive versions.
// Both forward and reverse patches are stored to avoid recomputing inverses.
type ElementVersionPatch struct {
	VersionId string
	ForwardPatch      []PatchOperation // transforms version N-1 into version N
	ReversePatch      []PatchOperation // transforms version N into version N-1
}

// PatchOperation is one step in a JSON Patch (RFC 6902).
type PatchOperation struct {
	Operation string      `json:"op"`    // "add", "remove", "replace", "move", "copy", "test"
	Path      string      `json:"path"`  // JSON Pointer (RFC 6901)
	Value     interface{} `json:"value,omitempty"`
	From      string      `json:"from,omitempty"`
}

// PendingChange represents an element that has uncommitted edits.
// The auto-commit background worker uses this to detect idle elements
// and create version snapshots.
type PendingChange struct {
	ElementId   string
	LastEditTime        time.Time
	SnapshotBeforeEdits map[string]interface{} // serialized element state at last commit
}
