package engine

import (
	"fmt"
	"testing"
	"time"
)

func TestCustomRulesShapeDungeonCreation(t *testing.T) {
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
	rules.DungeonName = func(character Character, run int) string {
		return fmt.Sprintf("Run %d: %s", run, character.Name)
	}
	rules.DungeonDepth = func(Character, int) int {
		return 42
	}
	rules.DungeonAttributes = func(Character, []CraftedItem) map[string]string {
		return map[string]string{"biome": "glass"}
	}

	world := NewWorldWithRules(character, rules)

	eligible, err := NewCraftedItem("bound-spear", "Bound Spear", "crafter-1", 99, 8)
	if err != nil {
		t.Fatal(err)
	}
	eligible.Attributes["binds_legacy"] = "true"
	ineligible, err := NewCraftedItem("camp-cup", "Camp Cup", "crafter-1", 99, 2)
	if err != nil {
		t.Fatal(err)
	}

	if err := world.AddCraftedItem(eligible); err != nil {
		t.Fatal(err)
	}
	if err := world.AddCraftedItem(ineligible); err != nil {
		t.Fatal(err)
	}

	dungeon, err := world.KillCharacter("custom rule test")
	if err != nil {
		t.Fatal(err)
	}

	if dungeon.Name != "Run 1: Ash" {
		t.Fatalf("dungeon name = %q, want custom name", dungeon.Name)
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
	if dungeon.Attributes["biome"] != "glass" {
		t.Fatalf("dungeon biome = %q, want glass", dungeon.Attributes["biome"])
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

	item, err := NewCraftedItem("old-map", "Old Map", "crafter-1", 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	if err := world.AddCraftedItem(item); err != nil {
		t.Fatal(err)
	}
	dungeon, err := world.KillCharacter("time")
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
	if err := world.AddCraftedItem(item); err == nil {
		t.Fatal("expected negative custom score to fail")
	}
	if len(world.Character.Inventory) != 0 {
		t.Fatalf("inventory count = %d, want 0", len(world.Character.Inventory))
	}
}
