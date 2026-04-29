package engine

import (
	"testing"
	"time"
)

func TestDeathCreatesLootDungeonFromCraftedItems(t *testing.T) {
	character, err := NewCharacter("hero", "Ash")
	if err != nil {
		t.Fatal(err)
	}
	world := NewWorld(character)
	world.SetClock(func() time.Time { return time.Unix(100, 0).UTC() })

	item, err := NewCraftedItem("ember-axe", "Ember Axe", "actor-smith", 3, 9)
	if err != nil {
		t.Fatal(err)
	}
	if err := world.AddCraftedItem(item); err != nil {
		t.Fatal(err)
	}

	dungeon, err := world.KillCharacter("a bad pull")
	if err != nil {
		t.Fatal(err)
	}

	if dungeon.Status != DungeonSealed {
		t.Fatalf("dungeon status = %q, want %q", dungeon.Status, DungeonSealed)
	}
	if dungeon.OriginRun != 1 {
		t.Fatalf("origin run = %d, want 1", dungeon.OriginRun)
	}
	if dungeon.Name != "Ash's Fallen Cache 1" {
		t.Fatalf("dungeon name = %q, want first cache", dungeon.Name)
	}
	if len(dungeon.Items) != 1 {
		t.Fatalf("dungeon item count = %d, want 1", len(dungeon.Items))
	}
	if dungeon.LegacyValue != item.LegacyValue() {
		t.Fatalf("legacy value = %d, want %d", dungeon.LegacyValue, item.LegacyValue())
	}
}

func TestLootingDungeonIncreasesSameCharacterLegacy(t *testing.T) {
	character, err := NewCharacter("hero", "Ash")
	if err != nil {
		t.Fatal(err)
	}
	world := NewWorld(character)

	item, err := NewCraftedItem("glass-key", "Glass Key", "actor-smith", 1, 5)
	if err != nil {
		t.Fatal(err)
	}
	if err := world.AddCraftedItem(item); err != nil {
		t.Fatal(err)
	}
	dungeon, err := world.KillCharacter("forgot to rest")
	if err != nil {
		t.Fatal(err)
	}
	world.RespawnCharacter()

	loot, err := world.LootDungeon(dungeon.ID, "actor-raider")
	if err != nil {
		t.Fatal(err)
	}

	if len(loot) != 1 {
		t.Fatalf("loot count = %d, want 1", len(loot))
	}
	if world.Character.Deaths != 1 {
		t.Fatalf("deaths = %d, want 1", world.Character.Deaths)
	}
	if world.Character.LegacyScore != item.LegacyValue() {
		t.Fatalf("legacy score = %d, want %d", world.Character.LegacyScore, item.LegacyValue())
	}
	if world.Character.RecoveredItems[0].ID != item.ID {
		t.Fatalf("recovered item = %q, want %q", world.Character.RecoveredItems[0].ID, item.ID)
	}
}

func TestFailedDeathDoesNotMutateCharacter(t *testing.T) {
	character, err := NewCharacter("hero", "Ash")
	if err != nil {
		t.Fatal(err)
	}
	world := NewWorld(character)

	if _, err := world.KillCharacter("empty pockets"); err == nil {
		t.Fatal("expected death without crafted loot to fail")
	}
	if !world.Character.Alive {
		t.Fatal("character should still be alive after failed dungeon creation")
	}
	if world.Character.Deaths != 0 {
		t.Fatalf("deaths = %d, want 0", world.Character.Deaths)
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
	if err := world.AddCraftedItem(item); err != nil {
		t.Fatal(err)
	}

	snapshot := world.Snapshot()
	snapshot.Character.Inventory[0].Name = "mutated outside"
	snapshot.Character.Inventory[0].Attributes["origin_game"] = "mutated"

	restored, err := RestoreWorld(world.Snapshot())
	if err != nil {
		t.Fatal(err)
	}

	if restored.Character.Inventory[0].Name != item.Name {
		t.Fatalf("restored item name = %q, want %q", restored.Character.Inventory[0].Name, item.Name)
	}
	if restored.Character.Inventory[0].Attributes["origin_game"] != "test-suite" {
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
	if err := world.AddCraftedItem(item); err != nil {
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
