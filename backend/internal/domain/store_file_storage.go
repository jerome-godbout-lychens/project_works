package domain

import (
	"context"
	"io"
	"time"
)

// FileStorage is the interface for the file blob backend (S3-compatible).
// Implementations can target SeaweedFS, AWS S3, or any S3-compatible storage.
type FileStorage interface {
	PutFile(context context.Context, storageKey string, data io.Reader, contentType string) error
	GetFile(context context.Context, storageKey string) (io.ReadCloser, error)
	DeleteFile(context context.Context, storageKey string) error
	GeneratePresignedURL(context context.Context, storageKey string, expiration time.Duration) (string, error)
}
