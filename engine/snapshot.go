package engine

import "fmt"

// CurrentSnapshotVersion is the newest snapshot schema version this engine supports.
const CurrentSnapshotVersion = 1

// WorldSnapshot is portable serializable state for a World.
type WorldSnapshot struct {
	Version          int                       `json:"version"`
	PrimaryCharacter Character                 `json:"primary_character"`
	Character        Character                 `json:"character"`
	Characters       map[CharacterID]Character `json:"characters,omitempty"`
	Dungeons         map[DungeonID]LootDungeon `json:"dungeons,omitempty"`
	Crafters         map[ActorID]int           `json:"crafters,omitempty"`
	Actors           map[ActorID]Actor         `json:"actors,omitempty"`
	Rumors           []Rumor                   `json:"rumors,omitempty"`
	Memories         map[ActorID]Memory        `json:"memories,omitempty"`
	Perceptions      []Perception              `json:"perceptions,omitempty"`
	Events           []Event                   `json:"events,omitempty"`
	NextRunID        int                       `json:"next_run_id"`
}

// Snapshot returns a cloned serializable view of the world.
func (w *World) Snapshot() WorldSnapshot {
	return WorldSnapshot{
		Version:          CurrentSnapshotVersion,
		PrimaryCharacter: cloneCharacter(w.PrimaryCharacter),
		Character:        cloneCharacter(w.PrimaryCharacter),
		Characters:       cloneCharacters(w.Characters),
		Dungeons:         cloneDungeons(w.Dungeons),
		Crafters:         cloneActorScores(w.Crafters),
		Actors:           cloneActors(w.Actors),
		Rumors:           cloneRumors(w.Rumors),
		Memories:         cloneMemories(w.Memories),
		Perceptions:      clonePerceptionList(w.Perceptions),
		Events:           cloneEvents(w.Events),
		NextRunID:        w.nextRunID,
	}
}

// RestoreWorld validates, migrates, and hydrates a World from a snapshot.
func RestoreWorld(snapshot WorldSnapshot) (*World, error) {
	snapshot, err := MigrateSnapshot(snapshot)
	if err != nil {
		return nil, err
	}
	if err := ValidateSnapshot(snapshot); err != nil {
		return nil, err
	}
	primary := snapshot.PrimaryCharacter
	if primary.ID == "" {
		primary = snapshot.Character
	}
	world := NewWorld(cloneCharacter(primary))
	world.Characters = cloneCharacters(snapshot.Characters)
	if len(world.Characters) == 0 {
		world.Characters = map[CharacterID]Character{world.PrimaryCharacter.ID: cloneCharacter(world.PrimaryCharacter)}
	}
	if registryPrimary, ok := world.Characters[world.PrimaryCharacter.ID]; ok {
		world.PrimaryCharacter = cloneCharacter(registryPrimary)
	}
	world.Dungeons = cloneDungeons(snapshot.Dungeons)
	world.Crafters = cloneActorScores(snapshot.Crafters)
	world.Actors = cloneActors(snapshot.Actors)
	world.Rumors = cloneRumors(snapshot.Rumors)
	world.Memories = cloneMemories(snapshot.Memories)
	world.Perceptions = perceptionsFromList(snapshot.Perceptions)
	world.Events = cloneEvents(snapshot.Events)
	world.nextRunID = snapshot.NextRunID
	world.rebuildIndices()
	return world, nil
}

// ValidateSnapshot rejects missing, unsupported, or future snapshot schemas.
func ValidateSnapshot(snapshot WorldSnapshot) error {
	if snapshot.Version <= 0 {
		return fmt.Errorf("snapshot version is required")
	}
	if snapshot.Version > CurrentSnapshotVersion {
		return fmt.Errorf("snapshot version %d is newer than supported version %d", snapshot.Version, CurrentSnapshotVersion)
	}
	if snapshot.PrimaryCharacter.ID == "" && snapshot.Character.ID == "" {
		return fmt.Errorf("snapshot character id is required")
	}
	return nil
}

// MigrateSnapshot upgrades older snapshot shapes to CurrentSnapshotVersion.
func MigrateSnapshot(snapshot WorldSnapshot) (WorldSnapshot, error) {
	switch snapshot.Version {
	case 0:
		snapshot.Version = CurrentSnapshotVersion
		if snapshot.PrimaryCharacter.ID == "" {
			snapshot.PrimaryCharacter = snapshot.Character
		}
		return snapshot, nil
	case CurrentSnapshotVersion:
		if snapshot.PrimaryCharacter.ID == "" {
			snapshot.PrimaryCharacter = snapshot.Character
		}
		return snapshot, nil
	default:
		if snapshot.Version > CurrentSnapshotVersion {
			return WorldSnapshot{}, fmt.Errorf("snapshot version %d is newer than supported version %d", snapshot.Version, CurrentSnapshotVersion)
		}
		return WorldSnapshot{}, fmt.Errorf("snapshot version %d is not supported", snapshot.Version)
	}
}

