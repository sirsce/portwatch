package monitor

import (
	"testing"
	"time"
)

func fixedHour(hour int) func() time.Time {
	return func() time.Time {
		return time.Date(2024, 1, 1, hour, 30, 0, 0, time.UTC)
	}
}

func TestAlertSchedule_NoWindowsReturnsFalse(t *testing.T) {
	s := NewAlertSchedule()
	s.nowFn = fixedHour(10)
	if s.IsActive() {
		t.Error("expected inactive when no windows registered")
	}
}

func TestAlertSchedule_HourInsideWindow(t *testing.T) {
	s := NewAlertSchedule()
	s.nowFn = fixedHour(14)
	if err := s.AddWindow(9, 17); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !s.IsActive() {
		t.Error("expected active at hour 14 within [9,17)")
	}
}

func TestAlertSchedule_HourOutsideWindow(t *testing.T) {
	s := NewAlertSchedule()
	s.nowFn = fixedHour(8)
	_ = s.AddWindow(9, 17)
	if s.IsActive() {
		t.Error("expected inactive at hour 8 outside [9,17)")
	}
}

func TestAlertSchedule_WindowEndIsExclusive(t *testing.T) {
	s := NewAlertSchedule()
	s.nowFn = fixedHour(17)
	_ = s.AddWindow(9, 17)
	if s.IsActive() {
		t.Error("expected inactive at exact end hour (exclusive)")
	}
}

func TestAlertSchedule_MultipleWindows(t *testing.T) {
	s := NewAlertSchedule()
	_ = s.AddWindow(8, 12)
	_ = s.AddWindow(18, 22)

	s.nowFn = fixedHour(10)
	if !s.IsActive() {
		t.Error("expected active at hour 10")
	}
	s.nowFn = fixedHour(20)
	if !s.IsActive() {
		t.Error("expected active at hour 20")
	}
	s.nowFn = fixedHour(15)
	if s.IsActive() {
		t.Error("expected inactive at hour 15")
	}
}

func TestAlertSchedule_AddWindow_InvalidRange(t *testing.T) {
	s := NewAlertSchedule()
	if err := s.AddWindow(10, 10); err == nil {
		t.Error("expected error for equal start and end")
	}
	if err := s.AddWindow(15, 9); err == nil {
		t.Error("expected error for start > end")
	}
	if err := s.AddWindow(-1, 10); err == nil {
		t.Error("expected error for negative startHour")
	}
}

func TestAlertSchedule_ClearWindows(t *testing.T) {
	s := NewAlertSchedule()
	s.nowFn = fixedHour(10)
	_ = s.AddWindow(8, 18)
	if !s.IsActive() {
		t.Fatal("expected active before clear")
	}
	s.ClearWindows()
	if s.IsActive() {
		t.Error("expected inactive after ClearWindows")
	}
}
