package monitor

import (
	"testing"
)

// TestPortGroupRegistry_IntegrationWithPortGroup verifies that the registry
// correctly stores and retrieves PortGroups and that Contains works end-to-end.
func TestPortGroupRegistry_IntegrationWithPortGroup(t *testing.T) {
	reg := NewPortGroupRegistry()

	webGroup := NewPortGroup("web", []int{80, 443, 8080})
	dbGroup := NewPortGroup("database", []int{5432, 3306, 27017})

	if err := reg.Register(webGroup); err != nil {
		t.Fatalf("register web: %v", err)
	}
	if err := reg.Register(dbGroup); err != nil {
		t.Fatalf("register db: %v", err)
	}

	if reg.Len() != 2 {
		t.Fatalf("expected 2 groups, got %d", reg.Len())
	}

	got, ok := reg.Get("web")
	if !ok {
		t.Fatal("web group not found")
	}
	if !got.Contains(443) {
		t.Error("web group should contain port 443")
	}
	if got.Contains(5432) {
		t.Error("web group should not contain port 5432")
	}

	// Remove and verify isolation
	reg.Remove("database")
	if reg.Len() != 1 {
		t.Fatalf("expected 1 group after remove, got %d", reg.Len())
	}
	if _, ok := reg.Get("database"); ok {
		t.Error("database group should have been removed")
	}

	// Re-register with same name should now succeed
	newDB := NewPortGroup("database", []int{5432})
	if err := reg.Register(newDB); err != nil {
		t.Fatalf("re-register db after removal: %v", err)
	}
}