func cloneCharacter(character Character) Character {
	character.Inventory = cloneItems(character.Inventory)
	character.RecoveredItems = cloneItems(character.RecoveredItems)
	return character
}

func cloneCharacters(characters map[CharacterID]Character) map[CharacterID]Character {
	cloned := make(map[CharacterID]Character, len(characters))
	for id, character := range characters {
		cloned[id] = cloneCharacter(character)
	}
	return cloned
}

func cloneItems(items []CraftedItem) []CraftedItem {
	if items == nil {
		return nil
	}
	cloned := make([]CraftedItem, len(items))
	for i, item := range items {
		cloned[i] = item
		cloned[i].Tags = append([]string(nil), item.Tags...)
		cloned[i].Attributes = cloneStringMap(item.Attributes)
	}
	return cloned
}

func cloneDungeons(dungeons map[DungeonID]LootDungeon) map[DungeonID]LootDungeon {
	cloned := make(map[DungeonID]LootDungeon, len(dungeons))
	for id, dungeon := range dungeons {
		dungeon.Items = cloneItems(dungeon.Items)
		dungeon.Deposits = cloneDeposits(dungeon.Deposits)
		dungeon.syncDepositsFromItems(dungeon.CreatedAt)
		dungeon.syncItemsFromDeposits()
		dungeon.Attributes = cloneStringMap(dungeon.Attributes)
		cloned[id] = dungeon
	}
	return cloned
}

func cloneDeposits(deposits []DepositedLoot) []DepositedLoot {
	if deposits == nil {
		return nil
	}
	cloned := make([]DepositedLoot, len(deposits))
	for i, deposit := range deposits {
		cloned[i] = cloneDeposit(deposit)
	}
	return cloned
}

func cloneDeposit(deposit DepositedLoot) DepositedLoot {
	deposit.Item = cloneItems([]CraftedItem{deposit.Item})[0]
	deposit.Attributes = cloneStringMap(deposit.Attributes)
	return deposit
}

func cloneActorScores(scores map[ActorID]int) map[ActorID]int {
	cloned := make(map[ActorID]int, len(scores))
	for id, score := range scores {
		cloned[id] = score
	}
	return cloned
}

func cloneEvents(events []Event) []Event {
	if events == nil {
		return nil
	}
	cloned := make([]Event, len(events))
	for i, event := range events {
		cloned[i] = cloneEvent(event)
	}
	return cloned
}

func cloneEvent(event Event) Event {
	event.Subject = cloneTargetRef(event.Subject)
	event.Object = cloneTargetRef(event.Object)
	event.Data = cloneStringMap(event.Data)
	return event
}

func cloneActors(actors map[ActorID]Actor) map[ActorID]Actor {
	cloned := make(map[ActorID]Actor, len(actors))
	for id, actor := range actors {
		actor.Attributes = cloneStringMap(actor.Attributes)
		cloned[id] = actor
	}
	return cloned
}

func cloneRumors(rumors []Rumor) []Rumor {
	if rumors == nil {
		return nil
	}
	cloned := make([]Rumor, len(rumors))
	for i, rumor := range rumors {
		cloned[i] = cloneRumor(rumor)
	}
	return cloned
}

func cloneRumor(rumor Rumor) Rumor {
	rumor.Subject = cloneTargetRef(rumor.Subject)
	rumor.Object = cloneTargetRef(rumor.Object)
	rumor.Attributes = cloneStringMap(rumor.Attributes)
	return rumor
}

func cloneMemories(memories map[ActorID]Memory) map[ActorID]Memory {
	cloned := make(map[ActorID]Memory, len(memories))
	for id, memory := range memories {
		cloned[id] = cloneMemory(memory)
	}
	return cloned
}

func cloneMemory(memory Memory) Memory {
	memory.KnownEvents = cloneEvents(memory.KnownEvents)
	memory.KnownRumors = cloneRumors(memory.KnownRumors)
	memory.Attributes = cloneStringMap(memory.Attributes)
	return memory
}

func clonePerceptionList(perceptions map[PerceptionKey]Perception) []Perception {
	if len(perceptions) == 0 {
		return nil
	}
	cloned := make([]Perception, 0, len(perceptions))
	for _, perception := range perceptions {
		cloned = append(cloned, clonePerception(perception))
	}
	return cloned
}

func perceptionsFromList(perceptions []Perception) map[PerceptionKey]Perception {
	cloned := make(map[PerceptionKey]Perception, len(perceptions))
	for _, perception := range perceptions {
		perception = clonePerception(perception)
		cloned[perceptionKey(perception.ObserverID, perception.TargetID)] = perception
	}
	return cloned
}

func clonePerception(perception Perception) Perception {
	perception.Target = cloneTargetRef(perception.Target)
	perception.Attributes = cloneStringMap(perception.Attributes)
	return perception
}

func cloneTargetRef(target TargetRef) TargetRef {
	target.Attributes = cloneStringMap(target.Attributes)
	return target
}

func cloneStringMap(values map[string]string) map[string]string {
	if values == nil {
		return nil
	}
	cloned := make(map[string]string, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}
