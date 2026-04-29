package engine

// EventFilter selects events by type and/or referenced target ID.
type EventFilter struct {
	Type     EventType
	TargetID string
}

// RumorFilter selects rumors by source actor and/or referenced target ID.
type RumorFilter struct {
	SourceID ActorID
	TargetID string
}

// EventsMatching returns cloned events matching the filter.
func (w *World) EventsMatching(filter EventFilter) []Event {
	if filter.TargetID != "" {
		return w.eventsFromIndices(w.indices.eventsByTarget[filter.TargetID], filter)
	}
	events := []Event{}
	for _, event := range w.Events {
		if filter.Type != "" && event.Type != filter.Type {
			continue
		}
		if filter.TargetID != "" && !eventReferences(event, filter.TargetID) {
			continue
		}
		events = append(events, cloneEvent(event))
	}
	return events
}

// EventsForTarget returns cloned events that reference a target ID.
func (w *World) EventsForTarget(targetID string) []Event {
	return w.EventsMatching(EventFilter{TargetID: targetID})
}

// EventsForActor returns cloned events that reference an actor ID.
func (w *World) EventsForActor(actorID ActorID) []Event {
	return w.EventsForTarget(string(actorID))
}

// RumorsMatching returns cloned rumors matching the filter.
func (w *World) RumorsMatching(filter RumorFilter) []Rumor {
	if filter.SourceID != "" && filter.TargetID == "" {
		return w.rumorsFromIndices(w.indices.rumorsBySource[filter.SourceID], filter)
	}
	if filter.TargetID != "" {
		return w.rumorsFromIndices(w.indices.rumorsByTarget[filter.TargetID], filter)
	}
	rumors := []Rumor{}
	for _, rumor := range w.Rumors {
		if filter.SourceID != "" && rumor.SourceID != filter.SourceID {
			continue
		}
		if filter.TargetID != "" && !rumorReferences(rumor, filter.TargetID) {
			continue
		}
		rumors = append(rumors, cloneRumor(rumor))
	}
	return rumors
}

// RumorsAbout returns cloned rumors that reference a target ID.
func (w *World) RumorsAbout(targetID string) []Rumor {
	return w.RumorsMatching(RumorFilter{TargetID: targetID})
}

// PerceptionsByObserver returns cloned perceptions held by an observer.
func (w *World) PerceptionsByObserver(observerID ActorID) []Perception {
	perceptions := []Perception{}
	for _, key := range w.indices.perceptionsByObserver[observerID] {
		perception, ok := w.Perceptions[key]
		if !ok {
			continue
		}
		perceptions = append(perceptions, clonePerception(perception))
	}
	return perceptions
}

// PerceptionsAbout returns cloned perceptions that reference a target ID.
func (w *World) PerceptionsAbout(targetID string) []Perception {
	perceptions := []Perception{}
	for _, key := range w.indices.perceptionsByTarget[targetID] {
		perception, ok := w.Perceptions[key]
		if !ok {
			continue
		}
		perceptions = append(perceptions, clonePerception(perception))
	}
	return perceptions
}

// MemoriesOf returns a cloned memory record for an actor.
func (w *World) MemoriesOf(actorID ActorID) Memory {
	return w.Memory(actorID)
}

// DepositsByDropper returns cloned deposits dropped by an actor.
func (w *World) DepositsByDropper(actorID ActorID) []DepositedLoot {
	return w.depositsMatching(func(deposit DepositedLoot) bool {
		return deposit.DroppedBy == actorID
	})
}

// DepositsByClaimer returns cloned deposits claimed by an actor.
func (w *World) DepositsByClaimer(actorID ActorID) []DepositedLoot {
	return w.depositsMatching(func(deposit DepositedLoot) bool {
		return deposit.ClaimedBy == actorID
	})
}

// DepositsInDungeon returns cloned deposits for a dungeon.
func (w *World) DepositsInDungeon(dungeonID DungeonID) []DepositedLoot {
	dungeon, ok := w.Dungeons[dungeonID]
	if !ok {
		return nil
	}
	return cloneDeposits(dungeon.Deposits)
}

// UnclaimedDepositsInDungeon returns cloned unclaimed deposits for a dungeon.
func (w *World) UnclaimedDepositsInDungeon(dungeonID DungeonID) []DepositedLoot {
	dungeon, ok := w.Dungeons[dungeonID]
	if !ok {
		return nil
	}
	return filterDeposits(dungeon.Deposits, func(deposit DepositedLoot) bool {
		return !deposit.Claimed()
	})
}

// DepositsInArea returns cloned deposits for an area.
func (w *World) DepositsInArea(areaID AreaID) []DepositedLoot {
	return w.depositsMatching(func(deposit DepositedLoot) bool {
		return deposit.AreaID == areaID
	})
}

// UnclaimedDepositsInArea returns cloned unclaimed deposits for an area.
func (w *World) UnclaimedDepositsInArea(areaID AreaID) []DepositedLoot {
	return w.depositsMatching(func(deposit DepositedLoot) bool {
		return deposit.AreaID == areaID && !deposit.Claimed()
	})
}

// DepositsForItem returns cloned deposits for a crafted item ID.
func (w *World) DepositsForItem(itemID ItemID) []DepositedLoot {
	return w.depositsMatching(func(deposit DepositedLoot) bool {
		return deposit.Item.ID == itemID
	})
}

func (w *World) depositsMatching(match func(DepositedLoot) bool) []DepositedLoot {
	deposits := []DepositedLoot{}
	for _, dungeon := range w.Dungeons {
		deposits = append(deposits, filterDeposits(dungeon.Deposits, match)...)
	}
	return deposits
}

func filterDeposits(deposits []DepositedLoot, match func(DepositedLoot) bool) []DepositedLoot {
	filtered := []DepositedLoot{}
	for _, deposit := range deposits {
		if !match(deposit) {
			continue
		}
		filtered = append(filtered, cloneDeposit(deposit))
	}
	return filtered
}

func eventReferences(event Event, targetID string) bool {
	return event.SubjectID == targetID ||
		event.ObjectID == targetID ||
		event.Subject.ID == targetID ||
		event.Object.ID == targetID
}

func rumorReferences(rumor Rumor, targetID string) bool {
	return rumor.SubjectID == targetID ||
		rumor.ObjectID == targetID ||
		rumor.Subject.ID == targetID ||
		rumor.Object.ID == targetID ||
		string(rumor.SourceID) == targetID
}

func (w *World) eventsFromIndices(indices []int, filter EventFilter) []Event {
	events := []Event{}
	for _, index := range indices {
		if index < 0 || index >= len(w.Events) {
			continue
		}
		event := w.Events[index]
		if filter.Type != "" && event.Type != filter.Type {
			continue
		}
		if filter.TargetID != "" && !eventReferences(event, filter.TargetID) {
			continue
		}
		events = append(events, cloneEvent(event))
	}
	return events
}

func (w *World) rumorsFromIndices(indices []int, filter RumorFilter) []Rumor {
	rumors := []Rumor{}
	for _, index := range indices {
		if index < 0 || index >= len(w.Rumors) {
			continue
		}
		rumor := w.Rumors[index]
		if filter.SourceID != "" && rumor.SourceID != filter.SourceID {
			continue
		}
		if filter.TargetID != "" && !rumorReferences(rumor, filter.TargetID) {
			continue
		}
		rumors = append(rumors, cloneRumor(rumor))
	}
	return rumors
}
