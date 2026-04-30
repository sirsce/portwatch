package monitor

import (
	"sync"
	"time"
)

// AlertBatch accumulates alert events within a time window and flushes them
// as a single batched notification, reducing alert noise during flapping.
type AlertBatch struct {
	mu       sync.Mutex
	window   time.Duration
	maxSize  int
	bucket   map[int][]string // port -> list of state labels
	timer    *time.Timer
	flushFn  func(batch map[int][]string)
}

// NewAlertBatch creates an AlertBatch that collects events for the given
// window duration before calling flushFn with the accumulated batch.
// Panics if window <= 0, maxSize <= 0, or flushFn is nil.
func NewAlertBatch(window time.Duration, maxSize int, flushFn func(map[int][]string)) *AlertBatch {
	if window <= 0 {
		panic("alertbatch: window must be positive")
	}
	if maxSize <= 0 {
		panic("alertbatch: maxSize must be positive")
	}
	if flushFn == nil {
		panic("alertbatch: flushFn must not be nil")
	}
	return &AlertBatch{
		window:  window,
		maxSize: maxSize,
		bucket:  make(map[int][]string),
		flushFn: flushFn,
	}
}

// Add records a state label for the given port. If the batch reaches maxSize
// it is flushed immediately; otherwise a flush is scheduled after the window.
func (b *AlertBatch) Add(port int, state string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.bucket[port] = append(b.bucket[port], state)

	total := 0
	for _, v := range b.bucket {
		total += len(v)
	}

	if total >= b.maxSize {
		if b.timer != nil {
			b.timer.Stop()
			b.timer = nil
		}
		b.flush()
		return
	}

	if b.timer == nil {
		b.timer = time.AfterFunc(b.window, func() {
			b.mu.Lock()
			defer b.mu.Unlock()
			b.flush()
		})
	}
}

// flush sends the current bucket to flushFn and resets state. Must be called
// with b.mu held.
func (b *AlertBatch) flush() {
	if len(b.bucket) == 0 {
		return
	}
	copy := make(map[int][]string, len(b.bucket))
	for k, v := range b.bucket {
		copy[k] = v
	}
	b.bucket = make(map[int][]string)
	b.timer = nil
	go b.flushFn(copy)
}

// Flush forces an immediate flush of any pending events.
func (b *AlertBatch) Flush() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.timer != nil {
		b.timer.Stop()
		b.timer = nil
	}
	b.flush()
}
