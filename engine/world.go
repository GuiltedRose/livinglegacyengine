package engine

import (
	"fmt"
	"time"
)

// EventType identifies a structured world event category.
type EventType string

const (
	// EventItemCrafted records a crafted item entering a character inventory.
	EventItemCrafted EventType = "item_crafted"
	// EventCharacterDied records a character death that successfully deposited loot.
	EventCharacterDied EventType = "character_died"
	// EventLootDeposited records eligible crafted loot entering an area dungeon.
	EventLootDeposited EventType = "loot_deposited"
	// EventDungeonSpawned records a pre-spawned loot dungeon being added to the world.
	EventDungeonSpawned EventType = "dungeon_spawned"
	// EventDungeonLooted records a dungeon clear and reward claim.
	EventDungeonLooted EventType = "dungeon_looted"
	// EventDungeonLocked records a dungeon locking to the actor who cleared it.
	EventDungeonLocked EventType = "dungeon_locked"
	// EventDungeonUnlocked records a locked dungeon reopening after its owner dies with eligible loot.
	EventDungeonUnlocked EventType = "dungeon_unlocked"
	// EventDungeonDecayed records a sealed dungeon decaying by rule policy.
	EventDungeonDecayed EventType = "dungeon_decayed"
	// EventRespawned records a character returning after death.
	EventRespawned EventType = "character_respawned"
)

// Event is a structured timeline entry recorded by the engine or host game.
type Event struct {
	Type        EventType         `json:"type"`
	At          time.Time         `json:"at"`
	Description string            `json:"description"`
	SubjectID   string            `json:"subject_id,omitempty"`
	ObjectID    string            `json:"object_id,omitempty"`
	Subject     TargetRef         `json:"subject,omitempty"`
	Object      TargetRef         `json:"object,omitempty"`
	Data        map[string]string `json:"data,omitempty"`
}

// World contains all mutable engine state for one simulated world.
type World struct {
	PrimaryCharacter Character
	Characters       map[CharacterID]Character
	Dungeons         map[DungeonID]LootDungeon
	Crafters         map[ActorID]int
	Actors           map[ActorID]Actor
	Rumors           []Rumor
	Memories         map[ActorID]Memory
	Perceptions      map[PerceptionKey]Perception
	Events           []Event

	now       func() time.Time
	nextRunID int
	rules     Rules
	indices   worldIndices
}

// NewWorld creates a world with DefaultRules and a primary character.
func NewWorld(character Character) *World {
	return NewWorldWithRules(character, DefaultRules())
}

// NewWorldWithRules creates a world with host-provided rules.
func NewWorldWithRules(character Character, rules Rules) *World {
	return &World{
		PrimaryCharacter: character,
		Characters:       map[CharacterID]Character{character.ID: cloneCharacter(character)},
		Dungeons:         make(map[DungeonID]LootDungeon),
		Crafters:         make(map[ActorID]int),
		Actors:           make(map[ActorID]Actor),
		Memories:         make(map[ActorID]Memory),
		Perceptions:      make(map[PerceptionKey]Perception),
		now:              time.Now,
		rules:            rules.normalized(),
		indices:          newWorldIndices(),
	}
}

// SetClock replaces the world's time source when now is non-nil.
func (w *World) SetClock(now func() time.Time) {
	if now != nil {
		w.now = now
	}
}

// SetRules replaces runtime policy hooks after normalizing missing hooks to defaults.
func (w *World) SetRules(rules Rules) {
	w.rules = rules.normalized()
}

// SpawnLootDungeon creates a dormant area dungeon during world setup.
func (w *World) SpawnLootDungeon(id DungeonID, name string, areaID AreaID, depth int) (LootDungeon, error) {
	dungeon, err := NewSpawnedLootDungeon(id, name, areaID, depth, w.now())
	if err != nil {
		return LootDungeon{}, err
	}
	if _, exists := w.Dungeons[dungeon.ID]; exists {
		return LootDungeon{}, fmt.Errorf("%w: %s", ErrDuplicateDungeon, dungeon.ID)
	}
	w.Dungeons[dungeon.ID] = dungeon
	w.record(Event{
		Type:        EventDungeonSpawned,
		Description: fmt.Sprintf("%s spawned in area %s", dungeon.Name, dungeon.AreaID),
		ObjectID:    string(dungeon.ID),
		Object:      DungeonRef(dungeon),
		Data: map[string]string{
			"area_id": string(dungeon.AreaID),
		},
	})
	return dungeon, nil
}

