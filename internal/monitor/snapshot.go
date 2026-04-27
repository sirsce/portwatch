package monitor

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// Snapshot holds a point-in-time view of all monitored port states.
type Snapshot struct {
	Timestamp time.Time        `json:"timestamp"`
	Ports     map[int]PortState `json:"ports"`
}

// PortState records whether a port is open and the last time its state changed.
type PortState struct {
	Open      bool      `json:"open"`
	ChangedAt time.Time `json:"changed_at"`
}

// SnapshotStore persists and retrieves the latest port snapshot to/from disk.
type SnapshotStore struct {
	mu   sync.RWMutex
	path string
}

// NewSnapshotStore creates a SnapshotStore that writes to the given file path.
func NewSnapshotStore(path string) *SnapshotStore {
	return &SnapshotStore{path: path}
}

// Save writes the snapshot to disk as JSON, replacing any previous file.
func (s *SnapshotStore) Save(snap Snapshot) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

// Load reads the most recently saved snapshot from disk.
// Returns a zero-value Snapshot and no error when the file does not yet exist.
func (s *SnapshotStore) Load() (Snapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(s.path)
	if os.IsNotExist(err) {
		return Snapshot{}, nil
	}
	if err != nil {
		return Snapshot{}, err
	}

	var snap Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return Snapshot{}, err
	}
	return snap, nil
}
