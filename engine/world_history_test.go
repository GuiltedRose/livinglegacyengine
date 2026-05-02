package engine

import (
	"strings"
	"testing"
	"time"
)

func TestGenerateWorldHistoryCreatesNPCsFactionsAndTies(t *testing.T) {
	character, err := NewCharacter("hero", "Ash")
	if err != nil {
		t.Fatal(err)
	}
	world := NewWorld(character)
	world.SetClock(func() time.Time { return time.Unix(100, 0).UTC() })

	result, err := world.GenerateWorldHistory(WorldGenerationOptions{
		Seed:         42,
		NPCCount:     3,
		FactionCount: 2,
		TieCount:     5,
		NPCNames:     []string{"Rin", "Mara", "Oren"},
		FactionNames: []string{"Lantern Compact", "Ash Stair Keepers"},
		Era:          "founding",
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Actors) != 5 {
		t.Fatalf("actors = %d, want 5", len(result.Actors))
	}
	if len(result.Ties) != 5 {
		t.Fatalf("ties = %d, want 5", len(result.Ties))
	}
	if len(result.History) != 5 {
		t.Fatalf("history = %d, want 5", len(result.History))
	}
	if world.Actors["npc-rin"].Kind != ActorNPC {
		t.Fatalf("npc-rin kind = %q, want npc", world.Actors["npc-rin"].Kind)
	}
	if world.Actors["faction-lantern-compact"].Kind != ActorFaction {
		t.Fatalf("faction kind = %q, want faction", world.Actors["faction-lantern-compact"].Kind)
	}
	if got := len(world.TiesForTarget("npc-rin")); got == 0 {
		t.Fatalf("ties for npc-rin = %d, want at least one", got)
	}
}

func TestGenerateRemarkUsesGeneratedWorldHistoryTies(t *testing.T) {
	character, err := NewCharacter("hero", "Ash")
	if err != nil {
		t.Fatal(err)
	}
	world := NewWorld(character)
	if _, err := world.GenerateWorldHistory(WorldGenerationOptions{
		Seed:         7,
		NPCCount:     1,
		FactionCount: 1,
		TieCount:     1,
		NPCNames:     []string{"Rin"},
		FactionNames: []string{"Lantern Compact"},
	}); err != nil {
		t.Fatal(err)
	}

	remark := world.GenerateRemark(TextGenerationContext{TargetID: "npc-rin"})
	if remark.Attributes["kind"] != "world_history_tie" {
		t.Fatalf("remark kind = %q, want world_history_tie", remark.Attributes["kind"])
	}
	if !strings.Contains(remark.Text, "Rin") {
		t.Fatalf("remark %q does not mention generated NPC", remark.Text)
	}
}

func TestWorldHistorySurvivesSnapshotRoundTrip(t *testing.T) {
	character, err := NewCharacter("hero", "Ash")
	if err != nil {
		t.Fatal(err)
	}
	world := NewWorld(character)
	if _, err := world.GenerateWorldHistory(WorldGenerationOptions{
		Seed:         9,
		NPCCount:     1,
		FactionCount: 1,
		TieCount:     1,
		NPCNames:     []string{"Rin"},
		FactionNames: []string{"Lantern Compact"},
	}); err != nil {
		t.Fatal(err)
	}

	restored, err := RestoreWorld(world.Snapshot())
	if err != nil {
		t.Fatal(err)
	}
	if len(restored.Ties) != 1 {
		t.Fatalf("restored ties = %d, want 1", len(restored.Ties))
	}
	if len(restored.WorldHistory) != 1 {
		t.Fatalf("restored world history = %d, want 1", len(restored.WorldHistory))
	}
	if restored.Actors["npc-rin"].Kind != ActorNPC {
		t.Fatalf("restored npc kind = %q, want npc", restored.Actors["npc-rin"].Kind)
	}
}
