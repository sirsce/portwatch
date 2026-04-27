package monitor

import (
	"testing"
)

func TestHistory_NoChangeOnFirstRecord(t *testing.T) {
	h := NewHistory(10)
	changed := h.Record(8080, true)
	if changed {
		t.Error("first record should never be reported as a state change")
	}
	if len(h.Changes()) != 0 {
		t.Errorf("expected 0 changes, got %d", len(h.Changes()))
	}
}

func TestHistory_DetectsTransition(t *testing.T) {
	h := NewHistory(10)
	h.Record(8080, true)
	changed := h.Record(8080, false)
	if !changed {
		t.Error("expected state change to be detected")
	}
	changes := h.Changes()
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	if changes[0].WasOpen != true || changes[0].NowOpen != false {
		t.Errorf("unexpected change values: %+v", changes[0])
	}
}

func TestHistory_NoChangeWhenStateUnchanged(t *testing.T) {
	h := NewHistory(10)
	h.Record(9090, false)
	changed := h.Record(9090, false)
	if changed {
		t.Error("same state should not be reported as a change")
	}
	if len(h.Changes()) != 0 {
		t.Errorf("expected 0 changes, got %d", len(h.Changes()))
	}
}

func TestHistory_Latest(t *testing.T) {
	h := NewHistory(10)
	_, ok := h.Latest(1234)
	if ok {
		t.Error("expected no entry for unseen port")
	}
	h.Record(1234, true)
	status, ok := h.Latest(1234)
	if !ok {
		t.Fatal("expected entry after record")
	}
	if !status.Open || status.Port != 1234 {
		t.Errorf("unexpected status: %+v", status)
	}
}

func TestHistory_MaxChangesEviction(t *testing.T) {
	h := NewHistory(3)
	h.Record(80, true)
	for i := 0; i < 5; i++ {
		h.Record(80, i%2 == 0) // alternate open/closed
	}
	changes := h.Changes()
	if len(changes) > 3 {
		t.Errorf("expected at most 3 changes retained, got %d", len(changes))
	}
}

func TestHistory_ChangesReturnsCopy(t *testing.T) {
	h := NewHistory(10)
	h.Record(443, true)
	h.Record(443, false)
	c1 := h.Changes()
	c1[0].Port = 9999
	c2 := h.Changes()
	if c2[0].Port == 9999 {
		t.Error("Changes() should return an independent copy")
	}
}
