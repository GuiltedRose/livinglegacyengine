package engine

import "fmt"

// ItemID is the stable host-game identifier for a crafted item.
type ItemID string

// ActorID is the stable host-game identifier for any actor-like entity.
type ActorID string

// CraftedItem is an actor-created item eligible for legacy dungeon pools.
type CraftedItem struct {
	ID         ItemID            `json:"id"`
	Name       string            `json:"name"`
	CrafterID  ActorID           `json:"crafter_id"`
	Rarity     int               `json:"rarity"`
	Quality    int               `json:"quality"`
	Power      int               `json:"power"`
	Tags       []string          `json:"tags,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

// NewCraftedItem validates and creates a CraftedItem. Rarity defaults to
// quality so callers get a useful rarity floor without extra setup.
func NewCraftedItem(id ItemID, name string, crafterID ActorID, quality int, power int, tags ...string) (CraftedItem, error) {
	if id == "" {
		return CraftedItem{}, fmt.Errorf("item id is required")
	}
	if name == "" {
		return CraftedItem{}, fmt.Errorf("item name is required")
	}
	if crafterID == "" {
		return CraftedItem{}, fmt.Errorf("crafter id is required")
	}
	if quality < 1 {
		return CraftedItem{}, fmt.Errorf("quality must be at least 1")
	}
	if power < 0 {
		return CraftedItem{}, fmt.Errorf("power cannot be negative")
	}

	return CraftedItem{
		ID:         id,
		Name:       name,
		CrafterID:  crafterID,
		Rarity:     quality,
		Quality:    quality,
		Power:      power,
		Tags:       append([]string(nil), tags...),
		Attributes: map[string]string{},
	}, nil
}

// LegacyValue returns the default item score used by the engine.
func (i CraftedItem) LegacyValue() int {
	return i.Quality*10 + i.Power
}