// DepositDeathLoot deposits eligible crafted items into the area's spawned dungeon.
func (w *World) DepositDeathLoot(actorID ActorID, areaID AreaID, items []CraftedItem, cause string) (LootDungeon, error) {
	if actorID == "" {
		return LootDungeon{}, fmt.Errorf("%w", ErrActorRequired)
	}
	if areaID == "" {
		areaID = DefaultAreaID
	}
	if cause == "" {
		cause = "unknown danger"
	}

	dungeonID, dungeon, ok := w.firstDungeonInArea(areaID)
	if !ok {
		return LootDungeon{}, fmt.Errorf("%w: %s", ErrNoLootDungeonInArea, areaID)
	}
	if dungeon.Status == DungeonDecayed {
		return LootDungeon{}, fmt.Errorf("%w: %s", ErrDungeonDecayed, dungeon.ID)
	}

	eligibleItems := make([]CraftedItem, 0, len(items))
	depositValue := 0
	for _, item := range items {
		if !w.rules.EligibleForDungeon(item) {
			continue
		}
		itemValue := w.rules.ItemLegacyValue(item)
		if itemValue < 0 {
			return LootDungeon{}, fmt.Errorf("item %q has negative legacy value", item.ID)
		}
		eligibleItems = append(eligibleItems, item)
		depositValue += itemValue
	}
	if len(eligibleItems) == 0 {
		return LootDungeon{}, fmt.Errorf("%w", ErrNoEligibleLoot)
	}

	w.unlockActorDungeons(actorID)
	dungeonID, dungeon, ok = w.firstDungeonInArea(areaID)
	if !ok {
		return LootDungeon{}, fmt.Errorf("%w: %s", ErrNoLootDungeonInArea, areaID)
	}
	for _, item := range eligibleItems {
		deposit := DepositedLoot{
			ID:          newLootDepositID(dungeon.ID, actorID, len(dungeon.Deposits)+1),
			Item:        item,
			DroppedBy:   actorID,
			AreaID:      areaID,
			DungeonID:   dungeon.ID,
			Cause:       cause,
			DepositedAt: w.now(),
			Attributes:  map[string]string{},
		}
		dungeon.Deposits = append(dungeon.Deposits, deposit)
		dungeon.Items = append(dungeon.Items, item)
		w.applyDepositHooks(deposit)
	}
	dungeon.LegacyValue += depositValue
	if dungeon.Status != DungeonLocked {
		dungeon.Status = DungeonActive
	}
	w.Dungeons[dungeonID] = dungeon
	w.record(Event{
		Type:        EventLootDeposited,
		Description: fmt.Sprintf("%s deposited %d crafted items into %s", actorID, len(eligibleItems), dungeon.Name),
		SubjectID:   string(actorID),
		ObjectID:    string(dungeon.ID),
		Subject:     NewTargetRef(string(actorID), TargetActor, ""),
		Object:      DungeonRef(dungeon),
		Data: map[string]string{
			"area_id": string(areaID),
			"cause":   cause,
		},
	})
	return dungeon, nil
}

// LootDungeon claims one reward from an active dungeon and locks it to the looter.
func (w *World) LootDungeon(id DungeonID, looter ActorID) ([]CraftedItem, error) {
	if looter == "" {
		return nil, fmt.Errorf("%w", ErrActorRequired)
	}
	dungeon, ok := w.Dungeons[id]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrDungeonNotFound, id)
	}
	if dungeon.Status == DungeonDormant {
		return nil, fmt.Errorf("%w: %s", ErrDungeonDormant, id)
	}
	if dungeon.Status == DungeonLocked {
		return nil, fmt.Errorf("%w: %s locked to %s", ErrDungeonLocked, id, dungeon.LockedTo)
	}
	if dungeon.Status != DungeonActive {
		return nil, fmt.Errorf("dungeon %q is %s", id, dungeon.Status)
	}
	if len(dungeon.Items) == 0 {
		dungeon.Status = DungeonDormant
		w.Dungeons[id] = dungeon
		return nil, fmt.Errorf("dungeon %q has no deposited loot", id)
	}

	index := w.rules.SelectLoot(dungeon, looter)
	if index < 0 || index >= len(dungeon.Items) {
		return nil, fmt.Errorf("dungeon %q has no deposited loot", id)
	}
	depositIndex := dungeon.unclaimedDepositIndex(index)
	if depositIndex < 0 {
		return nil, fmt.Errorf("dungeon %q has no deposited loot", id)
	}
	sourceDeposit := dungeon.Deposits[depositIndex]
	source := sourceDeposit.Item
	reward := w.rules.LootReward(source)
	if reward.Rarity < source.Rarity {
		return nil, fmt.Errorf("%w: reward rarity %d is lower than dropped item rarity %d", ErrRewardDowngrade, reward.Rarity, source.Rarity)
	}
	loot := []CraftedItem{reward}
	lootValue := w.rules.ItemLegacyValue(source)
	dungeon.Deposits[depositIndex].ClaimedBy = looter
	dungeon.Deposits[depositIndex].ClaimedAt = w.now()
	claimedDeposit := dungeon.Deposits[depositIndex]
	dungeon.Items = append(dungeon.Items[:index], dungeon.Items[index+1:]...)
	dungeon.LegacyValue -= lootValue
	if dungeon.LegacyValue < 0 {
		dungeon.LegacyValue = 0
	}
	dungeon.Status = DungeonLocked
	dungeon.LootedBy = looter
	dungeon.LockedTo = looter
	w.Dungeons[id] = dungeon
	w.applyLootToCharacter(looter, loot, lootValue)
	w.record(Event{
		Type:        EventDungeonLooted,
		Description: fmt.Sprintf("%s looted %s; legacy increased by %d", looter, dungeon.Name, lootValue),
		SubjectID:   string(looter),
		ObjectID:    string(dungeon.ID),
		Subject:     NewTargetRef(string(looter), TargetActor, ""),
		Object:      DungeonRef(dungeon),
		Data: map[string]string{
			"dungeon_name": dungeon.Name,
		},
	})
	w.record(Event{
		Type:        EventDungeonLocked,
		Description: fmt.Sprintf("%s locked to %s until their next death", dungeon.Name, looter),
		SubjectID:   string(looter),
		ObjectID:    string(dungeon.ID),
		Subject:     NewTargetRef(string(looter), TargetActor, ""),
		Object:      DungeonRef(dungeon),
	})
	w.applyClaimHooks(claimedDeposit, reward)
	return cloneItems(loot), nil
}

