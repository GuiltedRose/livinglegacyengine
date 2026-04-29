package engine

import (
	"testing"
	"time"
)

func TestDeathDepositsCraftedItemsIntoSpawnedAreaDungeon(t *testing.T) {
	character, err := NewCharacter("hero", "Ash")
	if err != nil {
		t.Fatal(err)
	}
	world := NewWorld(character)
	world.SetClock(func() time.Time { return time.Unix(100, 0).UTC() })
	spawned, err := world.SpawnLootDungeon("area-cache", "Area Cache", "ash-road", 3)
	if err != nil {
		t.Fatal(err)
	}

	item, err := NewCraftedItem("ember-axe", "Ember Axe", "actor-smith", 3, 9)
	if err != nil {
		t.Fatal(err)
	}
	if err := world.CarryItem(character.ID, item); err != nil {
		t.Fatal(err)
	}

	dungeon, err := world.KillCharacterByID(character.ID, "a bad pull", "ash-road")
	if err != nil {
		t.Fatal(err)
	}

	if dungeon.ID != spawned.ID {
		t.Fatalf("dungeon id = %q, want spawned dungeon %q", dungeon.ID, spawned.ID)
	}
	if dungeon.Status != DungeonActive {
		t.Fatalf("dungeon status = %q, want %q", dungeon.Status, DungeonActive)
	}
	if dungeon.AreaID != "ash-road" {
		t.Fatalf("area id = %q, want ash-road", dungeon.AreaID)
	}
	if len(dungeon.Items) != 1 {
		t.Fatalf("dungeon item count = %d, want 1", len(dungeon.Items))
	}
	if len(dungeon.Deposits) != 1 {
		t.Fatalf("deposit count = %d, want 1", len(dungeon.Deposits))
	}
	if dungeon.Deposits[0].DroppedBy != ActorID(character.ID) {
		t.Fatalf("deposit dropped by = %q, want %q", dungeon.Deposits[0].DroppedBy, character.ID)
	}
	if dungeon.Deposits[0].Cause != "a bad pull" {
		t.Fatalf("deposit cause = %q, want a bad pull", dungeon.Deposits[0].Cause)
	}
	if dungeon.LegacyValue != item.LegacyValue() {
		t.Fatalf("legacy value = %d, want %d", dungeon.LegacyValue, item.LegacyValue())
	}
	if item.Rarity != item.Quality {
		t.Fatalf("default rarity = %d, want quality %d", item.Rarity, item.Quality)
	}
}

func TestLootingDungeonIncreasesSameCharacterLegacy(t *testing.T) {
	character, err := NewCharacter("hero", "Ash")
	if err != nil {
		t.Fatal(err)
	}
	world := NewWorld(character)
	if _, err := world.SpawnLootDungeon("default-cache", "Default Cache", DefaultAreaID, 1); err != nil {
		t.Fatal(err)
	}

	item, err := NewCraftedItem("glass-key", "Glass Key", "actor-smith", 1, 5)
	if err != nil {
		t.Fatal(err)
	}
	if err := world.CarryItem(character.ID, item); err != nil {
		t.Fatal(err)
	}
	dungeon, err := world.KillCharacterByID(character.ID, "forgot to rest", DefaultAreaID)
	if err != nil {
		t.Fatal(err)
	}
	world.RespawnCharacterByID(character.ID)

	loot, err := world.LootDungeon(dungeon.ID, "actor-raider")
	if err != nil {
		t.Fatal(err)
	}

	if len(loot) != 1 {
		t.Fatalf("loot count = %d, want 1", len(loot))
	}
	if world.Dungeons[dungeon.ID].Status != DungeonLocked {
		t.Fatalf("dungeon status = %q, want locked", world.Dungeons[dungeon.ID].Status)
	}
	if len(world.Dungeons[dungeon.ID].Deposits) != 1 {
		t.Fatalf("deposit count = %d, want 1 claimed record", len(world.Dungeons[dungeon.ID].Deposits))
	}
	if world.Dungeons[dungeon.ID].Deposits[0].ClaimedBy != "actor-raider" {
		t.Fatalf("claimed by = %q, want actor-raider", world.Dungeons[dungeon.ID].Deposits[0].ClaimedBy)
	}
	if len(world.Dungeons[dungeon.ID].Items) != 0 {
		t.Fatalf("remaining item count = %d, want 0", len(world.Dungeons[dungeon.ID].Items))
	}
	if world.Dungeons[dungeon.ID].LockedTo != "actor-raider" {
		t.Fatalf("locked to = %q, want actor-raider", world.Dungeons[dungeon.ID].LockedTo)
	}
	if world.PrimaryCharacter.Deaths != 1 {
		t.Fatalf("deaths = %d, want 1", world.PrimaryCharacter.Deaths)
	}
	if world.PrimaryCharacter.LegacyScore != 0 {
		t.Fatalf("legacy score = %d, want 0 because another actor cleared it", world.PrimaryCharacter.LegacyScore)
	}
	if len(world.PrimaryCharacter.RecoveredItems) != 0 {
		t.Fatalf("recovered item count = %d, want 0", len(world.PrimaryCharacter.RecoveredItems))
	}
}

