package engine

import (
	"strings"
	"testing"
	"time"
)

func TestGenerateRumorsFromEventsUsesEventsAndLore(t *testing.T) {
	character, err := NewCharacter("mara", "Mara")
	if err != nil {
		t.Fatal(err)
	}
	rules := DefaultRules()
	rules.LoreFacts = func(context TextGenerationContext) []LoreFact {
		if context.TargetID != "ash-cache" {
			t.Fatalf("lore target = %q, want ash-cache", context.TargetID)
		}
		return []LoreFact{{
			ID:     "ash-cache-oath",
			Text:   "The stair keepers call this an unpaid oath.",
			Object: NewTargetRef("ash-cache", TargetDungeon, "Ash Cache"),
			Weight: 10,
		}}
	}
	world := NewWorldWithRules(character, rules)
	world.SetClock(func() time.Time { return time.Unix(10, 0).UTC() })

	dungeon, err := world.SpawnLootDungeon("ash-cache", "Ash Cache", "ash-road", 3)
	if err != nil {
		t.Fatal(err)
	}
	item, err := NewCraftedItem("iron-song", "Iron Song", "ada", 4, 12, "sword")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := world.DepositDeathLoot(ActorID(character.ID), dungeon.AreaID, []CraftedItem{item}, "ash storm"); err != nil {
		t.Fatal(err)
	}

	rumors, err := world.GenerateRumorsFromEvents(EventRumorOptions{
		SourceID: "witness",
		TargetID: string(dungeon.ID),
		Limit:    1,
		Truth:    0.8,
		Impact:   4,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rumors) != 1 {
		t.Fatalf("generated rumors = %d, want 1", len(rumors))
	}
	if rumors[0].Description == "Not here, try there" {
		t.Fatalf("generated stale placeholder text")
	}
	if !strings.Contains(rumors[0].Description, "unpaid oath") {
		t.Fatalf("generated rumor %q does not include lore", rumors[0].Description)
	}
	if rumors[0].Object.ID != string(dungeon.ID) {
		t.Fatalf("rumor object = %#v, want dungeon target", rumors[0].Object)
	}
	if rumors[0].Attributes["generated_from"] != "history" {
		t.Fatalf("generated_from = %q, want history", rumors[0].Attributes["generated_from"])
	}
}

func TestRecordEventWithGeneratedRumorRecordsAndCreatesRumor(t *testing.T) {
	character, err := NewCharacter("mara", "Mara")
	if err != nil {
		t.Fatal(err)
	}
	world := NewWorld(character)

	event, rumor, err := world.RecordEventWithGeneratedRumor(Event{
		Type:        EventRespawned,
		Description: "Mara returned under the old moon.",
		SubjectID:   string(character.ID),
		Subject:     CharacterRef(character),
	}, EventRumorOptions{
		SourceID: "witness",
		Truth:    0.9,
		Impact:   2,
	})
	if err != nil {
		t.Fatal(err)
	}
	if event.Type != EventRespawned {
		t.Fatalf("event type = %q, want respawned", event.Type)
	}
	if rumor.ID == "" {
		t.Fatalf("generated rumor id is empty")
	}
	if rumor.Subject.ID != string(character.ID) {
		t.Fatalf("rumor subject = %#v, want character", rumor.Subject)
	}
	if len(world.Events) != 1 {
		t.Fatalf("world events = %d, want 1", len(world.Events))
	}
	if len(world.Rumors) != 1 {
		t.Fatalf("world rumors = %d, want 1", len(world.Rumors))
	}
}

func TestGenerateRumorsFromEventsSkipsExistingGeneratedRumors(t *testing.T) {
	character, err := NewCharacter("mara", "Mara")
	if err != nil {
		t.Fatal(err)
	}
	world := NewWorld(character)
	if _, err := world.RecordEvent(Event{
		Type:        EventRespawned,
		Description: "Mara returned.",
		SubjectID:   string(character.ID),
		Subject:     CharacterRef(character),
	}); err != nil {
		t.Fatal(err)
	}

	first, err := world.GenerateRumorsFromEvents(EventRumorOptions{SourceID: "witness"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := world.GenerateRumorsFromEvents(EventRumorOptions{SourceID: "witness"})
	if err != nil {
		t.Fatal(err)
	}
	if len(first) != 1 {
		t.Fatalf("first generated rumors = %d, want 1", len(first))
	}
	if len(second) != 0 {
		t.Fatalf("second generated rumors = %d, want 0", len(second))
	}
}

func TestGenerateRumorsFromEventsUsesCustomTextRule(t *testing.T) {
	character, err := NewCharacter("mara", "Mara")
	if err != nil {
		t.Fatal(err)
	}
	rules := DefaultRules()
	rules.EventRumorText = func(context HistoryRumorTextContext) GeneratedText {
		return GeneratedText{
			Text:       "Custom chronicle: " + context.Event.Description,
			Attributes: map[string]string{"voice": "chronicle"},
		}
	}
	world := NewWorldWithRules(character, rules)
	if _, err := world.RecordEvent(Event{
		Type:        EventRespawned,
		Description: "Mara returned.",
		SubjectID:   string(character.ID),
		Subject:     CharacterRef(character),
	}); err != nil {
		t.Fatal(err)
	}

	rumors, err := world.GenerateRumorsFromEvents(EventRumorOptions{SourceID: "scribe"})
	if err != nil {
		t.Fatal(err)
	}
	if len(rumors) != 1 {
		t.Fatalf("generated rumors = %d, want 1", len(rumors))
	}
	if rumors[0].Description != "Custom chronicle: Mara returned." {
		t.Fatalf("description = %q", rumors[0].Description)
	}
	if rumors[0].Attributes["voice"] != "chronicle" {
		t.Fatalf("voice attribute = %q", rumors[0].Attributes["voice"])
	}
}

func TestGenerateRemarkFallsBackToHistory(t *testing.T) {
	character, err := NewCharacter("mara", "Mara")
	if err != nil {
		t.Fatal(err)
	}
	world := NewWorld(character)
	if _, err := world.RecordEvent(Event{
		Type:        EventRespawned,
		Description: "Mara returned.",
		SubjectID:   string(character.ID),
		Subject:     CharacterRef(character),
	}); err != nil {
		t.Fatal(err)
	}

	remark := world.GenerateRemark(TextGenerationContext{TargetID: string(character.ID)})
	if remark.Text == "" {
		t.Fatalf("remark text is empty")
	}
	if strings.Contains(remark.Text, "No one has heard") {
		t.Fatalf("remark ignored available history: %q", remark.Text)
	}
}
