package engine

import "sync"

type SafeWorld struct {
	mu    sync.RWMutex
	world *World
}

func NewSafeWorld(world *World) *SafeWorld {
	return &SafeWorld{world: world}
}

func (s *SafeWorld) View(fn func(*World) error) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return fn(s.world)
}

func (s *SafeWorld) Update(fn func(*World) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return fn(s.world)
}

func (s *SafeWorld) Snapshot() WorldSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.world.Snapshot()
}
