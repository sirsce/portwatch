package monitor_test

import (
	"context"
	"testing"
	"time"

	"github.com/user/portwatch/internal/alert"
	"github.com/user/portwatch/internal/monitor"
)

// TestAlertPipeline_IntegrationWithStateChangeLog verifies that a pipeline
// correctly dispatches alerts for transitions recorded in a StateChangeLog.
func TestAlertPipeline_IntegrationWithStateChangeLog(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	notifier := &fakeNotifier{}

	registry := monitor.NewPortGroupRegistry()
	_ = registry.Register(monitor.NewPortGroup("api", []int{3000, 3001}))

	router := monitor.NewAlertRouter(registry)
	_ = router.AddRoute("api", []alert.Notifier{notifier})

	disp := alert.NewDispatcher(16)
	go disp.Run(ctx)

	filter := monitor.NewAlertFilter(10*time.Millisecond, 16)
	pipeline := monitor.NewAlertPipeline(filter, router, disp)

	log := monitor.NewStateChangeLog(32)
	history := monitor.NewHistory(8)

	// Simulate port 3000 going from open → closed.
	history.Record(3000, true)
	changes := history.Record(3000, false)
	for _, ch := range changes {
		log.Record(ch)
		pipeline.Send(ctx, ch.Port, "port "+itoa(ch.Port)+" is now closed") //nolint:errcheck
	}

	time.Sleep(60 * time.Millisecond)

	if notifier.calls != 1 {
		t.Fatalf("expected 1 alert, got %d", notifier.calls)
	}
	if log.Len() != 1 {
		t.Errorf("expected 1 state change in log, got %d", log.Len())
	}
}
