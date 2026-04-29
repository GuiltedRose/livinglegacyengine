package engine

import "testing"

func TestSnapshotIncludesCurrentVersion(t *testing.T) {
	character, err := NewCharacter("hero", "Ash")
	if err != nil {
		t.Fatal(err)
	}
	world := NewWorld(character)

	snapshot := world.Snapshot()
	if snapshot.Version != CurrentSnapshotVersion {
		t.Fatalf("snapshot version = %d, want %d", snapshot.Version, CurrentSnapshotVersion)
	}
}

func TestRestoreMigratesLegacyMissingVersionSnapshot(t *testing.T) {
	character, err := NewCharacter("hero", "Ash")
	if err != nil {
		t.Fatal(err)
	}

	snapshot := WorldSnapshot{
		Character: character,
	}
	restored, err := RestoreWorld(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	if restored.PrimaryCharacter.ID != character.ID {
		t.Fatalf("restored character = %q, want %q", restored.PrimaryCharacter.ID, character.ID)
	}
}

func TestValidateSnapshotRejectsFutureVersion(t *testing.T) {
	character, err := NewCharacter("hero", "Ash")
	if err != nil {
		t.Fatal(err)
	}

	_, err = RestoreWorld(WorldSnapshot{
		Version:   CurrentSnapshotVersion + 1,
		Character: character,
	})
	if err == nil {
		t.Fatal("expected future snapshot version to fail")
	}
}

func TestRestoreRebuildsQueryIndices(t *testing.T) {
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
	rumor, err := NewRumor("rumor-1", "source-1", "The cache is under glass.", 1, 5, world.now())
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

	restored, err := RestoreWorld(world.Snapshot())
	if err != nil {
		t.Fatal(err)
	}

	if got := len(restored.EventsForTarget(string(dungeon.ID))); got != 2 {
		t.Fatalf("restored events for dungeon = %d, want 2", got)
	}
	if got := len(restored.RumorsAbout(string(dungeon.ID))); got != 1 {
		t.Fatalf("restored rumors about dungeon = %d, want 1", got)
	}
	if got := len(restored.PerceptionsAbout(string(dungeon.ID))); got != 1 {
		t.Fatalf("restored perceptions about dungeon = %d, want 1", got)
	}
}
