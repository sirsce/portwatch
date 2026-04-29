package monitor

import (
	"strings"
	"testing"
)

func TestNewPortGroup_Valid(t *testing.T) {
	g := NewPortGroup("web", []int{80, 443})
	if g.Name != "web" {
		t.Fatalf("expected name web, got %s", g.Name)
	}
	if g.Len() != 2 {
		t.Fatalf("expected 2 ports, got %d", g.Len())
	}
}

func TestNewPortGroup_PanicsOnEmptyName(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for empty name")
		}
	}()
	NewPortGroup("", []int{80})
}

func TestNewPortGroup_PanicsOnEmptyPorts(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for empty ports")
		}
	}()
	NewPortGroup("web", []int{})
}

func TestNewPortGroup_PanicsOnInvalidPort(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for invalid port")
		}
	}()
	NewPortGroup("bad", []int{0})
}

func TestPortGroup_Contains(t *testing.T) {
	g := NewPortGroup("db", []int{5432, 3306})
	if !g.Contains(5432) {
		t.Error("expected Contains(5432) = true")
	}
	if g.Contains(9999) {
		t.Error("expected Contains(9999) = false")
	}
}

func TestPortGroup_String(t *testing.T) {
	g := NewPortGroup("svc", []int{8080})
	s := g.String()
	if !strings.Contains(s, "svc") {
		t.Errorf("String() missing name: %s", s)
	}
}

func TestNewPortGroup_IsolatesMutation(t *testing.T) {
	ports := []int{80, 443}
	g := NewPortGroup("web", ports)
	ports[0] = 9999
	if g.Ports[0] != 80 {
		t.Error("PortGroup should not be affected by external slice mutation")
	}
}
