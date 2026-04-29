package engine

import (
	"encoding/json"
	"testing"
	"time"
)

func TestSnapshotJSONRoundTripRestoresWorld(t *testing.T) {
	character, err := NewCharacter("hero", "Ash")
	if err != nil {
		t.Fatal(err)
	}
	world := NewWorld(character)
	world.SetClock(func() time.Time { return time.Unix(123, 0).UTC() })

	actor, err := NewActor("crafter-1", "Glasswright", ActorCrafter)
	if err != nil {
		t.Fatal(err)
	}
	actor.Attributes["guild"] = "lantern"
	if err := world.AddActor(actor); err != nil {
		t.Fatal(err)
	}

	item, err := NewCraftedItem("glass-key", "Glass Key", actor.ID, 2, 5, "key")
	if err != nil {
		t.Fatal(err)
	}
	item.Attributes["material"] = "glass"
	if err := world.CarryItem(character.ID, item); err != nil {
		t.Fatal(err)
	}
	if _, err := world.SpawnLootDungeon("mirror-cache", "Mirror Cache", DefaultAreaID, 2); err != nil {
		t.Fatal(err)
	}
	dungeon, err := world.KillCharacterByID(character.ID, "the mirror hall", DefaultAreaID)
	if err != nil {
		t.Fatal(err)
	}

	rumor, err := NewRumor("rumor-1", actor.ID, "Ash fell in the mirror hall.", 0.8, 4, time.Unix(124, 0).UTC())
	if err != nil {
		t.Fatal(err)
	}
	rumor.SubjectID = string(character.ID)
	rumor.ObjectID = string(dungeon.ID)
	if err := world.AddRumor(rumor); err != nil {
		t.Fatal(err)
	}
	if _, err := world.SpreadRumor(rumor.ID, "listener-1", 0.1); err != nil {
		t.Fatal(err)
	}

	data, err := MarshalSnapshot(world.Snapshot())
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	if _, ok := raw["next_run_id"]; !ok {
		t.Fatal("expected stable snake_case JSON field next_run_id")
	}

	snapshot, err := UnmarshalSnapshot(data)
	if err != nil {
		t.Fatal(err)
	}
	restored, err := RestoreWorld(snapshot)
	if err != nil {
		t.Fatal(err)
	}

	if restored.PrimaryCharacter.ID != character.ID {
		t.Fatalf("character id = %q, want %q", restored.PrimaryCharacter.ID, character.ID)
	}
	if restored.Dungeons[dungeon.ID].Items[0].Attributes["material"] != "glass" {
		t.Fatalf("restored item attributes were lost")
	}
	if restored.Actors[actor.ID].Attributes["guild"] != "lantern" {
		t.Fatalf("restored actor attributes were lost")
	}
	if len(restored.Rumors) != 1 || restored.Rumors[0].Spread != 1 {
		t.Fatalf("restored rumors = %#v, want one spread rumor", restored.Rumors)
	}
	if len(restored.Memories["listener-1"].KnownRumors) != 1 {
		t.Fatalf("listener memory rumor count = %d, want 1", len(restored.Memories["listener-1"].KnownRumors))
	}
	perception := restored.Perception("listener-1", string(character.ID))
	if perception.Notoriety != rumor.Impact {
		t.Fatalf("restored perception notoriety = %d, want %d", perception.Notoriety, rumor.Impact)
	}
	if len(restored.Events) == 0 {
		t.Fatal("events should survive JSON round trip")
	}
}
