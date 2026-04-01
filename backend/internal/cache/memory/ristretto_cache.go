package memory

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto"

	"github.com/jerome-godbout-lychens/project_works/backend/internal/domain"
)

// RistrettoCache implements domain.CacheStore using an in-process ristretto cache.
type RistrettoCache struct {
	cache      *ristretto.Cache
	timeToLive time.Duration

	// keyTracker tracks all active keys so we can implement prefix invalidation.
	// ristretto does not natively support prefix-based eviction.
	keyTracker   map[string]struct{}
	keyTrackerMu sync.RWMutex
}

// NewRistrettoCache creates a new in-process cache.
func NewRistrettoCache(maxCostBytes int64, timeToLive time.Duration) (domain.CacheStore, error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: maxCostBytes / 100, // 1% of max cost for frequency tracking
		MaxCost:     maxCostBytes,
		BufferItems: 64,
	})
	if err != nil {
		return nil, err
	}

	return &RistrettoCache{
		cache:      cache,
		timeToLive: timeToLive,
		keyTracker: make(map[string]struct{}),
	}, nil
}

func (cacheStore *RistrettoCache) Get(_ context.Context, cacheKey string) (interface{}, bool) {
	return cacheStore.cache.Get(cacheKey)
}

func (cacheStore *RistrettoCache) Set(_ context.Context, cacheKey string, value interface{}, timeToLive time.Duration) {
	if timeToLive == 0 {
		timeToLive = cacheStore.timeToLive
	}

	cacheStore.cache.SetWithTTL(cacheKey, value, 1, timeToLive)

	cacheStore.keyTrackerMu.Lock()
	cacheStore.keyTracker[cacheKey] = struct{}{}
	cacheStore.keyTrackerMu.Unlock()
}

func (cacheStore *RistrettoCache) Invalidate(_ context.Context, cacheKey string) {
	cacheStore.cache.Del(cacheKey)

	cacheStore.keyTrackerMu.Lock()
	delete(cacheStore.keyTracker, cacheKey)
	cacheStore.keyTrackerMu.Unlock()
}

func (cacheStore *RistrettoCache) InvalidateByPrefix(_ context.Context, prefix string) {
	cacheStore.keyTrackerMu.Lock()
	keysToDelete := make([]string, 0)
	for key := range cacheStore.keyTracker {
		if strings.HasPrefix(key, prefix) {
			keysToDelete = append(keysToDelete, key)
		}
	}
	for _, key := range keysToDelete {
		cacheStore.cache.Del(key)
		delete(cacheStore.keyTracker, key)
	}
	cacheStore.keyTrackerMu.Unlock()
}
