package engine

type TargetKind string

const (
	TargetActor     TargetKind = "actor"
	TargetCharacter TargetKind = "character"
	TargetItem      TargetKind = "item"
	TargetDungeon   TargetKind = "dungeon"
	TargetRegion    TargetKind = "region"
	TargetFaction   TargetKind = "faction"
	TargetEvent     TargetKind = "event"
	TargetCustom    TargetKind = "custom"
)

type TargetRef struct {
	ID         string            `json:"id"`
	Kind       TargetKind        `json:"kind"`
	Name       string            `json:"name,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

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

func ActorRef(actor Actor) TargetRef {
	return NewTargetRef(string(actor.ID), TargetActor, actor.Name)
}

func CharacterRef(character Character) TargetRef {
	return NewTargetRef(string(character.ID), TargetCharacter, character.Name)
}

func ItemRef(item CraftedItem) TargetRef {
	return NewTargetRef(string(item.ID), TargetItem, item.Name)
}

func DungeonRef(dungeon LootDungeon) TargetRef {
	return NewTargetRef(string(dungeon.ID), TargetDungeon, dungeon.Name)
}

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
