package engine

import "sync"

// SafeWorld wraps a World with a read/write mutex for shared server access.
type SafeWorld struct {
	mu    sync.RWMutex
	world *World
}

// NewSafeWorld creates a concurrency-safe wrapper around a world pointer.
func NewSafeWorld(world *World) *SafeWorld {
	return &SafeWorld{world: world}
}

// View runs fn while holding a read lock.
func (s *SafeWorld) View(fn func(*World) error) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return fn(s.world)
}

// Update runs fn while holding a write lock.
func (s *SafeWorld) Update(fn func(*World) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return fn(s.world)
}

// Snapshot returns a world snapshot while holding a read lock.
func (s *SafeWorld) Snapshot() WorldSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.world.Snapshot()
}
