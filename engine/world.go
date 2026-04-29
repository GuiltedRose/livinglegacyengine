package engine

import (
	"fmt"
	"time"
)

type EventType string

const (
	EventItemCrafted    EventType = "item_crafted"
	EventCharacterDied  EventType = "character_died"
	EventDungeonBorn    EventType = "dungeon_born"
	EventDungeonLooted  EventType = "dungeon_looted"
	EventDungeonDecayed EventType = "dungeon_decayed"
	EventRespawned      EventType = "character_respawned"
)

type Event struct {
	Type        EventType         `json:"type"`
	At          time.Time         `json:"at"`
	Description string            `json:"description"`
	SubjectID   string            `json:"subject_id,omitempty"`
	ObjectID    string            `json:"object_id,omitempty"`
	Subject     TargetRef         `json:"subject,omitempty"`
	Object      TargetRef         `json:"object,omitempty"`
	Data        map[string]string `json:"data,omitempty"`
}

type World struct {
	Character   Character
	Dungeons    map[DungeonID]LootDungeon
	Crafters    map[ActorID]int
	Actors      map[ActorID]Actor
	Rumors      []Rumor
	Memories    map[ActorID]Memory
	Perceptions map[PerceptionKey]Perception
	Events      []Event

	now       func() time.Time
	nextRunID int
	rules     Rules
	indices   worldIndices
}

func NewWorld(character Character) *World {
	return NewWorldWithRules(character, DefaultRules())
}

func NewWorldWithRules(character Character, rules Rules) *World {
	return &World{
		Character:   character,
		Dungeons:    make(map[DungeonID]LootDungeon),
		Crafters:    make(map[ActorID]int),
		Actors:      make(map[ActorID]Actor),
		Memories:    make(map[ActorID]Memory),
		Perceptions: make(map[PerceptionKey]Perception),
		now:         time.Now,
		rules:       rules.normalized(),
		indices:     newWorldIndices(),
	}
}

func (w *World) SetClock(now func() time.Time) {
	if now != nil {
		w.now = now
	}
}

func (w *World) SetRules(rules Rules) {
	w.rules = rules.normalized()
}

func (w *World) AddCraftedItem(item CraftedItem) error {
	itemValue := w.rules.ItemLegacyValue(item)
	if itemValue < 0 {
		return fmt.Errorf("item %q has negative legacy value", item.ID)
	}
	if err := w.Character.Carry(item); err != nil {
		return err
	}
	w.Crafters[item.CrafterID] += itemValue
	w.record(Event{
		Type:        EventItemCrafted,
		Description: fmt.Sprintf("%s carries %s, crafted by %s", w.Character.Name, item.Name, item.CrafterID),
		SubjectID:   string(item.CrafterID),
		ObjectID:    string(item.ID),
		Subject:     NewTargetRef(string(item.CrafterID), TargetActor, ""),
		Object:      ItemRef(item),
		Data: map[string]string{
			"character_id": string(w.Character.ID),
			"item_name":    item.Name,
		},
	})
	return nil
}

func (w *World) KillCharacter(cause string) (LootDungeon, error) {
	if !w.Character.Alive {
		return LootDungeon{}, fmt.Errorf("%s is already dead", w.Character.Name)
	}
	if cause == "" {
		cause = "unknown danger"
	}

	nextDeath := w.Character.Deaths + 1
	w.nextRunID++
	dungeonID := DungeonID(fmt.Sprintf("dungeon-%04d", w.nextRunID))

	dungeonCharacter := w.Character
	dungeonCharacter.Deaths = nextDeath
	dungeon, err := NewLootDungeonWithRules(dungeonID, dungeonCharacter, w.now(), w.rules)
	if err != nil {
		w.nextRunID--
		return LootDungeon{}, err
	}

	w.Character.Alive = false
	w.Character.Deaths = nextDeath
	w.Dungeons[dungeon.ID] = dungeon
	w.record(Event{
		Type:        EventCharacterDied,
		Description: fmt.Sprintf("%s died to %s", w.Character.Name, cause),
		SubjectID:   string(w.Character.ID),
		Subject:     CharacterRef(w.Character),
		Data: map[string]string{
			"cause": cause,
		},
	})
	w.record(Event{
		Type:        EventDungeonBorn,
		Description: fmt.Sprintf("%s became a loot dungeon worth %d legacy", dungeon.Name, dungeon.LegacyValue),
		SubjectID:   string(w.Character.ID),
		ObjectID:    string(dungeon.ID),
		Subject:     CharacterRef(w.Character),
		Object:      DungeonRef(dungeon),
		Data: map[string]string{
			"dungeon_name": dungeon.Name,
		},
	})
	return dungeon, nil
}

func (w *World) RespawnCharacter() {
	w.Character.Respawn()
	w.record(Event{
		Type:        EventRespawned,
		Description: fmt.Sprintf("%s returns for another run", w.Character.Name),
		SubjectID:   string(w.Character.ID),
		Subject:     CharacterRef(w.Character),
	})
}

func (w *World) LootDungeon(id DungeonID, looter ActorID) ([]CraftedItem, error) {
	if looter == "" {
		return nil, fmt.Errorf("looter id is required")
	}
	dungeon, ok := w.Dungeons[id]
	if !ok {
		return nil, fmt.Errorf("dungeon %q does not exist", id)
	}
	if dungeon.Status != DungeonSealed {
		return nil, fmt.Errorf("dungeon %q is %s", id, dungeon.Status)
	}

	dungeon.Status = DungeonLooted
	dungeon.LootedBy = looter
	w.Dungeons[id] = dungeon
	w.Character.LegacyScore += dungeon.LegacyValue
	w.Character.RecoveredItems = append(w.Character.RecoveredItems, dungeon.Items...)
	w.record(Event{
		Type:        EventDungeonLooted,
		Description: fmt.Sprintf("%s looted %s; legacy increased by %d", looter, dungeon.Name, dungeon.LegacyValue),
		SubjectID:   string(looter),
		ObjectID:    string(dungeon.ID),
		Subject:     NewTargetRef(string(looter), TargetActor, ""),
		Object:      DungeonRef(dungeon),
		Data: map[string]string{
			"dungeon_name": dungeon.Name,
		},
	})
	return append([]CraftedItem(nil), dungeon.Items...), nil
}

func (w *World) DecayDungeons() []DungeonID {
	now := w.now()
	decayed := []DungeonID{}
	for id, dungeon := range w.Dungeons {
		if dungeon.Status != DungeonSealed {
			continue
		}
		if !w.rules.ShouldDecayDungeon(dungeon, now) {
			continue
		}
		dungeon.Status = DungeonDecayed
		w.Dungeons[id] = dungeon
		decayed = append(decayed, id)
		w.record(Event{
			Type:        EventDungeonDecayed,
			Description: fmt.Sprintf("%s decayed", dungeon.Name),
			ObjectID:    string(dungeon.ID),
			Object:      DungeonRef(dungeon),
			Data: map[string]string{
				"dungeon_name": dungeon.Name,
			},
		})
	}
	return decayed
}

func (w *World) record(event Event) {
	event = w.normalizeEvent(event)
	w.Events = append(w.Events, event)
	w.indexEvent(len(w.Events)-1, event)
}

func (w *World) normalizeEvent(event Event) Event {
	if event.At.IsZero() {
		event.At = w.now()
	}
	event.Subject = event.Subject.normalized(TargetCustom)
	event.Object = event.Object.normalized(TargetCustom)
	event.Data = cloneStringMap(event.Data)
	return event
}
