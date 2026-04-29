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

func TestDepositQueryHelpers(t *testing.T) {
	character, err := NewCharacter("hero", "Ash")
	if err != nil {
		t.Fatal(err)
	}
	world := NewWorld(character)
	if _, err := world.SpawnLootDungeon("cache", "Cache", "road", 1); err != nil {
		t.Fatal(err)
	}
	item, err := NewCraftedItem("pin", "Pin", "crafter-1", 2, 3)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := world.DepositDeathLoot("dropper-1", "road", []CraftedItem{item}, "ambush"); err != nil {
		t.Fatal(err)
	}

	if got := len(world.DepositsByDropper("dropper-1")); got != 1 {
		t.Fatalf("deposits by dropper = %d, want 1", got)
	}
	if got := len(world.DepositsInDungeon("cache")); got != 1 {
		t.Fatalf("deposits in dungeon = %d, want 1", got)
	}
	if got := len(world.UnclaimedDepositsInArea("road")); got != 1 {
		t.Fatalf("unclaimed deposits in area = %d, want 1", got)
	}
	if got := len(world.DepositsForItem("pin")); got != 1 {
		t.Fatalf("deposits for item = %d, want 1", got)
	}

	deposits := world.DepositsInDungeon("cache")
	deposits[0].Item.Name = "mutated"
	again := world.DepositsInDungeon("cache")
	if again[0].Item.Name != "Pin" {
		t.Fatalf("deposit query returned aliased item")
	}
}

func TestClaimQueries(t *testing.T) {
	character, err := NewCharacter("hero", "Ash")
	if err != nil {
		t.Fatal(err)
	}
	world := NewWorld(character)
	if _, err := world.SpawnLootDungeon("cache", "Cache", DefaultAreaID, 1); err != nil {
		t.Fatal(err)
	}
	item, err := NewCraftedItem("pin", "Pin", "crafter-1", 2, 3)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := world.DepositDeathLoot("dropper-1", DefaultAreaID, []CraftedItem{item}, "ambush"); err != nil {
		t.Fatal(err)
	}
	if _, err := world.LootDungeon("cache", "claimer-1"); err != nil {
		t.Fatal(err)
	}

	if got := len(world.DepositsByClaimer("claimer-1")); got != 1 {
		t.Fatalf("deposits by claimer = %d, want 1", got)
	}
	if got := len(world.UnclaimedDepositsInDungeon("cache")); got != 0 {
		t.Fatalf("unclaimed deposits = %d, want 0", got)
	}
}

func TestDepositAndClaimHooksCanEmitEventsAndRumors(t *testing.T) {
	character, err := NewCharacter("hero", "Ash")
	if err != nil {
		t.Fatal(err)
	}
	rules := DefaultRules()
	rules.DepositEvents = func(deposit DepositedLoot) []Event {
		if deposit.Item.Rarity < 3 {
			return nil
		}
		return []Event{{
			Type:        EventLootDeposited,
			Description: "rare loot entered the pool",
			SubjectID:   string(deposit.DroppedBy),
			Object:      ItemRef(deposit.Item),
		}}
	}
	rules.ClaimRumors = func(deposit DepositedLoot, reward CraftedItem) []Rumor {
		rumor, err := NewRumor("rare-claim", deposit.ClaimedBy, "rare loot was claimed", 1, reward.Rarity, time.Unix(1, 0).UTC())
		if err != nil {
			return nil
		}
		rumor.Subject = ItemRef(reward)
		return []Rumor{rumor}
	}
	world := NewWorldWithRules(character, rules)
	if _, err := world.SpawnLootDungeon("cache", "Cache", DefaultAreaID, 1); err != nil {
		t.Fatal(err)
	}
	item, err := NewCraftedItem("rare-pin", "Rare Pin", "crafter-1", 3, 3)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := world.DepositDeathLoot("dropper-1", DefaultAreaID, []CraftedItem{item}, "ambush"); err != nil {
		t.Fatal(err)
	}
	if _, err := world.LootDungeon("cache", "claimer-1"); err != nil {
		t.Fatal(err)
	}

	if got := len(world.EventsMatching(EventFilter{TargetID: string(item.ID)})); got == 0 {
		t.Fatalf("expected hook event for item target")
	}
	if got := len(world.RumorsAbout(string(item.ID))); got != 1 {
		t.Fatalf("rumors about item = %d, want 1", got)
	}
}

func itemlessTime() time.Time {
	return time.Unix(1, 0).UTC()
}
