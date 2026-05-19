package badges

import (
	"sync"
	"testing"
	"time"
)

func TestCache_SetGetWithinTTL(t *testing.T) {
	c := newCache(200 * time.Millisecond)
	c.Set("k1", 123)
	if v, ok := c.Get("k1"); !ok {
		t.Fatalf("expected cache hit")
	} else if v.(int) != 123 {
		t.Fatalf("unexpected value: %v", v)
	}
}

func TestCache_Expiry(t *testing.T) {
	c := newCache(50 * time.Millisecond)
	c.Set("k2", "v2")
	// immediately present
	if _, ok := c.Get("k2"); !ok {
		t.Fatalf("expected immediate hit")
	}
	// wait beyond TTL
	time.Sleep(70 * time.Millisecond)
	if _, ok := c.Get("k2"); ok {
		t.Fatalf("expected miss after expiry")
	}
}

func TestCache_Delete(t *testing.T) {
	c := newCache(time.Second)
	c.Set("k3", true)
	c.Delete("k3")
	if _, ok := c.Get("k3"); ok {
		t.Fatalf("expected miss after delete")
	}
}

func TestCache_ConcurrentAccess(t *testing.T) {
	c := newCache(500 * time.Millisecond)
	const N = 50
	wg := sync.WaitGroup{}
	for i := range N {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := "k"
			c.Set(key, i)
			c.Get(key)
		}(i)
	}
	wg.Wait()
}

func TestCache_DisabledTTLZero(t *testing.T) {
	c := newCache(0)
	c.Set("k", 1)
	if _, ok := c.Get("k"); ok {
		t.Fatalf("expected miss when cache disabled")
	}
	c.Delete("k") // should be no-op and not panic
}
