package monitor

import (
	"sync"
	"testing"
	"time"
)

func TestNewAlertBatch_PanicsOnZeroWindow(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on zero window")
		}
	}()
	NewAlertBatch(0, 10, func(map[int][]string) {})
}

func TestNewAlertBatch_PanicsOnZeroMaxSize(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on zero maxSize")
		}
	}()
	NewAlertBatch(time.Second, 0, func(map[int][]string) {})
}

func TestNewAlertBatch_PanicsOnNilFlushFn(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on nil flushFn")
		}
	}()
	NewAlertBatch(time.Second, 10, nil)
}

func TestAlertBatch_FlushOnMaxSize(t *testing.T) {
	var mu sync.Mutex
	var received map[int][]string

	b := NewAlertBatch(5*time.Second, 3, func(batch map[int][]string) {
		mu.Lock()
		received = batch
		mu.Unlock()
	})

	b.Add(8080, "closed")
	b.Add(8081, "closed")
	b.Add(8082, "closed") // triggers flush

	time.Sleep(50 * time.Millisecond)
	mu.Lock()
	defer mu.Unlock()

	if len(received) != 3 {
		t.Fatalf("expected 3 ports in batch, got %d", len(received))
	}
}

func TestAlertBatch_FlushOnWindowExpiry(t *testing.T) {
	var mu sync.Mutex
	var received map[int][]string

	b := NewAlertBatch(50*time.Millisecond, 100, func(batch map[int][]string) {
		mu.Lock()
		received = batch
		mu.Unlock()
	})

	b.Add(9000, "open")
	b.Add(9000, "closed")

	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(received[9000]) != 2 {
		t.Fatalf("expected 2 events for port 9000, got %d", len(received[9000]))
	}
}

func TestAlertBatch_ManualFlush(t *testing.T) {
	var mu sync.Mutex
	var received map[int][]string

	b := NewAlertBatch(10*time.Second, 100, func(batch map[int][]string) {
		mu.Lock()
		received = batch
		mu.Unlock()
	})

	b.Add(443, "closed")
	b.Flush()

	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(received[443]) != 1 {
		t.Fatalf("expected 1 event for port 443, got %v", received)
	}
}

func TestAlertBatch_FlushEmptyIsNoOp(t *testing.T) {
	called := false
	b := NewAlertBatch(time.Second, 10, func(batch map[int][]string) {
		called = true
	})
	b.Flush()
	time.Sleep(30 * time.Millisecond)
	if called {
		t.Fatal("flushFn should not be called on empty batch")
	}
}
