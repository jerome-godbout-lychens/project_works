package domain

import (
	"context"
	"time"
)

// CacheStore is the interface for the read cache layer.
// Implementations can use in-process caches (ristretto) or distributed caches (Redis).
type CacheStore interface {
	Get(context context.Context, cacheKey string) (interface{}, bool)
	Set(context context.Context, cacheKey string, value interface{}, timeToLive time.Duration)
	Invalidate(context context.Context, cacheKey string)
	InvalidateByPrefix(context context.Context, prefix string)
}
