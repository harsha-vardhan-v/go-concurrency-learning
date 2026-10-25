package ch09

import (
	"sync"
	"time"
)

// TODO(reader): This cache is correct. Run `go test -race ./...`
// and the race detector has nothing to say about it — every field
// access is under the lock, reads included.
//
// It is also serializing every read. §9.3 opens with exactly this
// picture: three lookups that could have run at once, queued behind
// one another because a sync.Mutex does not distinguish a reader from
// a writer.
//
// Swap the lock for the one that does. Two things to get right:
//
//  1. Reads take the read lock and writes take the write lock. Using
//     Lock for a read compiles, passes -race, and buys you nothing —
//     that is §9.3's "Mistake 3".
//  2. Do not try to upgrade. If you find yourself holding RLock and
//     wanting Lock, release first and re-check after acquiring, or you
//     get §9.3's lock-upgrade deadlock.
//
// The second test measures whether reads actually overlap, so a fix
// that keeps the exclusive lock will fail it even though nothing
// races.
type Cache struct {
	mu   sync.RWMutex
	data map[string]string
}

func New() *Cache {
	return &Cache{data: make(map[string]string)}
}

func (c *Cache) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}

// Get looks up a key. The sleep stands in for a read that costs
// something — §9.3's threshold for RWMutex paying its way.
func (c *Cache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	time.Sleep(2 * time.Millisecond)
	v, ok := c.data[key]
	return v, ok
}
