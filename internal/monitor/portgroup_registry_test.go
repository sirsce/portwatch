package monitor

import (
	"testing"
)

func TestPortGroupRegistry_RegisterAndGet(t *testing.T) {
	r := NewPortGroupRegistry()
	g := NewPortGroup("web", []int{80, 443})
	if err := r.Register(g); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, ok := r.Get("web")
	if !ok {
		t.Fatal("expected group to be found")
	}
	if got.Name != "web" {
		t.Errorf("expected name web, got %s", got.Name)
	}
}

func TestPortGroupRegistry_DuplicateReturnsError(t *testing.T) {
	r := NewPortGroupRegistry()
	g := NewPortGroup("web", []int{80})
	_ = r.Register(g)
	err := r.Register(NewPortGroup("web", []int{8080}))
	if err == nil {
		t.Fatal("expected error for duplicate group name")
	}
}

func TestPortGroupRegistry_GetMissing(t *testing.T) {
	r := NewPortGroupRegistry()
	_, ok := r.Get("nonexistent")
	if ok {
		t.Fatal("expected not found")
	}
}

func TestPortGroupRegistry_Remove(t *testing.T) {
	r := NewPortGroupRegistry()
	_ = r.Register(NewPortGroup("db", []int{5432}))
	if !r.Remove("db") {
		t.Fatal("expected Remove to return true")
	}
	if r.Remove("db") {
		t.Fatal("expected Remove to return false for missing group")
	}
}

func TestPortGroupRegistry_All(t *testing.T) {
	r := NewPortGroupRegistry()
	_ = r.Register(NewPortGroup("web", []int{80}))
	_ = r.Register(NewPortGroup("db", []int{5432}))
	all := r.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(all))
	}
}

func TestPortGroupRegistry_Len(t *testing.T) {
	r := NewPortGroupRegistry()
	if r.Len() != 0 {
		t.Fatal("expected empty registry")
	}
	_ = r.Register(NewPortGroup("svc", []int{9090}))
	if r.Len() != 1 {
		t.Fatalf("expected Len 1, got %d", r.Len())
	}
}
