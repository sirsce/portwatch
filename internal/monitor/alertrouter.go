package monitor

import (
	"fmt"
	"sync"
)

// AlertRoute maps a port group name to a list of notifier names.
type AlertRoute struct {
	GroupName     string
	NotifierNames []string
}

// AlertRouter routes alerts for a given port to the appropriate notifier names
// based on registered port group memberships.
type AlertRouter struct {
	mu       sync.RWMutex
	routes   []AlertRoute
	registry *PortGroupRegistry
}

// NewAlertRouter creates an AlertRouter backed by the given PortGroupRegistry.
// Panics if registry is nil.
func NewAlertRouter(registry *PortGroupRegistry) *AlertRouter {
	if registry == nil {
		panic("alertrouter: registry must not be nil")
	}
	return &AlertRouter{
		registry: registry,
	}
}

// AddRoute registers a mapping from a port group name to one or more notifier names.
// Returns an error if the group does not exist in the registry.
func (r *AlertRouter) AddRoute(groupName string, notifierNames []string) error {
	if _, ok := r.registry.Get(groupName); !ok {
		return fmt.Errorf("alertrouter: group %q not found in registry", groupName)
	}
	if len(notifierNames) == 0 {
		return fmt.Errorf("alertrouter: notifierNames must not be empty")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.routes = append(r.routes, AlertRoute{
		GroupName:     groupName,
		NotifierNames: append([]string(nil), notifierNames...),
	})
	return nil
}

// Resolve returns the deduplicated list of notifier names that should receive
// an alert for the given port number. If no routes match, an empty slice is returned.
func (r *AlertRouter) Resolve(port int) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	seen := make(map[string]struct{})
	var result []string

	for _, route := range r.routes {
		group, ok := r.registry.Get(route.GroupName)
		if !ok {
			continue
		}
		if group.Contains(port) {
			for _, n := range route.NotifierNames {
				if _, dup := seen[n]; !dup {
					seen[n] = struct{}{}
					result = append(result, n)
				}
			}
		}
	}
	return result
}

// Routes returns a copy of all registered routes.
func (r *AlertRouter) Routes() []AlertRoute {
	r.mu.RLock()
	defer r.mu.RUnlock()
	copy := make([]AlertRoute, len(r.routes))
	for i, rt := range r.routes {
		copy[i] = AlertRoute{
			GroupName:     rt.GroupName,
			NotifierNames: append([]string(nil), rt.NotifierNames...),
		}
	}
	return copy
}
