package monitor

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestAlertBatch_ConcurrentAdd(t *testing.T) {
	var count int64

	b := NewAlertBatch(200*time.Millisecond, 1000, func(batch map[int][]string) {
		for _, v := range batch {
			atomic.AddInt64(&count, int64(len(v)))
		}
	})

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(port int) {
			defer wg.Done()
			b.Add(port, "closed")
			b.Add(port, "open")
		}(8000 + i)
	}
	wg.Wait()
	b.Flush()
	time.Sleep(100 * time.Millisecond)

	got := atomic.LoadInt64(&count)
	if got != 100 {
		t.Fatalf("expected 100 total events, got %d", got)
	}
}

func TestAlertBatch_MultipleFlushesDoNotDuplicate(t *testing.T) {
	var mu sync.Mutex
	totalEvents := 0

	b := NewAlertBatch(500*time.Millisecond, 1000, func(batch map[int][]string) {
		mu.Lock()
		for _, v := range batch {
			totalEvents += len(v)
		}
		mu.Unlock()
	})

	b.Add(9090, "closed")
	b.Add(9091, "open")

	b.Flush()
	b.Flush() // second flush on empty bucket should be no-op

	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if totalEvents != 2 {
		t.Fatalf("expected exactly 2 events across all flushes, got %d", totalEvents)
	}
}
