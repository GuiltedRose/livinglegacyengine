package engine

// TargetKind classifies a structured event, rumor, or perception target.
type TargetKind string

const (
	// TargetActor identifies a generic actor target.
	TargetActor TargetKind = "actor"
	// TargetCharacter identifies a character target.
	TargetCharacter TargetKind = "character"
	// TargetItem identifies a crafted item target.
	TargetItem TargetKind = "item"
	// TargetDungeon identifies a loot dungeon target.
	TargetDungeon TargetKind = "dungeon"
	// TargetRegion identifies a host-game region target.
	TargetRegion TargetKind = "region"
	// TargetFaction identifies a host-game faction target.
	TargetFaction TargetKind = "faction"
	// TargetEvent identifies an event target.
	TargetEvent TargetKind = "event"
	// TargetCustom identifies a host-defined target.
	TargetCustom TargetKind = "custom"
)

// TargetRef is a structured reference to an engine or host-game object.
type TargetRef struct {
	ID         string            `json:"id"`
	Kind       TargetKind        `json:"kind"`
	Name       string            `json:"name,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

// NewTargetRef creates a target reference and defaults empty kinds to TargetCustom.
func NewTargetRef(id string, kind TargetKind, name string) TargetRef {
	if kind == "" {
		kind = TargetCustom
	}
	return TargetRef{
		ID:         id,
		Kind:       kind,
		Name:       name,
		Attributes: map[string]string{},
	}
}

// ActorRef creates a target reference for an actor.
func ActorRef(actor Actor) TargetRef {
	return NewTargetRef(string(actor.ID), TargetActor, actor.Name)
}

// CharacterRef creates a target reference for a character.
func CharacterRef(character Character) TargetRef {
	return NewTargetRef(string(character.ID), TargetCharacter, character.Name)
}

// ItemRef creates a target reference for a crafted item.
func ItemRef(item CraftedItem) TargetRef {
	return NewTargetRef(string(item.ID), TargetItem, item.Name)
}

// DungeonRef creates a target reference for a loot dungeon.
func DungeonRef(dungeon LootDungeon) TargetRef {
	return NewTargetRef(string(dungeon.ID), TargetDungeon, dungeon.Name)
}

// Empty reports whether the target has no ID.
func (t TargetRef) Empty() bool {
	return t.ID == ""
}

func (t TargetRef) normalized(kind TargetKind) TargetRef {
	if t.Kind == "" {
		t.Kind = kind
	}
	t.Attributes = cloneStringMap(t.Attributes)
	if t.Attributes == nil && t.ID != "" {
		t.Attributes = map[string]string{}
	}
	return t
}
