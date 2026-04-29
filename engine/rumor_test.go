package engine

import (
	"testing"
	"time"
)

func TestActorRumorMemoryFlow(t *testing.T) {
	character, err := NewCharacter("hero", "Ash")
	if err != nil {
		t.Fatal(err)
	}
	world := NewWorld(character)
	world.SetClock(func() time.Time { return time.Unix(10, 0).UTC() })

	source, err := NewActor("source", "Source", ActorUnknown)
	if err != nil {
		t.Fatal(err)
	}
	if err := world.AddActor(source); err != nil {
		t.Fatal(err)
	}

	rumor, err := NewRumor("rumor-1", source.ID, "A sealed cache appeared.", 0.75, 3, world.now())
	if err != nil {
		t.Fatal(err)
	}
	if err := world.AddRumor(rumor); err != nil {
		t.Fatal(err)
	}

	if len(world.Memory(source.ID).KnownRumors) != 1 {
		t.Fatalf("source memory rumor count = %d, want 1", len(world.Memory(source.ID).KnownRumors))
	}

	spread, err := world.SpreadRumor(rumor.ID, "listener", 0.25)
	if err != nil {
		t.Fatal(err)
	}
	if spread.Truth != 0.5 {
		t.Fatalf("spread truth = %f, want 0.5", spread.Truth)
	}
	if spread.Spread != 1 {
		t.Fatalf("spread count = %d, want 1", spread.Spread)
	}
	if len(world.Memory("listener").KnownRumors) != 1 {
		t.Fatalf("listener memory rumor count = %d, want 1", len(world.Memory("listener").KnownRumors))
	}
}
