package engine

import (
	"testing"
	"time"
)

func TestCustomRulesShapeDeathLootDeposit(t *testing.T) {
	character, err := NewCharacter("hero", "Ash")
	if err != nil {
		t.Fatal(err)
	}
	character.Level = 7

	rules := DefaultRules()
	rules.EligibleForDungeon = func(item CraftedItem) bool {
		return item.Attributes["binds_legacy"] == "true"
	}
	rules.ItemLegacyValue = func(item CraftedItem) int {
		return item.Power
	}

	world := NewWorldWithRules(character, rules)
	if _, err := world.SpawnLootDungeon("glass-cache", "Glass Cache", "glass-field", 42); err != nil {
		t.Fatal(err)
	}

	eligible, err := NewCraftedItem("bound-spear", "Bound Spear", "crafter-1", 99, 8)
	if err != nil {
		t.Fatal(err)
	}
	eligible.Attributes["binds_legacy"] = "true"
	ineligible, err := NewCraftedItem("camp-cup", "Camp Cup", "crafter-1", 99, 2)
	if err != nil {
		t.Fatal(err)
	}

	if err := world.CarryItem(character.ID, eligible); err != nil {
		t.Fatal(err)
	}
	if err := world.CarryItem(character.ID, ineligible); err != nil {
		t.Fatal(err)
	}

	dungeon, err := world.KillCharacterByID(character.ID, "custom rule test", "glass-field")
	if err != nil {
		t.Fatal(err)
	}

	if dungeon.Depth != 42 {
		t.Fatalf("dungeon depth = %d, want 42", dungeon.Depth)
	}
	if dungeon.LegacyValue != eligible.Power {
		t.Fatalf("legacy value = %d, want %d", dungeon.LegacyValue, eligible.Power)
	}
	if len(dungeon.Items) != 1 || dungeon.Items[0].ID != eligible.ID {
		t.Fatalf("dungeon items = %#v, want only eligible item", dungeon.Items)
	}
}

func TestCustomDecayRuleMarksSealedDungeons(t *testing.T) {
	character, err := NewCharacter("hero", "Ash")
	if err != nil {
		t.Fatal(err)
	}

	rules := DefaultRules()
	rules.ShouldDecayDungeon = func(dungeon LootDungeon, now time.Time) bool {
		return now.Sub(dungeon.CreatedAt) >= time.Hour
	}
	world := NewWorldWithRules(character, rules)
	world.SetClock(func() time.Time { return time.Unix(0, 0).UTC() })
	if _, err := world.SpawnLootDungeon("old-cache", "Old Cache", DefaultAreaID, 1); err != nil {
		t.Fatal(err)
	}

	item, err := NewCraftedItem("old-map", "Old Map", "crafter-1", 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if err := world.CarryItem(character.ID, item); err != nil {
		t.Fatal(err)
	}
	dungeon, err := world.KillCharacterByID(character.ID, "time", DefaultAreaID)
	if err != nil {
		t.Fatal(err)
	}

	world.SetClock(func() time.Time { return time.Unix(3600, 0).UTC() })
	decayed := world.DecayDungeons()

	if len(decayed) != 1 || decayed[0] != dungeon.ID {
		t.Fatalf("decayed ids = %#v, want %q", decayed, dungeon.ID)
	}
	if world.Dungeons[dungeon.ID].Status != DungeonDecayed {
		t.Fatalf("dungeon status = %q, want %q", world.Dungeons[dungeon.ID].Status, DungeonDecayed)
	}
	lastEvent := world.Events[len(world.Events)-1]
	if lastEvent.Type != EventDungeonDecayed {
		t.Fatalf("last event = %q, want %q", lastEvent.Type, EventDungeonDecayed)
	}
}

func TestNegativeCustomScoreRejectsBeforeInventoryMutation(t *testing.T) {
	character, err := NewCharacter("hero", "Ash")
	if err != nil {
		t.Fatal(err)
	}

	rules := DefaultRules()
	rules.ItemLegacyValue = func(CraftedItem) int {
		return -1
	}
	world := NewWorldWithRules(character, rules)

	item, err := NewCraftedItem("cursed", "Cursed Thing", "crafter-1", 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if err := world.CarryItem(character.ID, item); err == nil {
		t.Fatal("expected negative custom score to fail")
	}
	if len(world.PrimaryCharacter.Inventory) != 0 {
		t.Fatalf("inventory count = %d, want 0", len(world.PrimaryCharacter.Inventory))
	}
}
