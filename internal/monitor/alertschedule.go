package monitor

import (
	"fmt"
	"sync"
	"time"
)

// AlertSchedule suppresses alerts outside of defined active time windows.
// Windows are defined as [start, end) in wall-clock hours (0–23).
type AlertSchedule struct {
	mu      sync.RWMutex
	windows []timeWindow
	nowFn   func() time.Time
}

type timeWindow struct {
	startHour int
	endHour   int
}

// NewAlertSchedule creates an AlertSchedule with no active windows.
// By default all alerts are suppressed until at least one window is added.
func NewAlertSchedule() *AlertSchedule {
	return &AlertSchedule{nowFn: time.Now}
}

// AddWindow registers an active hour window [startHour, endHour).
// Both values must be in [0, 23] and startHour must be less than endHour.
func (s *AlertSchedule) AddWindow(startHour, endHour int) error {
	if startHour < 0 || startHour > 23 {
		return fmt.Errorf("alertschedule: startHour %d out of range [0,23]", startHour)
	}
	if endHour < 1 || endHour > 24 {
		return fmt.Errorf("alertschedule: endHour %d out of range [1,24]", endHour)
	}
	if startHour >= endHour {
		return fmt.Errorf("alertschedule: startHour %d must be less than endHour %d", startHour, endHour)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.windows = append(s.windows, timeWindow{startHour, endHour})
	return nil
}

// IsActive reports whether the current wall-clock hour falls within any
// registered window. Returns false when no windows have been added.
func (s *AlertSchedule) IsActive() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.windows) == 0 {
		return false
	}
	hour := s.nowFn().Hour()
	for _, w := range s.windows {
		if hour >= w.startHour && hour < w.endHour {
			return true
		}
	}
	return false
}

// ClearWindows removes all registered windows.
func (s *AlertSchedule) ClearWindows() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.windows = nil
}
