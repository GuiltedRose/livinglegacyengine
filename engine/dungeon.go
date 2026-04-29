package engine

import (
	"fmt"
	"hash/fnv"
	"time"
)

type DungeonID string
type AreaID string

type DungeonStatus string

const (
	DefaultAreaID AreaID = "default"

	DungeonDormant DungeonStatus = "dormant"
	DungeonActive  DungeonStatus = "active"
	DungeonLocked  DungeonStatus = "locked"
	DungeonSealed  DungeonStatus = DungeonActive
	DungeonLooted  DungeonStatus = "looted"
	DungeonDecayed DungeonStatus = "decayed"
)

type LootDungeon struct {
	ID          DungeonID         `json:"id"`
	Name        string            `json:"name"`
	AreaID      AreaID            `json:"area_id"`
	OriginRun   int               `json:"origin_run,omitempty"`
	Depth       int               `json:"depth"`
	CreatedAt   time.Time         `json:"created_at"`
	Status      DungeonStatus     `json:"status"`
	Items       []CraftedItem     `json:"items,omitempty"`
	Deposits    []DepositedLoot   `json:"deposits,omitempty"`
	LegacyValue int               `json:"legacy_value"`
	LootedBy    ActorID           `json:"looted_by,omitempty"`
	LockedTo    ActorID           `json:"locked_to,omitempty"`
	Attributes  map[string]string `json:"attributes,omitempty"`
}

func NewSpawnedLootDungeon(id DungeonID, name string, areaID AreaID, depth int, now time.Time) (LootDungeon, error) {
	if id == "" {
		return LootDungeon{}, fmt.Errorf("dungeon id is required")
	}
	if name == "" {
		return LootDungeon{}, fmt.Errorf("dungeon name is required")
	}
	if areaID == "" {
		areaID = DefaultAreaID
	}
	return LootDungeon{
		ID:         id,
		Name:       name,
		AreaID:     areaID,
		Depth:      max(1, depth),
		CreatedAt:  now,
		Status:     DungeonDormant,
		Attributes: map[string]string{},
	}, nil
}

func (d *LootDungeon) syncItemsFromDeposits() {
	if len(d.Deposits) == 0 {
		return
	}
	items := make([]CraftedItem, 0, len(d.Deposits))
	for _, deposit := range d.Deposits {
		if deposit.Claimed() {
			continue
		}
		items = append(items, deposit.Item)
	}
	d.Items = items
}

func (d *LootDungeon) syncDepositsFromItems(now time.Time) {
	if len(d.Deposits) > 0 || len(d.Items) == 0 {
		return
	}
	for index, item := range d.Items {
		d.Deposits = append(d.Deposits, DepositedLoot{
			ID:          newLootDepositID(d.ID, item.CrafterID, index+1),
			Item:        item,
			DroppedBy:   item.CrafterID,
			AreaID:      d.AreaID,
			DungeonID:   d.ID,
			DepositedAt: now,
			Attributes:  map[string]string{},
		})
	}
}

func (d LootDungeon) unclaimedDepositIndex(itemIndex int) int {
	if itemIndex < 0 {
		return -1
	}
	unclaimedIndex := 0
	for index, deposit := range d.Deposits {
		if deposit.Claimed() {
			continue
		}
		if unclaimedIndex == itemIndex {
			return index
		}
		unclaimedIndex++
	}
	return -1
}

func selectLootIndex(dungeon LootDungeon, looter ActorID) int {
	if len(dungeon.Items) == 0 {
		return -1
	}
	hash := fnv.New64a()
	_, _ = hash.Write([]byte(dungeon.ID))
	_, _ = hash.Write([]byte(looter))
	_, _ = hash.Write([]byte(fmt.Sprint(dungeon.LegacyValue)))
	for _, item := range dungeon.Items {
		_, _ = hash.Write([]byte(item.ID))
	}
	return int(hash.Sum64() % uint64(len(dungeon.Items)))
}
