package engine

import (
	"testing"
	"time"
)

func TestAdjustPerceptionIsScopedByObserverAndTarget(t *testing.T) {
	character, err := NewCharacter("hero", "Ash")
	if err != nil {
		t.Fatal(err)
	}
	world := NewWorld(character)

	first, err := world.AdjustPerception("observer-1", "target-1", PerceptionDelta{
		Trust:      15,
		Fear:       200,
		Confidence: 20,
		Attributes: map[string]string{
			"reason": "witnessed rescue",
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if first.Fear != 100 {
		t.Fatalf("fear = %d, want clamped 100", first.Fear)
	}
	if first.Attributes["reason"] != "witnessed rescue" {
		t.Fatalf("reason attribute = %q, want witnessed rescue", first.Attributes["reason"])
	}
	second := world.Perception("observer-2", "target-1")
	if second.Trust != 0 || second.Fear != 0 {
		t.Fatalf("second observer perception = %#v, want neutral", second)
	}
}

func TestSpreadRumorAppliesDefaultPerception(t *testing.T) {
	character, err := NewCharacter("hero", "Ash")
	if err != nil {
		t.Fatal(err)
	}
	world := NewWorld(character)
	world.SetClock(func() time.Time { return time.Unix(10, 0).UTC() })

	rumor, err := NewRumor("rumor-1", "source-1", "Ash opened a sealed gate.", 0.75, 12, world.now())
	if err != nil {
		t.Fatal(err)
	}
	rumor.SubjectID = string(character.ID)
	if err := world.AddRumor(rumor); err != nil {
		t.Fatal(err)
	}
	if _, err := world.SpreadRumor(rumor.ID, "listener-1", 0.25); err != nil {
		t.Fatal(err)
	}

	perception := world.Perception("listener-1", string(character.ID))
	if perception.Notoriety != 12 {
		t.Fatalf("notoriety = %d, want 12", perception.Notoriety)
	}
	if perception.Confidence != 50 {
		t.Fatalf("confidence = %d, want 50", perception.Confidence)
	}
}

func TestCustomPerceptionRules(t *testing.T) {
	character, err := NewCharacter("hero", "Ash")
	if err != nil {
		t.Fatal(err)
	}

	rules := DefaultRules()
	rules.RumorPerception = func(rumor Rumor) PerceptionDelta {
		return PerceptionDelta{
			Trust:      -rumor.Impact,
			Fear:       rumor.Impact * 2,
			Confidence: 7,
		}
	}
	rules.EventPerception = func(event Event) PerceptionDelta {
		return PerceptionDelta{
			Respect:    20,
			Confidence: 100,
			Attributes: map[string]string{"event_type": string(event.Type)},
		}
	}
	world := NewWorldWithRules(character, rules)

	rumor, err := NewRumor("rumor-1", "source-1", "Ash angered the hill.", 1, 3, time.Unix(1, 0).UTC())
	if err != nil {
		t.Fatal(err)
	}
	rumor.SubjectID = string(character.ID)
	if err := world.AddRumor(rumor); err != nil {
		t.Fatal(err)
	}
	if _, err := world.SpreadRumor(rumor.ID, "listener-1", 0); err != nil {
		t.Fatal(err)
	}

	perception := world.Perception("listener-1", string(character.ID))
	if perception.Trust != -3 || perception.Fear != 6 || perception.Confidence != 7 {
		t.Fatalf("custom rumor perception = %#v", perception)
	}

	event := Event{
		Type:        EventDungeonLooted,
		SubjectID:   string(character.ID),
		Description: "Ash cleared a dungeon.",
	}
	perception, err = world.ApplyEventToPerception("listener-1", event)
	if err != nil {
		t.Fatal(err)
	}
	if perception.Respect != 20 {
		t.Fatalf("respect = %d, want 20", perception.Respect)
	}
	if perception.Attributes["event_type"] != string(EventDungeonLooted) {
		t.Fatalf("event_type = %q, want dungeon event", perception.Attributes["event_type"])
	}
}
