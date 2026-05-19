package badges

import (
	"sync"
	"time"
)

type cacheEntry struct {
	value  any
	expiry time.Time
}

type cache struct {
	m   map[string]cacheEntry
	mu  sync.RWMutex
	ttl time.Duration
}

func newCache(ttl time.Duration) *cache {
	return &cache{m: make(map[string]cacheEntry), ttl: ttl}
}

func (c *cache) Get(name string) (any, bool) {
	// Disabled cache when ttl == 0
	if c.ttl == 0 {
		return nil, false
	}
	c.mu.RLock()
	ce, ok := c.m[name]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}
	if time.Now().After(ce.expiry) {
		// expired: treat as miss
		return nil, false
	}
	return ce.value, true
}

func (c *cache) Set(name string, value any) {
	// Disabled mode: no-op
	if c.ttl == 0 {
		return
	}
	c.mu.Lock()
	c.m[name] = cacheEntry{value: value, expiry: time.Now().Add(c.ttl)}
	c.mu.Unlock()
}

func (c *cache) Delete(name string) {
	if c.ttl == 0 {
		return
	}
	c.mu.Lock()
	delete(c.m, name)
	c.mu.Unlock()
}
