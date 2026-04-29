package monitor

import "fmt"

// PortGroup represents a named collection of ports to monitor together.
type PortGroup struct {
	Name  string
	Ports []int
}

// NewPortGroup creates a new PortGroup with the given name and ports.
// It panics if name is empty or ports slice is empty.
func NewPortGroup(name string, ports []int) *PortGroup {
	if name == "" {
		panic("portgroup: name must not be empty")
	}
	if len(ports) == 0 {
		panic("portgroup: ports must not be empty")
	}
	for _, p := range ports {
		if p < 1 || p > 65535 {
			panic(fmt.Sprintf("portgroup: invalid port %d", p))
		}
	}
	copy := make([]int, len(ports))
	for i, p := range ports {
		copy[i] = p
	}
	return &PortGroup{Name: name, Ports: copy}
}

// Contains reports whether the group contains the given port.
func (g *PortGroup) Contains(port int) bool {
	for _, p := range g.Ports {
		if p == port {
			return true
		}
	}
	return false
}

// Len returns the number of ports in the group.
func (g *PortGroup) Len() int {
	return len(g.Ports)
}

// String returns a human-readable representation of the group.
func (g *PortGroup) String() string {
	return fmt.Sprintf("PortGroup(%s, %d ports)", g.Name, len(g.Ports))
}
