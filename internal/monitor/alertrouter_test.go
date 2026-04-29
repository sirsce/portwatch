package monitor

import (
	"testing"
)

func makeRegistry(t *testing.T, groups map[string][]int) *PortGroupRegistry {
	t.Helper()
	reg := NewPortGroupRegistry()
	for name, ports := range groups {
		pg := NewPortGroup(name, ports)
		if err := reg.Register(pg); err != nil {
			t.Fatalf("register group %q: %v", name, err)
		}
	}
	return reg
}

func TestNewAlertRouter_PanicsOnNilRegistry(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil registry")
		}
	}()
	NewAlertRouter(nil)
}

func TestAlertRouter_AddRoute_UnknownGroup(t *testing.T) {
	reg := makeRegistry(t, map[string][]int{"web": {80, 443}})
	ar := NewAlertRouter(reg)
	err := ar.AddRoute("nonexistent", []string{"slack"})
	if err == nil {
		t.Fatal("expected error for unknown group")
	}
}

func TestAlertRouter_AddRoute_EmptyNotifiers(t *testing.T) {
	reg := makeRegistry(t, map[string][]int{"web": {80, 443}})
	ar := NewAlertRouter(reg)
	err := ar.AddRoute("web", []string{})
	if err == nil {
		t.Fatal("expected error for empty notifier names")
	}
}

func TestAlertRouter_Resolve_MatchingGroup(t *testing.T) {
	reg := makeRegistry(t, map[string][]int{"web": {80, 443}})
	ar := NewAlertRouter(reg)
	if err := ar.AddRoute("web", []string{"slack", "email"}); err != nil {
		t.Fatalf("AddRoute: %v", err)
	}
	notifiers := ar.Resolve(80)
	if len(notifiers) != 2 {
		t.Fatalf("expected 2 notifiers, got %d", len(notifiers))
	}
}

func TestAlertRouter_Resolve_NoMatch(t *testing.T) {
	reg := makeRegistry(t, map[string][]int{"web": {80, 443}})
	ar := NewAlertRouter(reg)
	_ = ar.AddRoute("web", []string{"slack"})
	notifiers := ar.Resolve(9999)
	if len(notifiers) != 0 {
		t.Fatalf("expected 0 notifiers, got %d", len(notifiers))
	}
}

func TestAlertRouter_Resolve_DeduplicatesNotifiers(t *testing.T) {
	reg := makeRegistry(t, map[string][]int{
		"web": {80, 443},
		"all": {80, 8080},
	})
	ar := NewAlertRouter(reg)
	_ = ar.AddRoute("web", []string{"slack"})
	_ = ar.AddRoute("all", []string{"slack", "email"})
	notifiers := ar.Resolve(80)
	if len(notifiers) != 2 {
		t.Fatalf("expected 2 deduplicated notifiers, got %d: %v", len(notifiers), notifiers)
	}
}

func TestAlertRouter_Routes_ReturnsCopy(t *testing.T) {
	reg := makeRegistry(t, map[string][]int{"web": {80}})
	ar := NewAlertRouter(reg)
	_ = ar.AddRoute("web", []string{"email"})
	routes := ar.Routes()
	if len(routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(routes))
	}
	// Mutate returned slice — should not affect internal state
	routes[0].NotifierNames[0] = "mutated"
	original := ar.Routes()
	if original[0].NotifierNames[0] == "mutated" {
		t.Fatal("Routes() returned a reference instead of a copy")
	}
}