func (w *World) applyDepositHooks(deposit DepositedLoot) {
	for _, event := range w.rules.DepositEvents(deposit) {
		w.record(event)
	}
	for _, rumor := range w.rules.DepositRumors(deposit) {
		if err := w.AddRumor(rumor); err != nil {
			continue
		}
	}
}

func (w *World) applyClaimHooks(deposit DepositedLoot, reward CraftedItem) {
	for _, event := range w.rules.ClaimEvents(deposit, reward) {
		w.record(event)
	}
	for _, rumor := range w.rules.ClaimRumors(deposit, reward) {
		if err := w.AddRumor(rumor); err != nil {
			continue
		}
	}
}

// DecayDungeons applies the configured decay rule to sealed dungeons.
func (w *World) DecayDungeons() []DungeonID {
	now := w.now()
	decayed := []DungeonID{}
	for id, dungeon := range w.Dungeons {
		if dungeon.Status != DungeonSealed {
			continue
		}
		if !w.rules.ShouldDecayDungeon(dungeon, now) {
			continue
		}
		dungeon.Status = DungeonDecayed
		w.Dungeons[id] = dungeon
		decayed = append(decayed, id)
		w.record(Event{
			Type:        EventDungeonDecayed,
			Description: fmt.Sprintf("%s decayed", dungeon.Name),
			ObjectID:    string(dungeon.ID),
			Object:      DungeonRef(dungeon),
			Data: map[string]string{
				"dungeon_name": dungeon.Name,
			},
		})
	}
	return decayed
}

func (w *World) firstDungeonInArea(areaID AreaID) (DungeonID, LootDungeon, bool) {
	for id, dungeon := range w.Dungeons {
		if dungeon.AreaID == areaID && dungeon.Status != DungeonDecayed {
			return id, dungeon, true
		}
	}
	return "", LootDungeon{}, false
}

func (w *World) unlockActorDungeons(actorID ActorID) {
	for id, dungeon := range w.Dungeons {
		if dungeon.Status != DungeonLocked || dungeon.LockedTo != actorID {
			continue
		}
		dungeon.LockedTo = ""
		if len(dungeon.Items) == 0 {
			dungeon.Status = DungeonDormant
		} else {
			dungeon.Status = DungeonActive
		}
		w.Dungeons[id] = dungeon
		w.record(Event{
			Type:        EventDungeonUnlocked,
			Description: fmt.Sprintf("%s unlocked after %s died again", dungeon.Name, actorID),
			SubjectID:   string(actorID),
			ObjectID:    string(dungeon.ID),
			Subject:     NewTargetRef(string(actorID), TargetActor, ""),
			Object:      DungeonRef(dungeon),
		})
	}
}

func (w *World) record(event Event) {
	event = w.normalizeEvent(event)
	w.Events = append(w.Events, event)
	w.indexEvent(len(w.Events)-1, event)
}

func (w *World) normalizeEvent(event Event) Event {
	if event.At.IsZero() {
		event.At = w.now()
	}
	event.Subject = event.Subject.normalized(TargetCustom)
	event.Object = event.Object.normalized(TargetCustom)
	event.Data = cloneStringMap(event.Data)
	return event
}
