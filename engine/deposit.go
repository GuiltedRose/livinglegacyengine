package engine

import (
	"fmt"
	"time"
)

type LootDepositID string

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

func (d DepositedLoot) Claimed() bool {
	return d.ClaimedBy != ""
}
