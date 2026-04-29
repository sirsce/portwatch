package monitor

import (
	"sync"
	"testing"
	"time"
)

func TestNewAlertFilter_PanicsOnZero(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on zero cooldown")
		}
	}()
	NewAlertFilter(0)
}

func TestNewAlertFilter_PanicsOnNegative(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on negative cooldown")
		}
	}()
	NewAlertFilter(-5)
}

func TestAlertFilter_FirstCallAlwaysAllowed(t *testing.T) {
	af := NewAlertFilter(60)
	if !af.Allow(8080) {
		t.Fatal("first alert for a port should always be allowed")
	}
}

func TestAlertFilter_SecondCallWithinCooldownDenied(t *testing.T) {
	af := NewAlertFilter(60)
	af.Allow(8080)
	if af.Allow(8080) {
		t.Fatal("second alert within cooldown window should be suppressed")
	}
}

func TestAlertFilter_DifferentPortsAreIndependent(t *testing.T) {
	af := NewAlertFilter(60)
	if !af.Allow(8080) {
		t.Fatal("port 8080 first call should be allowed")
	}
	if !af.Allow(9090) {
		t.Fatal("port 9090 first call should be allowed independently")
	}
}

func TestAlertFilter_ResetAllowsNextCall(t *testing.T) {
	af := NewAlertFilter(60)
	af.Allow(8080) // consume the first free pass
	af.Reset(8080)
	if !af.Allow(8080) {
		t.Fatal("after Reset, next alert should be allowed")
	}
}

func TestAlertFilter_AllowsAfterCooldownExpires(t *testing.T) {
	// Use a 1-second cooldown and a real RateLimiter to exercise time logic.
	af := NewAlertFilter(1)
	af.Allow(3000)
	time.Sleep(1100 * time.Millisecond)
	if !af.Allow(3000) {
		t.Fatal("alert should be allowed after cooldown window expires")
	}
}

func TestAlertFilter_ConcurrentAccess(t *testing.T) {
	af := NewAlertFilter(60)
	var wg sync.WaitGroup
	ports := []int{80, 443, 8080, 9090, 3306}

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			port := ports[i%len(ports)]
			af.Allow(port)
		}(i)
	}
	wg.Wait() // should not race or panic
}
