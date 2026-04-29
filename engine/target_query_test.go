package engine

import (
	"testing"
	"time"
)

func TestTargetRefsAreStoredOnEventsRumorsAndPerceptions(t *testing.T) {
	character, err := NewCharacter("hero", "Ash")
	if err != nil {
		t.Fatal(err)
	}
	world := NewWorld(character)

	item, err := NewCraftedItem("glass-key", "Glass Key", "crafter-1", 2, 5)
	if err != nil {
		t.Fatal(err)
	}
	if err := world.CarryItem(character.ID, item); err != nil {
		t.Fatal(err)
	}
	if _, err := world.SpawnLootDungeon("mirror-cache", "Mirror Cache", DefaultAreaID, 2); err != nil {
		t.Fatal(err)
	}
	dungeon, err := world.KillCharacterByID(character.ID, "mirror hall", DefaultAreaID)
	if err != nil {
		t.Fatal(err)
	}

	events := world.EventsForTarget(string(dungeon.ID))
	if len(events) != 2 {
		t.Fatalf("dungeon event count = %d, want 2", len(events))
	}
	for _, event := range events {
		if event.Object.Kind != TargetDungeon {
			t.Fatalf("event object kind = %q, want %q", event.Object.Kind, TargetDungeon)
		}
	}

	rumor, err := NewRumor("rumor-1", "source-1", "The cache is under glass.", 1, 5, time.Unix(1, 0).UTC())
	if err != nil {
		t.Fatal(err)
	}
	rumor.Subject = DungeonRef(dungeon)
	if err := world.AddRumor(rumor); err != nil {
		t.Fatal(err)
	}
	if _, err := world.SpreadRumor(rumor.ID, "listener-1", 0); err != nil {
		t.Fatal(err)
	}

	rumors := world.RumorsAbout(string(dungeon.ID))
	if len(rumors) != 1 {
		t.Fatalf("rumors about dungeon = %d, want 1", len(rumors))
	}
	perception := world.Perception("listener-1", string(dungeon.ID))
	if perception.Target.Kind != TargetDungeon {
		t.Fatalf("perception target kind = %q, want %q", perception.Target.Kind, TargetDungeon)
	}
}

func TestTimelineQueryHelpersReturnClones(t *testing.T) {
	character, err := NewCharacter("hero", "Ash")
	if err != nil {
		t.Fatal(err)
	}
	world := NewWorld(character)
	world.record(Event{
		Type:      EventRespawned,
		Subject:   CharacterRef(character),
		SubjectID: string(character.ID),
		Data:      map[string]string{"mutable": "original"},
	})

	events := world.EventsForActor(ActorID(character.ID))
	if len(events) != 1 {
		t.Fatalf("events for actor = %d, want 1", len(events))
	}
	events[0].Data["mutable"] = "changed"

	again := world.EventsForActor(ActorID(character.ID))
	if again[0].Data["mutable"] != "original" {
		t.Fatalf("query returned aliased event data")
	}
}

func TestPerceptionQueries(t *testing.T) {
	character, err := NewCharacter("hero", "Ash")
	if err != nil {
		t.Fatal(err)
	}
	world := NewWorld(character)

	if _, err := world.AdjustPerception("observer-1", "target-1", PerceptionDelta{Trust: 5}); err != nil {
		t.Fatal(err)
	}
	if _, err := world.AdjustPerception("observer-2", "target-1", PerceptionDelta{Fear: 3}); err != nil {
		t.Fatal(err)
	}

	if got := len(world.PerceptionsByObserver("observer-1")); got != 1 {
		t.Fatalf("perceptions by observer = %d, want 1", got)
	}
	if got := len(world.PerceptionsAbout("target-1")); got != 2 {
		t.Fatalf("perceptions about target = %d, want 2", got)
	}
}
