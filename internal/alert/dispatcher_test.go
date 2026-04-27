package alert

import (
	"context"
	"sync"
	"testing"
	"time"
)

// fakeNotifier records every Notify call.
type fakeNotifier struct {
	mu       sync.Mutex
	messages []Message
	errOnce  error
}

func (f *fakeNotifier) Notify(subject, body string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.errOnce != nil {
		err := f.errOnce
		f.errOnce = nil
		return err
	}
	f.messages = append(f.messages, Message{Subject: subject, Body: body})
	return nil
}

func (f *fakeNotifier) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.messages)
}

func TestDispatcher_SendAndReceive(t *testing.T) {
	n := &fakeNotifier{}
	d := NewDispatcher(8, n)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go d.Run(ctx)

	if err := d.Send("port closed", "port 80 is down"); err != nil {
		t.Fatalf("Send returned unexpected error: %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if n.count() == 1 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Errorf("expected 1 notification, got %d", n.count())
}

func TestDispatcher_QueueFull(t *testing.T) {
	n := &fakeNotifier{}
	d := NewDispatcher(1, n) // buffer of 1, no runner

	if err := d.Send("first", "body"); err != nil {
		t.Fatalf("first Send failed: %v", err)
	}
	if err := d.Send("overflow", "body"); err == nil {
		t.Error("expected error when queue is full, got nil")
	}
}

func TestDispatcher_DrainOnCancel(t *testing.T) {
	n := &fakeNotifier{}
	d := NewDispatcher(16, n)

	for i := 0; i < 5; i++ {
		_ = d.Send("subject", "body")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately so Run drains then exits
	d.Run(ctx)

	if n.count() != 5 {
		t.Errorf("expected 5 drained notifications, got %d", n.count())
	}
}
