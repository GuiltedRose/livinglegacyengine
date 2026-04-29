package engine

type worldIndices struct {
	eventsByTarget        map[string][]int
	rumorsByTarget        map[string][]int
	rumorsBySource        map[ActorID][]int
	perceptionsByObserver map[ActorID][]PerceptionKey
	perceptionsByTarget   map[string][]PerceptionKey
}

func newWorldIndices() worldIndices {
	return worldIndices{
		eventsByTarget:        map[string][]int{},
		rumorsByTarget:        map[string][]int{},
		rumorsBySource:        map[ActorID][]int{},
		perceptionsByObserver: map[ActorID][]PerceptionKey{},
		perceptionsByTarget:   map[string][]PerceptionKey{},
	}
}

func (w *World) rebuildIndices() {
	w.indices = newWorldIndices()
	for index, event := range w.Events {
		w.indexEvent(index, event)
	}
	for index, rumor := range w.Rumors {
		w.indexRumor(index, rumor)
	}
	for key, perception := range w.Perceptions {
		w.indexPerception(key, perception)
	}
}

func (w *World) indexEvent(index int, event Event) {
	for _, targetID := range eventTargetIDs(event) {
		w.indices.eventsByTarget[targetID] = appendUniqueInt(w.indices.eventsByTarget[targetID], index)
	}
}

func (w *World) indexRumor(index int, rumor Rumor) {
	w.indices.rumorsBySource[rumor.SourceID] = appendUniqueInt(w.indices.rumorsBySource[rumor.SourceID], index)
	for _, targetID := range rumorTargetIDs(rumor) {
		w.indices.rumorsByTarget[targetID] = appendUniqueInt(w.indices.rumorsByTarget[targetID], index)
	}
}

func (w *World) indexPerception(key PerceptionKey, perception Perception) {
	w.indices.perceptionsByObserver[perception.ObserverID] = appendUniquePerceptionKey(w.indices.perceptionsByObserver[perception.ObserverID], key)
	for _, targetID := range perceptionTargetIDs(perception) {
		w.indices.perceptionsByTarget[targetID] = appendUniquePerceptionKey(w.indices.perceptionsByTarget[targetID], key)
	}
}

func eventTargetIDs(event Event) []string {
	return uniqueStrings(event.SubjectID, event.ObjectID, event.Subject.ID, event.Object.ID)
}

func rumorTargetIDs(rumor Rumor) []string {
	return uniqueStrings(rumor.SubjectID, rumor.ObjectID, rumor.Subject.ID, rumor.Object.ID, string(rumor.SourceID))
}

func perceptionTargetIDs(perception Perception) []string {
	return uniqueStrings(perception.TargetID, perception.Target.ID)
}

func uniqueStrings(values ...string) []string {
	seen := map[string]bool{}
	unique := []string{}
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		unique = append(unique, value)
	}
	return unique
}

func appendUniqueInt(values []int, next int) []int {
	for _, value := range values {
		if value == next {
			return values
		}
	}
	return append(values, next)
}

func appendUniquePerceptionKey(values []PerceptionKey, next PerceptionKey) []PerceptionKey {
	for _, value := range values {
		if value == next {
			return values
		}
	}
	return append(values, next)
}
