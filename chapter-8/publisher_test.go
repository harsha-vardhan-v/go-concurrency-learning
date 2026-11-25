package ch08

import (
	"testing"
	"time"
)

// The value handed over must be the value published. A torn handover
// shows up here as a zero: Wait saw done before it saw data.
func TestPublisherHandsOverTheValue(t *testing.T) {
	for round := 0; round < 200; round++ {
		p := &Publisher{}
		got := make(chan int, 1)
		go func() { got <- p.Wait() }()
		go p.Publish(42)

		select {
		case v := <-got:
			if v != 42 {
				t.Fatalf("round %d: Wait returned %d, want 42. "+
					"Wait saw done before it saw data — nothing "+
					"ordered the writes against the reads. See §8.4.",
					round, v)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("Wait never returned: the reader is still " +
				"spinning on a value it may never observe. §8.4.")
		}
	}
}

// Several publishers and readers at once. This is the shape the race
// detector needs to see the conflicting accesses.
func TestPublisherUnderConcurrentLoad(t *testing.T) {
	for round := 0; round < 100; round++ {
		p := &Publisher{}
		done := make(chan int, 4)
		for i := 0; i < 3; i++ {
			go func() { done <- p.Wait() }()
		}
		go p.Publish(7)

		deadline := time.After(2 * time.Second)
		for i := 0; i < 3; i++ {
			select {
			case v := <-done:
				if v != 7 {
					t.Fatalf("round %d: got %d, want 7", round, v)
				}
			case <-deadline:
				t.Fatal("a reader never returned. See §8.4.")
			}
		}
	}
}
