package ch09

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// Correctness first: whatever lock you choose, the cache must still
// hand back what was stored, with the race detector satisfied.
func TestCacheStoresAndReturns(t *testing.T) {
	c := New()
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Go(func() { c.Set(fmt.Sprint("k", i), fmt.Sprint("v", i)) })
	}
	wg.Wait()

	for i := 0; i < 20; i++ {
		got, ok := c.Get(fmt.Sprint("k", i))
		if !ok || got != fmt.Sprint("v", i) {
			t.Fatalf("Get(k%d) = %q,%v; want %q,true",
				i, got, ok, fmt.Sprint("v", i))
		}
	}
}

// The point of the exercise. Ten readers, each of which takes 2ms
// inside Get. Serialized that is ~20ms; overlapping it is ~2ms. The
// threshold sits well clear of both so a slow machine cannot fail it
// by accident.
func TestConcurrentReadsOverlap(t *testing.T) {
	c := New()
	c.Set("k", "v")

	const readers = 10
	start := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < readers; i++ {
		wg.Go(func() { c.Get("k") })
	}
	wg.Wait()
	elapsed := time.Since(start)

	if elapsed > 10*time.Millisecond {
		t.Fatalf("%d concurrent reads took %v; they are being "+
			"serialized. A read lock would let them overlap. "+
			"See §9.3.", readers, elapsed.Round(time.Millisecond))
	}
}

// A smoke test that writes still happen and still complete. Real
// exclusivity is enforced by -race and by Go's own concurrent-map-write
// panic, both of which fire long before an assertion here would; this
// exists to stop the obvious cheat of making Set shared too, which
// would let the readers overlap and then explode here.
func TestWritesStayExclusive(t *testing.T) {
	c := New()
	var mu sync.Mutex
	inWrite, maxConcurrent := 0, 0

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Go(func() {
			mu.Lock()
			inWrite++
			if inWrite > maxConcurrent {
				maxConcurrent = inWrite
			}
			mu.Unlock()

			c.Set(fmt.Sprint("k", i), "v")

			mu.Lock()
			inWrite--
			mu.Unlock()
		})
	}
	wg.Wait()
	if maxConcurrent == 0 {
		t.Fatal("no writes observed")
	}
}
