package engine

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestMultiCharacterDeathLootAndRecovery(t *testing.T) {
	primary, err := NewCharacter("ash", "Ash")
	if err != nil {
		t.Fatal(err)
	}
	other, err := NewCharacter("mara", "Mara")
	if err != nil {
		t.Fatal(err)
	}
	world := NewWorld(primary)
	if err := world.AddCharacter(other); err != nil {
		t.Fatal(err)
	}
	if _, err := world.SpawnLootDungeon("road-cache", "Road Cache", "road", 1); err != nil {
		t.Fatal(err)
	}

	item, err := NewCraftedItem("mara-pin", "Mara Pin", "crafter-1", 2, 4)
	if err != nil {
		t.Fatal(err)
	}
	if err := world.CarryItem(other.ID, item); err != nil {
		t.Fatal(err)
	}
	dungeon, err := world.KillCharacterByID(other.ID, "bad road", "road")
	if err != nil {
		t.Fatal(err)
	}
	if err := world.RespawnCharacterByID(other.ID); err != nil {
		t.Fatal(err)
	}

	loot, err := world.LootDungeon(dungeon.ID, ActorID(other.ID))
	if err != nil {
		t.Fatal(err)
	}
	restored, ok := world.CharacterByID(other.ID)
	if !ok {
		t.Fatal("expected character to exist")
	}
	if restored.LegacyScore != item.LegacyValue() {
		t.Fatalf("legacy score = %d, want %d", restored.LegacyScore, item.LegacyValue())
	}
	if restored.RecoveredItems[0].ID != loot[0].ID {
		t.Fatalf("recovered item = %q, want %q", restored.RecoveredItems[0].ID, loot[0].ID)
	}
}

func TestCustomLootSelectionRule(t *testing.T) {
	character, err := NewCharacter("ash", "Ash")
	if err != nil {
		t.Fatal(err)
	}
	rules := DefaultRules()
	rules.SelectLoot = func(LootDungeon, ActorID) int {
		return 1
	}
	world := NewWorldWithRules(character, rules)
	if _, err := world.SpawnLootDungeon("cache", "Cache", DefaultAreaID, 1); err != nil {
		t.Fatal(err)
	}
	first, err := NewCraftedItem("first", "First", "crafter-1", 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	second, err := NewCraftedItem("second", "Second", "crafter-1", 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := world.DepositDeathLoot("actor-1", DefaultAreaID, []CraftedItem{first, second}, "test"); err != nil {
		t.Fatal(err)
	}

	loot, err := world.LootDungeon("cache", "actor-2")
	if err != nil {
		t.Fatal(err)
	}
	if loot[0].ID != "second" {
		t.Fatalf("loot id = %q, want second", loot[0].ID)
	}
}

func TestSentinelErrors(t *testing.T) {
	character, err := NewCharacter("ash", "Ash")
	if err != nil {
		t.Fatal(err)
	}
	world := NewWorld(character)

	if _, err := world.DepositDeathLoot("actor-1", "missing", nil, "test"); !errors.Is(err, ErrNoLootDungeonInArea) {
		t.Fatalf("error = %v, want ErrNoLootDungeonInArea", err)
	}
	if _, err := world.LootDungeon("missing", "actor-1"); !errors.Is(err, ErrDungeonNotFound) {
		t.Fatalf("error = %v, want ErrDungeonNotFound", err)
	}
}

func TestFileSnapshotStore(t *testing.T) {
	character, err := NewCharacter("ash", "Ash")
	if err != nil {
		t.Fatal(err)
	}
	world := NewWorld(character)
	store := NewFileSnapshotStore(filepath.Join(t.TempDir(), "world.json"))

	if err := store.Save(world.Snapshot()); err != nil {
		t.Fatal(err)
	}
	loaded, err := store.Load()
	if err != nil {
		t.Fatal(err)
	}
	restored, err := RestoreWorld(loaded)
	if err != nil {
		t.Fatal(err)
	}
	if restored.PrimaryCharacter.ID != character.ID {
		t.Fatalf("restored character = %q, want %q", restored.PrimaryCharacter.ID, character.ID)
	}
}

func TestSafeWorldUpdateAndSnapshot(t *testing.T) {
	character, err := NewCharacter("ash", "Ash")
	if err != nil {
		t.Fatal(err)
	}
	safe := NewSafeWorld(NewWorld(character))

	if err := safe.Update(func(world *World) error {
		_, err := world.SpawnLootDungeon("cache", "Cache", DefaultAreaID, 1)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	snapshot := safe.Snapshot()
	if _, ok := snapshot.Dungeons["cache"]; !ok {
		t.Fatal("expected dungeon in safe snapshot")
	}
}
