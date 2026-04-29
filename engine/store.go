package engine

import (
	"os"
	"path/filepath"
)

// SnapshotStore persists and loads world snapshots.
type SnapshotStore interface {
	Save(snapshot WorldSnapshot) error
	Load() (WorldSnapshot, error)
}

// FileSnapshotStore saves snapshots as JSON at a filesystem path.
type FileSnapshotStore struct {
	Path string
}

// NewFileSnapshotStore creates a file-backed snapshot store.
func NewFileSnapshotStore(path string) FileSnapshotStore {
	return FileSnapshotStore{Path: path}
}

// Save writes a snapshot as JSON, creating parent directories as needed.
func (s FileSnapshotStore) Save(snapshot WorldSnapshot) error {
	data, err := MarshalSnapshot(snapshot)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(s.Path, data, 0o644)
}

// Load reads and decodes a snapshot from disk.
func (s FileSnapshotStore) Load() (WorldSnapshot, error) {
	data, err := os.ReadFile(s.Path)
	if err != nil {
		return WorldSnapshot{}, err
	}
	return UnmarshalSnapshot(data)
}
