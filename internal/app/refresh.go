package app

import (
	"sync"
	"time"
)

type AutoRefresher struct {
	mu      sync.Mutex
	stopCh  chan struct{}
	running bool
}

func (r *AutoRefresher) Start(interval time.Duration, fn func()) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.running {
		close(r.stopCh)
	}
	r.stopCh = make(chan struct{})
	r.running = true
	go func(stop <-chan struct{}) {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				fn()
			case <-stop:
				return
			}
		}
	}(r.stopCh)
}

func (r *AutoRefresher) Stop() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.running {
		close(r.stopCh)
		r.running = false
	}
}
