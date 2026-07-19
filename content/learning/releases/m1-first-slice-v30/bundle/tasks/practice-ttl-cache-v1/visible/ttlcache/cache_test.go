package ttlcache

import (
	"testing"
	"time"
)

func TestCacheDistinguishesHitMissNegativeAndExpiry(t *testing.T) {
	now := time.Unix(100, 0)
	cache, err := New[string](func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	if _, _, hit := cache.Get("missing"); hit {
		t.Fatal("cold key was a hit")
	}
	if err := cache.Set("project", "gopher", true, time.Minute); err != nil {
		t.Fatal(err)
	}
	if value, found, hit := cache.Get("project"); value != "gopher" || !found || !hit {
		t.Fatalf("fresh=(%q,%v,%v)", value, found, hit)
	}
	if err := cache.Set("absent", "", false, 10*time.Second); err != nil {
		t.Fatal(err)
	}
	if _, found, hit := cache.Get("absent"); found || !hit {
		t.Fatalf("negative=(%v,%v)", found, hit)
	}
	now = now.Add(time.Minute)
	if _, _, hit := cache.Get("project"); hit {
		t.Fatal("expired entry was a hit")
	}
}

func TestCacheValidatesWritesAndDeletes(t *testing.T) {
	cache, _ := New[int](time.Now)
	if cache.Set("", 1, true, time.Second) == nil || cache.Set("key", 1, true, 0) == nil {
		t.Fatal("invalid entry accepted")
	}
	if err := cache.Set("key", 1, true, time.Second); err != nil {
		t.Fatal(err)
	}
	cache.Delete("key")
	if _, _, hit := cache.Get("key"); hit {
		t.Fatal("deleted key remained")
	}
}
