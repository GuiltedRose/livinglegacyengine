package engine

type EventFilter struct {
	Type     EventType
	TargetID string
}

type RumorFilter struct {
	SourceID ActorID
	TargetID string
}

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

func (w *World) EventsForTarget(targetID string) []Event {
	return w.EventsMatching(EventFilter{TargetID: targetID})
}

func (w *World) EventsForActor(actorID ActorID) []Event {
	return w.EventsForTarget(string(actorID))
}

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

func (w *World) RumorsAbout(targetID string) []Rumor {
	return w.RumorsMatching(RumorFilter{TargetID: targetID})
}

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

func (w *World) MemoriesOf(actorID ActorID) Memory {
	return w.Memory(actorID)
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
