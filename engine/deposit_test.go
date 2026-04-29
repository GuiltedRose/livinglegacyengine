package engine

import (
	"testing"
	"time"
)

func TestLegacyItemOnlyDungeonRebuildsDepositsOnRestore(t *testing.T) {
	character, err := NewCharacter("hero", "Ash")
	if err != nil {
		t.Fatal(err)
	}
	item, err := NewCraftedItem("old-item", "Old Item", "crafter-1", 2, 3)
	if err != nil {
		t.Fatal(err)
	}
	snapshot := WorldSnapshot{
		Version:          CurrentSnapshotVersion,
		PrimaryCharacter: character,
		Dungeons: map[DungeonID]LootDungeon{
			"old-cache": {
				ID:        "old-cache",
				Name:      "Old Cache",
				AreaID:    DefaultAreaID,
				CreatedAt: itemlessTime(),
				Status:    DungeonActive,
				Items:     []CraftedItem{item},
			},
		},
	}

	restored, err := RestoreWorld(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	dungeon := restored.Dungeons["old-cache"]
	if len(dungeon.Deposits) != 1 {
		t.Fatalf("deposit count = %d, want 1", len(dungeon.Deposits))
	}
	if dungeon.Deposits[0].Item.ID != item.ID {
		t.Fatalf("deposit item = %q, want %q", dungeon.Deposits[0].Item.ID, item.ID)
	}
}

func itemlessTime() time.Time {
	return time.Unix(1, 0).UTC()
}
