package monitor

import (
	"context"
	"time"
)

// Ticker wraps a time.Ticker and provides a channel-based interface
// for driving periodic scan cycles in the monitor loop.
type Ticker struct {
	ticker   *time.Ticker
	C        <-chan time.Time
	interval time.Duration
}

// NewTicker creates a new Ticker that fires at the given interval.
// Panics if interval is zero or negative.
func NewTicker(interval time.Duration) *Ticker {
	if interval <= 0 {
		panic("monitor: ticker interval must be positive")
	}
	t := time.NewTicker(interval)
	return &Ticker{
		ticker:   t,
		C:        t.C,
		interval: interval,
	}
}

// Stop stops the ticker, releasing associated resources.
func (t *Ticker) Stop() {
	t.ticker.Stop()
}

// Interval returns the configured tick interval.
func (t *Ticker) Interval() time.Duration {
	return t.interval
}

// RunEvery calls fn on each tick until ctx is cancelled.
// It returns when the context is done.
func RunEvery(ctx context.Context, interval time.Duration, fn func(ctx context.Context)) {
	tk := NewTicker(interval)
	defer tk.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-tk.C:
			fn(ctx)
		}
	}
}
