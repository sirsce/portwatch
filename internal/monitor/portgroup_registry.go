package monitor

import (
	"fmt"
	"sync"
)

// PortGroupRegistry manages a collection of named port groups.
type PortGroupRegistry struct {
	mu     sync.RWMutex
	groups map[string]*PortGroup
}

// NewPortGroupRegistry returns an empty PortGroupRegistry.
func NewPortGroupRegistry() *PortGroupRegistry {
	return &PortGroupRegistry{
		groups: make(map[string]*PortGroup),
	}
}

// Register adds a PortGroup to the registry.
// Returns an error if a group with the same name already exists.
func (r *PortGroupRegistry) Register(g *PortGroup) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.groups[g.Name]; exists {
		return fmt.Errorf("portgroup registry: group %q already registered", g.Name)
	}
	r.groups[g.Name] = g
	return nil
}

// Get retrieves a PortGroup by name. Returns nil and false if not found.
func (r *PortGroupRegistry) Get(name string) (*PortGroup, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	g, ok := r.groups[name]
	return g, ok
}

// All returns a slice of all registered PortGroups.
func (r *PortGroupRegistry) All() []*PortGroup {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*PortGroup, 0, len(r.groups))
	for _, g := range r.groups {
		out = append(out, g)
	}
	return out
}

// Remove deletes a group by name. Returns false if not found.
func (r *PortGroupRegistry) Remove(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.groups[name]; !ok {
		return false
	}
	delete(r.groups, name)
	return true
}

// Len returns the number of registered groups.
func (r *PortGroupRegistry) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.groups)
}
