// Package alert provides notification mechanisms for portwatch.
package alert

import (
	"context"
	"fmt"
	"log"
	"sync"
)

// Notifier is the interface implemented by all alert backends.
type Notifier interface {
	Notify(subject, body string) error
}

// Dispatcher fans out alert messages to one or more Notifier backends.
type Dispatcher struct {
	mu        sync.RWMutex
	notifiers []Notifier
	queue     chan Message
}

// Message holds the data for a single alert event.
type Message struct {
	Subject string
	Body    string
}

// NewDispatcher creates a Dispatcher with the given notifiers and an internal
// queue of the specified buffer size.
func NewDispatcher(bufSize int, notifiers ...Notifier) *Dispatcher {
	return &Dispatcher{
		notifiers: notifiers,
		queue:     make(chan Message, bufSize),
	}
}

// Send enqueues a message for dispatch. It is non-blocking; if the queue is
// full the message is dropped and an error is returned.
func (d *Dispatcher) Send(subject, body string) error {
	select {
	case d.queue <- Message{Subject: subject, Body: body}:
		return nil
	default:
		return fmt.Errorf("alert queue full, message dropped: %s", subject)
	}
}

// Run processes queued messages until ctx is cancelled.
func (d *Dispatcher) Run(ctx context.Context) {
	for {
		select {
		case msg := <-d.queue:
			d.dispatch(msg)
		case <-ctx.Done():
			// Drain remaining messages before exiting.
			for {
				select {
				case msg := <-d.queue:
					d.dispatch(msg)
				default:
					return
				}
			}
		}
	}
}

func (d *Dispatcher) dispatch(msg Message) {
	d.mu.RLock()
	notifiers := d.notifiers
	d.mu.RUnlock()
	for _, n := range notifiers {
		if err := n.Notify(msg.Subject, msg.Body); err != nil {
			log.Printf("alert dispatcher: notifier error: %v", err)
		}
	}
}