func TestFailedDeathDoesNotMutateCharacter(t *testing.T) {
	character, err := NewCharacter("hero", "Ash")
	if err != nil {
		t.Fatal(err)
	}
	world := NewWorld(character)
	if _, err := world.SpawnLootDungeon("default-cache", "Default Cache", DefaultAreaID, 1); err != nil {
		t.Fatal(err)
	}

	if _, err := world.KillCharacterByID(character.ID, "empty pockets", DefaultAreaID); err == nil {
		t.Fatal("expected death without crafted loot to fail")
	}
	if !world.PrimaryCharacter.Alive {
		t.Fatal("character should still be alive after failed loot deposit")
	}
	if world.PrimaryCharacter.Deaths != 0 {
		t.Fatalf("deaths = %d, want 0", world.PrimaryCharacter.Deaths)
	}
}

func TestSpawnedDungeonIsInaccessibleUntilDeathLootDeposited(t *testing.T) {
	character, err := NewCharacter("hero", "Ash")
	if err != nil {
		t.Fatal(err)
	}
	world := NewWorld(character)
	dungeon, err := world.SpawnLootDungeon("default-cache", "Default Cache", DefaultAreaID, 1)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := world.LootDungeon(dungeon.ID, "actor-raider"); err == nil {
		t.Fatal("expected dormant dungeon to be inaccessible")
	}
}

func TestLockedDungeonUnlocksWhenClearerDiesAgain(t *testing.T) {
	character, err := NewCharacter("hero", "Ash")
	if err != nil {
		t.Fatal(err)
	}
	world := NewWorld(character)
	if _, err := world.SpawnLootDungeon("default-cache", "Default Cache", DefaultAreaID, 1); err != nil {
		t.Fatal(err)
	}

	item, err := NewCraftedItem("glass-key", "Glass Key", "actor-smith", 1, 5)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := world.DepositDeathLoot("actor-raider", DefaultAreaID, []CraftedItem{item}, "first death"); err != nil {
		t.Fatal(err)
	}
	if _, err := world.LootDungeon("default-cache", "actor-raider"); err != nil {
		t.Fatal(err)
	}
	if world.Dungeons["default-cache"].Status != DungeonLocked {
		t.Fatalf("dungeon status = %q, want locked", world.Dungeons["default-cache"].Status)
	}
	if _, err := world.DepositDeathLoot("actor-raider", DefaultAreaID, []CraftedItem{item}, "second death"); err != nil {
		t.Fatal(err)
	}
	if world.Dungeons["default-cache"].Status != DungeonActive {
		t.Fatalf("dungeon status = %q, want active after clearer dies again", world.Dungeons["default-cache"].Status)
	}
}

func TestLockedDungeonDoesNotUnlockWhenClearerDiesWithoutCraftedLoot(t *testing.T) {
	character, err := NewCharacter("hero", "Ash")
	if err != nil {
		t.Fatal(err)
	}
	world := NewWorld(character)
	if _, err := world.SpawnLootDungeon("default-cache", "Default Cache", DefaultAreaID, 1); err != nil {
		t.Fatal(err)
	}

	item, err := NewCraftedItem("glass-key", "Glass Key", "actor-smith", 1, 5)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := world.DepositDeathLoot("actor-raider", DefaultAreaID, []CraftedItem{item}, "first death"); err != nil {
		t.Fatal(err)
	}
	if _, err := world.LootDungeon("default-cache", "actor-raider"); err != nil {
		t.Fatal(err)
	}
	if _, err := world.DepositDeathLoot("actor-raider", DefaultAreaID, nil, "empty death"); err == nil {
		t.Fatal("expected crafted-loot-free death deposit to fail")
	}
	if world.Dungeons["default-cache"].Status != DungeonLocked {
		t.Fatalf("dungeon status = %q, want locked", world.Dungeons["default-cache"].Status)
	}
	if world.Dungeons["default-cache"].LockedTo != "actor-raider" {
		t.Fatalf("locked to = %q, want actor-raider", world.Dungeons["default-cache"].LockedTo)
	}
}

