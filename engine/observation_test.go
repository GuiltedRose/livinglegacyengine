package engine

import (
	"testing"
	"time"
)

func TestRecordObservedEventTeachesMemoryAndAppliesPerception(t *testing.T) {
	character, err := NewCharacter("hero", "Ash")
	if err != nil {
		t.Fatal(err)
	}

	rules := DefaultRules()
	rules.EventPerception = func(event Event) PerceptionDelta {
		return PerceptionDelta{
			Respect:    9,
			Confidence: 80,
			Attributes: map[string]string{
				"learned_from": string(event.Type),
			},
		}
	}
	world := NewWorldWithRules(character, rules)
	world.SetClock(func() time.Time { return time.Unix(20, 0).UTC() })

	event := Event{
		Type:        EventDungeonLooted,
		Description: "Ash cleared a public dungeon.",
		Subject:     CharacterRef(character),
	}
	recorded, perceptions, err := world.RecordObservedEvent(event, "observer-1", "observer-2")
	if err != nil {
		t.Fatal(err)
	}

	if recorded.At != time.Unix(20, 0).UTC() {
		t.Fatalf("recorded time = %s, want clock time", recorded.At)
	}
	if len(perceptions) != 2 {
		t.Fatalf("perception count = %d, want 2", len(perceptions))
	}
	if len(world.MemoriesOf("observer-1").KnownEvents) != 1 {
		t.Fatalf("observer memory event count = %d, want 1", len(world.MemoriesOf("observer-1").KnownEvents))
	}
	perception := world.Perception("observer-1", string(character.ID))
	if perception.Respect != 9 || perception.Confidence != 80 {
		t.Fatalf("perception = %#v, want custom event delta", perception)
	}
	if perception.Attributes["learned_from"] != string(EventDungeonLooted) {
		t.Fatalf("learned_from = %q, want event type", perception.Attributes["learned_from"])
	}
}

func TestRecordEventPreservesExplicitTimeAndReturnsClone(t *testing.T) {
	character, err := NewCharacter("hero", "Ash")
	if err != nil {
		t.Fatal(err)
	}
	world := NewWorld(character)

	when := time.Unix(99, 0).UTC()
	recorded, err := world.RecordEvent(Event{
		Type:        EventRespawned,
		At:          when,
		Description: "Ash returned.",
		Subject:     CharacterRef(character),
		Data:        map[string]string{"mutable": "original"},
	})
	if err != nil {
		t.Fatal(err)
	}
	recorded.Data["mutable"] = "changed"

	if world.Events[0].At != when {
		t.Fatalf("event time = %s, want explicit time", world.Events[0].At)
	}
	if world.Events[0].Data["mutable"] != "original" {
		t.Fatalf("recorded event data was aliased")
	}
}

func TestObserveEventRequiresPerceptionTarget(t *testing.T) {
	character, err := NewCharacter("hero", "Ash")
	if err != nil {
		t.Fatal(err)
	}
	world := NewWorld(character)

	if _, err := world.ObserveEvent("observer-1", Event{
		Type:        EventRespawned,
		Description: "Untargeted event.",
	}); err == nil {
		t.Fatal("expected untargeted observed event to fail")
	}
}
