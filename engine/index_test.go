package engine

import "testing"

func TestIndexedQueriesMatchTargetAndTypeFilters(t *testing.T) {
	character, err := NewCharacter("hero", "Ash")
	if err != nil {
		t.Fatal(err)
	}
	world := NewWorld(character)

	target := NewTargetRef("region-1", TargetRegion, "North Road")
	if _, err := world.RecordEvent(Event{
		Type:        EventRespawned,
		Description: "Ash was seen on the road.",
		Subject:     target,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := world.RecordEvent(Event{
		Type:        EventDungeonLooted,
		Description: "Ash cleared the road cache.",
		Subject:     target,
	}); err != nil {
		t.Fatal(err)
	}

	events := world.EventsMatching(EventFilter{
		Type:     EventDungeonLooted,
		TargetID: target.ID,
	})
	if len(events) != 1 {
		t.Fatalf("filtered events = %d, want 1", len(events))
	}
	if events[0].Type != EventDungeonLooted {
		t.Fatalf("event type = %q, want %q", events[0].Type, EventDungeonLooted)
	}
}

func TestIndexedRumorQueriesUseSourceAndTarget(t *testing.T) {
	character, err := NewCharacter("hero", "Ash")
	if err != nil {
		t.Fatal(err)
	}
	world := NewWorld(character)

	first, err := NewRumor("rumor-1", "source-1", "The north road is haunted.", 1, 2, world.now())
	if err != nil {
		t.Fatal(err)
	}
	first.Subject = NewTargetRef("region-1", TargetRegion, "North Road")
	if err := world.AddRumor(first); err != nil {
		t.Fatal(err)
	}
	second, err := NewRumor("rumor-2", "source-2", "The north road is safe.", 1, 2, world.now())
	if err != nil {
		t.Fatal(err)
	}
	second.Subject = NewTargetRef("region-1", TargetRegion, "North Road")
	if err := world.AddRumor(second); err != nil {
		t.Fatal(err)
	}

	rumors := world.RumorsMatching(RumorFilter{
		SourceID: "source-1",
		TargetID: "region-1",
	})
	if len(rumors) != 1 {
		t.Fatalf("filtered rumors = %d, want 1", len(rumors))
	}
	if rumors[0].SourceID != "source-1" {
		t.Fatalf("rumor source = %q, want source-1", rumors[0].SourceID)
	}
}
