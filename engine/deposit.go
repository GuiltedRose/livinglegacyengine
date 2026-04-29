package engine

import (
	"fmt"
	"time"
)

// LootDepositID identifies a single item deposit into a dungeon pool.
type LootDepositID string

// DepositedLoot records the provenance of an item deposited into a loot
// dungeon, including who dropped it and who later claimed it.
type DepositedLoot struct {
	ID          LootDepositID     `json:"id"`
	Item        CraftedItem       `json:"item"`
	DroppedBy   ActorID           `json:"dropped_by"`
	AreaID      AreaID            `json:"area_id"`
	DungeonID   DungeonID         `json:"dungeon_id"`
	Cause       string            `json:"cause,omitempty"`
	DepositedAt time.Time         `json:"deposited_at"`
	ClaimedBy   ActorID           `json:"claimed_by,omitempty"`
	ClaimedAt   time.Time         `json:"claimed_at,omitempty"`
	Attributes  map[string]string `json:"attributes,omitempty"`
}

func newLootDepositID(dungeonID DungeonID, actorID ActorID, sequence int) LootDepositID {
	return LootDepositID(fmt.Sprintf("%s:%s:%04d", dungeonID, actorID, sequence))
}

// Claimed reports whether this deposit has been claimed by a dungeon clear.
func (d DepositedLoot) Claimed() bool {
	return d.ClaimedBy != ""
}
