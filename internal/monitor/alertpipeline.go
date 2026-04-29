package monitor

import (
	"context"
	"fmt"

	"github.com/user/portwatch/internal/alert"
)

// AlertPipeline wires together an AlertFilter, AlertRouter, and Dispatcher
// to form a single send path: filter → route → dispatch.
type AlertPipeline struct {
	filter     *AlertFilter
	router     *AlertRouter
	dispatcher *alert.Dispatcher
}

// NewAlertPipeline creates an AlertPipeline.
// Panics if any argument is nil.
func NewAlertPipeline(f *AlertFilter, r *AlertRouter, d *alert.Dispatcher) *AlertPipeline {
	if f == nil {
		panic("alertpipeline: filter must not be nil")
	}
	if r == nil {
		panic("alertpipeline: router must not be nil")
	}
	if d == nil {
		panic("alertpipeline: dispatcher must not be nil")
	}
	return &AlertPipeline{filter: f, router: r, dispatcher: d}
}

// Send evaluates the pipeline for the given port and state change message.
// It returns false (and no error) when the alert is suppressed by the filter.
// It returns an error if no notifiers are registered for the port.
func (p *AlertPipeline) Send(ctx context.Context, port int, msg string) (bool, error) {
	if !p.filter.Allow(port) {
		return false, nil
	}

	notifiers := p.router.Resolve(port)
	if len(notifiers) == 0 {
		return false, fmt.Errorf("alertpipeline: no notifiers for port %d", port)
	}

	for _, n := range notifiers {
		p.dispatcher.Send(alert.Message{
			Notifier: n,
			Subject:  fmt.Sprintf("portwatch: port %d state change", port),
			Body:     msg,
		})
	}
	return true, nil
}