func TestLootRewardCannotDowngradeDroppedRarity(t *testing.T) {
	character, err := NewCharacter("hero", "Ash")
	if err != nil {
		t.Fatal(err)
	}
	rules := DefaultRules()
	rules.LootReward = func(source CraftedItem) CraftedItem {
		reward := source
		reward.Rarity = source.Rarity - 1
		return reward
	}
	world := NewWorldWithRules(character, rules)
	if _, err := world.SpawnLootDungeon("default-cache", "Default Cache", DefaultAreaID, 1); err != nil {
		t.Fatal(err)
	}

	item, err := NewCraftedItem("glass-key", "Glass Key", "actor-smith", 3, 5)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := world.DepositDeathLoot("actor-raider", DefaultAreaID, []CraftedItem{item}, "death"); err != nil {
		t.Fatal(err)
	}
	if _, err := world.LootDungeon("default-cache", "actor-raider"); err == nil {
		t.Fatal("expected downgraded reward rarity to fail")
	}
	if world.Dungeons["default-cache"].Status != DungeonActive {
		t.Fatalf("dungeon status = %q, want active after failed reward", world.Dungeons["default-cache"].Status)
	}
}

func TestLootRewardMayUpgradeDroppedRarity(t *testing.T) {
	character, err := NewCharacter("hero", "Ash")
	if err != nil {
		t.Fatal(err)
	}
	rules := DefaultRules()
	rules.LootReward = func(source CraftedItem) CraftedItem {
		reward := source
		reward.ID = "upgraded-" + source.ID
		reward.Name = "Upgraded " + source.Name
		reward.Rarity = source.Rarity + 2
		return reward
	}
	world := NewWorldWithRules(character, rules)
	if _, err := world.SpawnLootDungeon("default-cache", "Default Cache", DefaultAreaID, 1); err != nil {
		t.Fatal(err)
	}

	item, err := NewCraftedItem("glass-key", "Glass Key", "actor-smith", 3, 5)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := world.DepositDeathLoot("actor-raider", DefaultAreaID, []CraftedItem{item}, "death"); err != nil {
		t.Fatal(err)
	}
	loot, err := world.LootDungeon("default-cache", "actor-raider")
	if err != nil {
		t.Fatal(err)
	}

	if loot[0].ID != "upgraded-glass-key" {
		t.Fatalf("loot id = %q, want upgraded item", loot[0].ID)
	}
	if loot[0].Rarity != item.Rarity+2 {
		t.Fatalf("loot rarity = %d, want %d", loot[0].Rarity, item.Rarity+2)
	}
}

func TestSnapshotRestoreKeepsWorldPortable(t *testing.T) {
	character, err := NewCharacter("hero", "Ash")
	if err != nil {
		t.Fatal(err)
	}
	world := NewWorld(character)

	item, err := NewCraftedItem("star-thread", "Star Thread", "crafter-1", 2, 3, "cloth")
	if err != nil {
		t.Fatal(err)
	}
	item.Attributes["origin_game"] = "test-suite"
	if err := world.CarryItem(character.ID, item); err != nil {
		t.Fatal(err)
	}

	snapshot := world.Snapshot()
	snapshot.PrimaryCharacter.Inventory[0].Name = "mutated outside"
	snapshot.PrimaryCharacter.Inventory[0].Attributes["origin_game"] = "mutated"

	restored, err := RestoreWorld(world.Snapshot())
	if err != nil {
		t.Fatal(err)
	}

	if restored.PrimaryCharacter.Inventory[0].Name != item.Name {
		t.Fatalf("restored item name = %q, want %q", restored.PrimaryCharacter.Inventory[0].Name, item.Name)
	}
	if restored.PrimaryCharacter.Inventory[0].Attributes["origin_game"] != "test-suite" {
		t.Fatalf("restored attributes were not preserved")
	}
	if restored.Crafters[item.CrafterID] != item.LegacyValue() {
		t.Fatalf("crafter score = %d, want %d", restored.Crafters[item.CrafterID], item.LegacyValue())
	}
}

func TestEventsIncludeStructuredData(t *testing.T) {
	character, err := NewCharacter("hero", "Ash")
	if err != nil {
		t.Fatal(err)
	}
	world := NewWorld(character)

	item, err := NewCraftedItem("mirror-pin", "Mirror Pin", "crafter-1", 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if err := world.CarryItem(character.ID, item); err != nil {
		t.Fatal(err)
	}

	event := world.Events[0]
	if event.SubjectID != string(item.CrafterID) {
		t.Fatalf("subject id = %q, want %q", event.SubjectID, item.CrafterID)
	}
	if event.ObjectID != string(item.ID) {
		t.Fatalf("object id = %q, want %q", event.ObjectID, item.ID)
	}
	if event.Data["character_id"] != string(character.ID) {
		t.Fatalf("event character id = %q, want %q", event.Data["character_id"], character.ID)
	}
}
