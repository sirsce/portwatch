package monitor

import (
	"sync"
	"testing"
	"time"
)

// TestAlertBatch_IntegrationWithStateChangeLog verifies that AlertBatch
// correctly accumulates events produced by StateChangeLog transitions.
func TestAlertBatch_IntegrationWithStateChangeLog(t *testing.T) {
	log := NewStateChangeLog(20)

	var mu sync.Mutex
	var flushed []map[int][]string

	batch := NewAlertBatch(60*time.Millisecond, 50, func(b map[int][]string) {
		mu.Lock()
		flushed = append(flushed, b)
		mu.Unlock()
	})

	ports := []int{8080, 8081, 8082}
	states := []string{"open", "closed", "open"}

	for i, p := range ports {
		log.Record(p, states[i])
		batch.Add(p, states[i])
	}

	batch.Flush()
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(flushed) == 0 {
		t.Fatal("expected at least one flush")
	}

	merged := make(map[int][]string)
	for _, b := range flushed {
		for k, v := range b {
			merged[k] = append(merged[k], v...)
		}
	}

	if len(merged) != 3 {
		t.Fatalf("expected 3 ports in merged batch, got %d", len(merged))
	}

	all := log.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 state change entries, got %d", len(all))
	}
}
