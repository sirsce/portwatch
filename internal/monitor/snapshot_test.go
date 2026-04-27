package monitor

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func tempSnapshotPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "snapshot.json")
}

func TestSnapshotStore_SaveAndLoad(t *testing.T) {
	store := NewSnapshotStore(tempSnapshotPath(t))

	now := time.Now().Truncate(time.Second)
	orig := Snapshot{
		Timestamp: now,
		Ports: map[int]PortState{
			80:  {Open: true, ChangedAt: now},
			443: {Open: false, ChangedAt: now.Add(-time.Minute)},
		},
	}

	if err := store.Save(orig); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if !loaded.Timestamp.Equal(orig.Timestamp) {
		t.Errorf("Timestamp mismatch: got %v, want %v", loaded.Timestamp, orig.Timestamp)
	}
	if len(loaded.Ports) != len(orig.Ports) {
		t.Errorf("Ports length mismatch: got %d, want %d", len(loaded.Ports), len(orig.Ports))
	}
	if loaded.Ports[80].Open != true {
		t.Errorf("expected port 80 to be open")
	}
	if loaded.Ports[443].Open != false {
		t.Errorf("expected port 443 to be closed")
	}
}

func TestSnapshotStore_LoadMissingFile(t *testing.T) {
	store := NewSnapshotStore(filepath.Join(t.TempDir(), "nonexistent.json"))

	snap, err := store.Load()
	if err != nil {
		t.Fatalf("expected no error for missing file, got: %v", err)
	}
	if snap.Ports != nil {
		t.Errorf("expected nil Ports for empty snapshot, got: %v", snap.Ports)
	}
}

func TestSnapshotStore_OverwritesPreviousSave(t *testing.T) {
	path := tempSnapshotPath(t)
	store := NewSnapshotStore(path)

	first := Snapshot{Timestamp: time.Now(), Ports: map[int]PortState{22: {Open: true}}}
	if err := store.Save(first); err != nil {
		t.Fatalf("first Save: %v", err)
	}

	second := Snapshot{Timestamp: time.Now(), Ports: map[int]PortState{8080: {Open: false}}}
	if err := store.Save(second); err != nil {
		t.Fatalf("second Save: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if _, ok := loaded.Ports[22]; ok {
		t.Error("expected port 22 to be absent after overwrite")
	}
	if _, ok := loaded.Ports[8080]; !ok {
		t.Error("expected port 8080 to be present after overwrite")
	}
}

func TestSnapshotStore_LoadCorruptFile(t *testing.T) {
	path := tempSnapshotPath(t)
	if err := os.WriteFile(path, []byte("not-json{"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	store := NewSnapshotStore(path)
	_, err := store.Load()
	if err == nil {
		t.Error("expected error for corrupt JSON, got nil")
	}
}
