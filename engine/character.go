package engine

import "fmt"

type CharacterID string

type Character struct {
	ID             CharacterID   `json:"id"`
	Name           string        `json:"name"`
	Level          int           `json:"level"`
	Deaths         int           `json:"deaths"`
	Alive          bool          `json:"alive"`
	LegacyScore    int           `json:"legacy_score"`
	Inventory      []CraftedItem `json:"inventory,omitempty"`
	RecoveredItems []CraftedItem `json:"recovered_items,omitempty"`
}

func NewCharacter(id CharacterID, name string) (Character, error) {
	if id == "" {
		return Character{}, fmt.Errorf("character id is required")
	}
	if name == "" {
		return Character{}, fmt.Errorf("character name is required")
	}

	return Character{
		ID:    id,
		Name:  name,
		Level: 1,
		Alive: true,
	}, nil
}

func (c *Character) Carry(item CraftedItem) error {
	if !c.Alive {
		return fmt.Errorf("dead characters cannot carry new items")
	}
	c.Inventory = append(c.Inventory, item)
	return nil
}

func (c *Character) Respawn() {
	c.Alive = true
	if c.Level < 1 {
		c.Level = 1
	}
	c.Inventory = nil
}
