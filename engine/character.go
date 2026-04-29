package engine

import "fmt"

// CharacterID is the stable host-game identifier for a Character.
type CharacterID string

// Character is a returning playable or simulated character tracked by the
// engine. Host games may register many characters on one World.
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

// NewCharacter validates and creates a living level-one Character.
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

// Carry adds an item to the character inventory if the character is alive.
func (c *Character) Carry(item CraftedItem) error {
	if !c.Alive {
		return fmt.Errorf("dead characters cannot carry new items")
	}
	c.Inventory = append(c.Inventory, item)
	return nil
}

// Respawn marks the character alive and clears transient carried inventory.
func (c *Character) Respawn() {
	c.Alive = true
	if c.Level < 1 {
		c.Level = 1
	}
	c.Inventory = nil
}
