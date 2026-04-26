package alert_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/portwatch/internal/alert"
)

func TestWebhookNotifier_Notify_Success(t *testing.T) {
	var received alert.Event

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", ct)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	notifier := alert.NewWebhookNotifier(server.URL)
	event := alert.Event{
		Host:      "localhost",
		Port:      9090,
		Status:    "closed",
		Timestamp: time.Now().UTC().Truncate(time.Second),
		Message:   "port 9090 became unreachable",
	}

	if err := notifier.Notify(event); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if received.Port != event.Port {
		t.Errorf("expected port %d, got %d", event.Port, received.Port)
	}
	if received.Status != event.Status {
		t.Errorf("expected status %q, got %q", event.Status, received.Status)
	}
}

func TestWebhookNotifier_Notify_Non2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	notifier := alert.NewWebhookNotifier(server.URL)
	event := alert.Event{
		Host:      "localhost",
		Port:      80,
		Status:    "open",
		Timestamp: time.Now(),
		Message:   "port 80 is open",
	}

	if err := notifier.Notify(event); err == nil {
		t.Fatal("expected error for non-2xx response, got nil")
	}
}

func TestWebhookNotifier_Notify_InvalidURL(t *testing.T) {
	notifier := alert.NewWebhookNotifier("http://127.0.0.1:0/no-server")
	event := alert.Event{
		Host:      "localhost",
		Port:      443,
		Status:    "closed",
		Timestamp: time.Now(),
		Message:   "port 443 unreachable",
	}

	if err := notifier.Notify(event); err == nil {
		t.Fatal("expected error when server is unreachable, got nil")
	}
}
