package monitor_test

import (
	"context"
	"testing"
	"time"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/monitor"
)

// fakeNotifier records the last message it received.
type fakeNotifier struct {
	lastSubject string
	lastBody    string
	calls       int
}

func (f *fakeNotifier) Notify(subject, body string) error {
	f.calls++
	f.lastSubject = subject
	f.lastBody = body
	return nil
}

func buildPipeline(t *testing.T, ports []int) (*monitor.AlertPipeline, *fakeNotifier, *alert.Dispatcher) {
	t.Helper()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	disp := alert.NewDispatcher(8)
	go disp.Run(ctx)

	notifier := &fakeNotifier{}

	registry := monitor.NewPortGroupRegistry()
	_ = registry.Register(monitor.NewPortGroup("web", ports))

	router := monitor.NewAlertRouter(registry)
	_ = router.AddRoute("web", []alert.Notifier{notifier})

	filter := monitor.NewAlertFilter(500*time.Millisecond, 16)

	return monitor.NewAlertPipeline(filter, router, disp), notifier, disp
}

func TestAlertPipeline_SendDispatches(t *testing.T) {
	pipeline, notifier, _ := buildPipeline(t, []int{8080})

	sent, err := pipeline.Send(context.Background(), 8080, "port closed")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !sent {
		t.Fatal("expected alert to be sent")
	}

	time.Sleep(50 * time.Millisecond)
	if notifier.calls != 1 {
		t.Fatalf("expected 1 call, got %d", notifier.calls)
	}
	if notifier.lastBody != "port closed" {
		t.Errorf("unexpected body: %q", notifier.lastBody)
	}
}

func TestAlertPipeline_FilterSuppresses(t *testing.T) {
	pipeline, notifier, _ := buildPipeline(t, []int{9090})

	pipeline.Send(context.Background(), 9090, "first") //nolint:errcheck
	sent, err := pipeline.Send(context.Background(), 9090, "second")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sent {
		t.Fatal("expected second alert to be suppressed")
	}

	time.Sleep(50 * time.Millisecond)
	if notifier.calls != 1 {
		t.Errorf("expected 1 call after suppression, got %d", notifier.calls)
	}
}

func TestAlertPipeline_NoNotifiersReturnsError(t *testing.T) {
	pipeline, _, _ := buildPipeline(t, []int{8080})

	_, err := pipeline.Send(context.Background(), 9999, "msg")
	if err == nil {
		t.Fatal("expected error for unregistered port")
	}
}

func TestNewAlertPipeline_PanicsOnNilFilter(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on nil filter")
		}
	}()
	monitor.NewAlertPipeline(nil, &monitor.AlertRouter{}, &alert.Dispatcher{})
}
