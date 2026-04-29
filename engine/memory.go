package engine

import "time"

type Memory struct {
	ActorID     ActorID           `json:"actor_id"`
	KnownEvents []Event           `json:"known_events,omitempty"`
	KnownRumors []Rumor           `json:"known_rumors,omitempty"`
	Attributes  map[string]string `json:"attributes,omitempty"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

func NewMemory(actorID ActorID) Memory {
	return Memory{
		ActorID:    actorID,
		Attributes: map[string]string{},
	}
}

func (m *Memory) RememberEvent(event Event) {
	m.KnownEvents = append(m.KnownEvents, cloneEvent(event))
	if event.At.After(m.UpdatedAt) {
		m.UpdatedAt = event.At
	}
}

func (m *Memory) RememberRumor(rumor Rumor) {
	m.KnownRumors = append(m.KnownRumors, cloneRumor(rumor))
	if rumor.UpdatedAt.After(m.UpdatedAt) {
		m.UpdatedAt = rumor.UpdatedAt
	}
}

func (w *World) Memory(actorID ActorID) Memory {
	memory, ok := w.Memories[actorID]
	if !ok {
		return NewMemory(actorID)
	}
	return cloneMemory(memory)
}

func (w *World) TeachEvent(actorID ActorID, event Event) {
	memory := w.Memory(actorID)
	memory.RememberEvent(event)
	w.Memories[actorID] = memory
}

func (w *World) TeachRumor(actorID ActorID, rumor Rumor) {
	memory := w.Memory(actorID)
	memory.RememberRumor(rumor)
	w.Memories[actorID] = memory
}
