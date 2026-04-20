package memory

import (
	"context"
	"testing"
	"time"
)

// waitForCacheWrite gives ristretto's async buffer time to apply a Set.
// Ristretto batches writes through a buffer; a short sleep is the canonical
// way to get deterministic reads in tests.
func waitForCacheWrite() {
	time.Sleep(20 * time.Millisecond)
}

func newTestCache(t *testing.T) *RistrettoCache {
	t.Helper()
	store, err := NewRistrettoCache(10*1024*1024, time.Minute)
	if err != nil {
		t.Fatalf("failed to create cache: %v", err)
	}
	return store.(*RistrettoCache)
}

func TestRistrettoCache_SetAndGet(t *testing.T) {
	cache := newTestCache(t)
	ctx := context.Background()

	cache.Set(ctx, "greeting", "hello", 0)
	waitForCacheWrite()

	value, found := cache.Get(ctx, "greeting")
	if !found {
		t.Fatal("expected cached value to be found")
	}
	if value != "hello" {
		t.Errorf("expected 'hello', got %v", value)
	}
}

func TestRistrettoCache_GetMissReturnsFalse(t *testing.T) {
	cache := newTestCache(t)
	_, found := cache.Get(context.Background(), "missing")
	if found {
		t.Error("expected missing key to return false")
	}
}

func TestRistrettoCache_InvalidateRemovesKey(t *testing.T) {
	cache := newTestCache(t)
	ctx := context.Background()

	cache.Set(ctx, "k", "v", 0)
	waitForCacheWrite()
	cache.Invalidate(ctx, "k")

	_, found := cache.Get(ctx, "k")
	if found {
		t.Error("expected invalidated key to be gone")
	}
}

func TestRistrettoCache_InvalidateByPrefix(t *testing.T) {
	cache := newTestCache(t)
	ctx := context.Background()

	cache.Set(ctx, "element:1", "a", 0)
	cache.Set(ctx, "element:2", "b", 0)
	cache.Set(ctx, "project:1", "c", 0)
	waitForCacheWrite()

	cache.InvalidateByPrefix(ctx, "element:")

	if _, found := cache.Get(ctx, "element:1"); found {
		t.Error("expected element:1 to be invalidated")
	}
	if _, found := cache.Get(ctx, "element:2"); found {
		t.Error("expected element:2 to be invalidated")
	}
	if _, found := cache.Get(ctx, "project:1"); !found {
		t.Error("expected project:1 to survive prefix invalidation")
	}
}

func TestRistrettoCache_TTLExpires(t *testing.T) {
	cache := newTestCache(t)
	ctx := context.Background()

	cache.Set(ctx, "short-lived", "gone-soon", 50*time.Millisecond)
	waitForCacheWrite()

	if _, found := cache.Get(ctx, "short-lived"); !found {
		t.Fatal("expected value to be present immediately after Set")
	}

	time.Sleep(100 * time.Millisecond)
	if _, found := cache.Get(ctx, "short-lived"); found {
		t.Error("expected value to expire after TTL")
	}
}
