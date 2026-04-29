package engine

import (
	"os"
	"path/filepath"
)

type SnapshotStore interface {
	Save(snapshot WorldSnapshot) error
	Load() (WorldSnapshot, error)
}

type FileSnapshotStore struct {
	Path string
}

func NewFileSnapshotStore(path string) FileSnapshotStore {
	return FileSnapshotStore{Path: path}
}

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

func (s FileSnapshotStore) Load() (WorldSnapshot, error) {
	data, err := os.ReadFile(s.Path)
	if err != nil {
		return WorldSnapshot{}, err
	}
	return UnmarshalSnapshot(data)
}
