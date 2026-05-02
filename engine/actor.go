package engine

import "fmt"

// ActorKind classifies an Actor without forcing a host game into one taxonomy.
type ActorKind string

const (
	// ActorNPC represents an engine-owned non-player character.
	ActorNPC ActorKind = "npc"
	// ActorCharacter represents a character-like actor.
	ActorCharacter ActorKind = "character"
	// ActorCrafter represents an actor primarily known for creating items.
	ActorCrafter ActorKind = "crafter"
	// ActorFaction represents a faction, guild, polity, or similar group.
	ActorFaction ActorKind = "faction"
	// ActorSettlement represents a town, city, camp, or other place-as-actor.
	ActorSettlement ActorKind = "settlement"
	// ActorOrganization represents an organization that is not necessarily a faction.
	ActorOrganization ActorKind = "organization"
	// ActorUnknown is used when no more specific kind is supplied.
	ActorUnknown ActorKind = "unknown"
)

// Actor is a generic world entity that can craft, spread rumors, hold memory,
// or be used by a host game as a source or subject of events.
type Actor struct {
	ID         ActorID           `json:"id"`
	Name       string            `json:"name"`
	Kind       ActorKind         `json:"kind"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

// NewActor validates and creates an Actor with an initialized Attributes map.
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

// AddActor registers an actor and initializes an empty memory record for it.
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

// Actor returns a clone of the registered actor and whether it was found.
func (w *World) Actor(id ActorID) (Actor, bool) {
	actor, ok := w.Actors[id]
	if !ok {
		return Actor{}, false
	}
	actor.Attributes = cloneStringMap(actor.Attributes)
	return actor, true
}
