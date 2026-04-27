package monitor

import (
	"testing"
	"time"
)

func TestNewStateChangeLog_PanicsOnZero(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for max=0")
		}
	}()
	NewStateChangeLog(0)
}

func TestStateChangeLog_RecordAndAll(t *testing.T) {
	log := NewStateChangeLog(10)
	now := time.Now()

	log.Record(80, ChangeOpened, now)
	log.Record(443, ChangeClosed, now.Add(time.Second))

	changes := log.All()
	if len(changes) != 2 {
		t.Fatalf("expected 2 changes, got %d", len(changes))
	}
	if changes[0].Port != 80 || changes[0].Kind != ChangeOpened {
		t.Errorf("unexpected first change: %+v", changes[0])
	}
	if changes[1].Port != 443 || changes[1].Kind != ChangeClosed {
		t.Errorf("unexpected second change: %+v", changes[1])
	}
}

func TestStateChangeLog_Eviction(t *testing.T) {
	log := NewStateChangeLog(3)
	now := time.Now()

	for i := 0; i < 5; i++ {
		log.Record(i, ChangeOpened, now.Add(time.Duration(i)*time.Second))
	}

	if log.Len() != 3 {
		t.Fatalf("expected 3 entries after eviction, got %d", log.Len())
	}
	changes := log.All()
	// Oldest three entries should be ports 2, 3, 4.
	if changes[0].Port != 2 {
		t.Errorf("expected oldest retained port=2, got port=%d", changes[0].Port)
	}
	if changes[2].Port != 4 {
		t.Errorf("expected newest port=4, got port=%d", changes[2].Port)
	}
}

func TestStateChangeLog_AllReturnsCopy(t *testing.T) {
	log := NewStateChangeLog(5)
	log.Record(8080, ChangeOpened, time.Now())

	copy1 := log.All()
	copy1[0].Port = 9999

	copy2 := log.All()
	if copy2[0].Port == 9999 {
		t.Error("All() should return an independent copy")
	}
}

func TestStateChangeLog_Len(t *testing.T) {
	log := NewStateChangeLog(10)
	if log.Len() != 0 {
		t.Errorf("expected 0, got %d", log.Len())
	}
	log.Record(22, ChangeClosed, time.Now())
	if log.Len() != 1 {
		t.Errorf("expected 1, got %d", log.Len())
	}
}
