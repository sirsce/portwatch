package monitor

import (
	"context"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/user/portwatch/internal/config"
)

// fakeNotifier records every Notify call for assertions.
type fakeNotifier struct {
	mu   sync.Mutex
	calls []string
}

func (f *fakeNotifier) Notify(subject, body string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, subject)
	return nil
}

func (f *fakeNotifier) count() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.calls)
}

// startTCPServer opens a listener on a random port and returns the port number
// together with a closer function.
func startTCPServer(t *testing.T) (int, func()) {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("startTCPServer: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	return port, func() { ln.Close() }
}

func TestMonitor_DetectsPortClose(t *testing.T) {
	port, closeServer := startTCPServer(t)

	cfg := &config.Config{
		Host:         "127.0.0.1",
		Ports:        []int{port},
		ScanInterval: 50 * time.Millisecond,
		ScanTimeout:  200 * time.Millisecond,
	}

	m := New(cfg)
	notifier := &fakeNotifier{}
	m.notifiers = []Notifier{notifier}

	// First scan — establishes baseline, no alerts expected.
	m.scan()
	if notifier.count() != 0 {
		t.Fatalf("expected 0 alerts after baseline scan, got %d", notifier.count())
	}

	// Close the server so the port disappears.
	closeServer()
	time.Sleep(20 * time.Millisecond)

	// Second scan — port is now closed, should trigger one alert.
	m.scan()
	if notifier.count() != 1 {
		t.Fatalf("expected 1 alert after port closed, got %d", notifier.count())
	}
}

func TestMonitor_RunCancellation(t *testing.T) {
	cfg := &config.Config{
		Host:         "127.0.0.1",
		Ports:        []int{},
		ScanInterval: 10 * time.Millisecond,
		ScanTimeout:  100 * time.Millisecond,
	}

	m := New(cfg)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		m.Run(ctx)
		close(done)
	}()

	cancel()
	select {
	case <-done:
		// success
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Run did not return after context cancellation")
	}
}

func TestItoa(t *testing.T) {
	if got := itoa(8080); got != fmt.Sprintf("%d", 8080) {
		t.Errorf("itoa(8080) = %q, want \"8080\"", got)
	}
}
