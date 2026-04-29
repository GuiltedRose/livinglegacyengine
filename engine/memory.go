package engine

import "time"

// Memory stores what an actor knows about events and rumors.
type Memory struct {
	ActorID     ActorID           `json:"actor_id"`
	KnownEvents []Event           `json:"known_events,omitempty"`
	KnownRumors []Rumor           `json:"known_rumors,omitempty"`
	Attributes  map[string]string `json:"attributes,omitempty"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// NewMemory creates an empty memory record for an actor.
func NewMemory(actorID ActorID) Memory {
	return Memory{
		ActorID:    actorID,
		Attributes: map[string]string{},
	}
}

// RememberEvent appends a cloned event to memory.
func (m *Memory) RememberEvent(event Event) {
	m.KnownEvents = append(m.KnownEvents, cloneEvent(event))
	if event.At.After(m.UpdatedAt) {
		m.UpdatedAt = event.At
	}
}

// RememberRumor appends a cloned rumor to memory.
func (m *Memory) RememberRumor(rumor Rumor) {
	m.KnownRumors = append(m.KnownRumors, cloneRumor(rumor))
	if rumor.UpdatedAt.After(m.UpdatedAt) {
		m.UpdatedAt = rumor.UpdatedAt
	}
}

// Memory returns an actor's memory or an empty memory if none exists.
func (w *World) Memory(actorID ActorID) Memory {
	memory, ok := w.Memories[actorID]
	if !ok {
		return NewMemory(actorID)
	}
	return cloneMemory(memory)
}

// TeachEvent records an event in an actor's memory.
func (w *World) TeachEvent(actorID ActorID, event Event) {
	memory := w.Memory(actorID)
	memory.RememberEvent(event)
	w.Memories[actorID] = memory
}

// TeachRumor records a rumor in an actor's memory.
func (w *World) TeachRumor(actorID ActorID, rumor Rumor) {
	memory := w.Memory(actorID)
	memory.RememberRumor(rumor)
	w.Memories[actorID] = memory
}
