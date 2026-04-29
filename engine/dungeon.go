package engine

import (
	"fmt"
	"time"
)

type DungeonID string

type DungeonStatus string

const (
	DungeonSealed  DungeonStatus = "sealed"
	DungeonLooted  DungeonStatus = "looted"
	DungeonDecayed DungeonStatus = "decayed"
)

type LootDungeon struct {
	ID          DungeonID         `json:"id"`
	Name        string            `json:"name"`
	OriginRun   int               `json:"origin_run"`
	Depth       int               `json:"depth"`
	CreatedAt   time.Time         `json:"created_at"`
	Status      DungeonStatus     `json:"status"`
	Items       []CraftedItem     `json:"items,omitempty"`
	LegacyValue int               `json:"legacy_value"`
	LootedBy    ActorID           `json:"looted_by,omitempty"`
	Attributes  map[string]string `json:"attributes,omitempty"`
}

func NewLootDungeon(id DungeonID, character Character, now time.Time) (LootDungeon, error) {
	return NewLootDungeonWithRules(id, character, now, DefaultRules())
}

func NewLootDungeonWithRules(id DungeonID, character Character, now time.Time, rules Rules) (LootDungeon, error) {
	rules = rules.normalized()
	if id == "" {
		return LootDungeon{}, fmt.Errorf("dungeon id is required")
	}
	if len(character.Inventory) == 0 {
		return LootDungeon{}, fmt.Errorf("cannot create a loot dungeon without crafted items")
	}

	items := make([]CraftedItem, 0, len(character.Inventory))
	value := 0
	for _, item := range character.Inventory {
		if !rules.EligibleForDungeon(item) {
			continue
		}
		itemValue := rules.ItemLegacyValue(item)
		if itemValue < 0 {
			return LootDungeon{}, fmt.Errorf("item %q has negative legacy value", item.ID)
		}
		items = append(items, item)
		value += itemValue
	}
	if len(items) == 0 {
		return LootDungeon{}, fmt.Errorf("cannot create a loot dungeon without eligible crafted items")
	}

	return LootDungeon{
		ID:          id,
		Name:        rules.DungeonName(character, character.Deaths),
		OriginRun:   character.Deaths,
		Depth:       max(1, rules.DungeonDepth(character, character.Deaths)),
		CreatedAt:   now,
		Status:      DungeonSealed,
		Items:       items,
		LegacyValue: value,
		Attributes:  cloneStringMap(rules.DungeonAttributes(character, items)),
	}, nil
}
