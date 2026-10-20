package ch08

import "sync"

// TODO(reader): Publisher hands one value from a producer to a
// consumer goroutine. It works, most of the time, on most machines
// — which is the whole problem.
//
// `done` is a plain bool and `data` is a plain int, and nothing in
// this file establishes a happens-before relationship between the two
// writes in Publish and the two reads in Wait. §8.4 lists everything
// that would: a channel, a mutex, sync.Once, a WaitGroup, an atomic.
// This file uses none of them.
//
// Two consequences, and the tests below check for both:
//
//  1. `go test -race` reports a data race. That gate does not care
//     how lucky your timing was.
//  2. Wait can observe `done == true` while `data` is still 0,
//     because nothing orders the two writes against the two reads.
//
// Fix it with any mechanism from §8.4 that gives you a happens-before
// edge. Do not add a sleep: sleeping is not synchronization (§8.4,
// "No Guarantee: Time-Based Ordering"), and the race detector will
// still flag it.
type Publisher struct {
	data  int
	ready chan struct{}
	init  sync.Once
	once  sync.Once
}

func (p *Publisher) Signal() chan struct{} {
	p.init.Do(func() {
		p.ready = make(chan struct{})
	})

	return p.ready
}

func (p *Publisher) Publish(v int) {
	ready := p.Signal()

	p.once.Do(func() {
		p.data = v
		close(ready)
	})
}

// Wait blocks until Publish has been called, then returns the value.
func (p *Publisher) Wait() int {
	<-p.Signal()
	return p.data
}
