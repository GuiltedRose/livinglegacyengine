package engine

import "fmt"

type ActorKind string

const (
	ActorCharacter    ActorKind = "character"
	ActorCrafter      ActorKind = "crafter"
	ActorFaction      ActorKind = "faction"
	ActorSettlement   ActorKind = "settlement"
	ActorOrganization ActorKind = "organization"
	ActorUnknown      ActorKind = "unknown"
)

type Actor struct {
	ID         ActorID           `json:"id"`
	Name       string            `json:"name"`
	Kind       ActorKind         `json:"kind"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

func NewActor(id ActorID, name string, kind ActorKind) (Actor, error) {
	if id == "" {
		return Actor{}, fmt.Errorf("actor id is required")
	}
	if name == "" {
		return Actor{}, fmt.Errorf("actor name is required")
	}
	if kind == "" {
		kind = ActorUnknown
	}
	return Actor{
		ID:         id,
		Name:       name,
		Kind:       kind,
		Attributes: map[string]string{},
	}, nil
}

func (w *World) AddActor(actor Actor) error {
	if actor.ID == "" {
		return fmt.Errorf("actor id is required")
	}
	if actor.Name == "" {
		return fmt.Errorf("actor name is required")
	}
	if actor.Kind == "" {
		actor.Kind = ActorUnknown
	}
	actor.Attributes = cloneStringMap(actor.Attributes)
	if actor.Attributes == nil {
		actor.Attributes = map[string]string{}
	}
	w.Actors[actor.ID] = actor
	if _, ok := w.Memories[actor.ID]; !ok {
		w.Memories[actor.ID] = NewMemory(actor.ID)
	}
	return nil
}

func (w *World) Actor(id ActorID) (Actor, bool) {
	actor, ok := w.Actors[id]
	if !ok {
		return Actor{}, false
	}
	actor.Attributes = cloneStringMap(actor.Attributes)
	return actor, true
}
