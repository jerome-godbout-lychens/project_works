package domain

import (
	"context"
	"time"
)

// ElementVersionStore handles version history operations (cold path).
type ElementVersionStore interface {
	CreateVersion(context context.Context, version *ElementVersion, patch *ElementVersionPatch) error
	ListVersionsByElement(context context.Context, elementIdentifier string, limit int, offset int) ([]ElementVersion, error)
	GetPatchesInRange(context context.Context, elementIdentifier string, fromVersionNumber int, toVersionNumber int) ([]ElementVersionPatch, error)
}

// ElementPendingChangeStore manages the staging area for auto-commit.
type ElementPendingChangeStore interface {
	UpsertPendingChange(context context.Context, elementIdentifier string, snapshotBeforeEdits map[string]interface{}) error
	GetStalePendingChanges(context context.Context, inactivityThreshold time.Duration) ([]PendingChange, error)
	DeletePendingChange(context context.Context, elementIdentifier string) error
}
