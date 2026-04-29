package engine

import "fmt"

type PerceptionKey struct {
	ObserverID ActorID `json:"observer_id"`
	TargetID   string  `json:"target_id"`
}

type Perception struct {
	ObserverID ActorID           `json:"observer_id"`
	TargetID   string            `json:"target_id"`
	Target     TargetRef         `json:"target,omitempty"`
	Trust      int               `json:"trust"`
	Fear       int               `json:"fear"`
	Respect    int               `json:"respect"`
	Affinity   int               `json:"affinity"`
	Notoriety  int               `json:"notoriety"`
	Confidence int               `json:"confidence"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

type PerceptionDelta struct {
	Trust      int
	Fear       int
	Respect    int
	Affinity   int
	Notoriety  int
	Confidence int
	Attributes map[string]string
}

func NewPerception(observerID ActorID, targetID string) Perception {
	return Perception{
		ObserverID: observerID,
		TargetID:   targetID,
		Target:     NewTargetRef(targetID, TargetCustom, ""),
		Attributes: map[string]string{},
	}
}

func (w *World) Perception(observerID ActorID, targetID string) Perception {
	key := perceptionKey(observerID, targetID)
	perception, ok := w.Perceptions[key]
	if !ok {
		return NewPerception(observerID, targetID)
	}
	return clonePerception(perception)
}

func (w *World) AdjustPerception(observerID ActorID, targetID string, delta PerceptionDelta) (Perception, error) {
	return w.AdjustPerceptionForTarget(observerID, NewTargetRef(targetID, TargetCustom, ""), delta)
}

func (w *World) AdjustPerceptionForTarget(observerID ActorID, target TargetRef, delta PerceptionDelta) (Perception, error) {
	if observerID == "" {
		return Perception{}, fmt.Errorf("observer actor id is required")
	}
	if target.ID == "" {
		return Perception{}, fmt.Errorf("target id is required")
	}

	target = target.normalized(TargetCustom)
	perception := w.Perception(observerID, target.ID)
	perception.Target = target
	perception.Trust = clampScore(perception.Trust + delta.Trust)
	perception.Fear = clampScore(perception.Fear + delta.Fear)
	perception.Respect = clampScore(perception.Respect + delta.Respect)
	perception.Affinity = clampScore(perception.Affinity + delta.Affinity)
	perception.Notoriety = clampScore(perception.Notoriety + delta.Notoriety)
	perception.Confidence = clampScore(perception.Confidence + delta.Confidence)
	if perception.Attributes == nil {
		perception.Attributes = map[string]string{}
	}
	for key, value := range delta.Attributes {
		perception.Attributes[key] = value
	}

	key := perceptionKey(observerID, target.ID)
	w.Perceptions[key] = perception
	w.rebuildIndices()
	return clonePerception(perception), nil
}

func (w *World) ApplyRumorToPerception(observerID ActorID, rumor Rumor) (Perception, error) {
	delta := w.rules.RumorPerception(rumor)
	target := rumor.Subject
	if target.Empty() {
		target = rumor.Object
	}
	if target.Empty() && rumor.SubjectID != "" {
		target = NewTargetRef(rumor.SubjectID, TargetCustom, "")
	}
	if target.Empty() && rumor.ObjectID != "" {
		target = NewTargetRef(rumor.ObjectID, TargetCustom, "")
	}
	if target.Empty() {
		target = NewTargetRef(string(rumor.SourceID), TargetActor, "")
	}
	return w.AdjustPerceptionForTarget(observerID, target, delta)
}

func (w *World) ApplyEventToPerception(observerID ActorID, event Event) (Perception, error) {
	delta := w.rules.EventPerception(event)
	target := event.Subject
	if target.Empty() {
		target = event.Object
	}
	if target.Empty() && event.SubjectID != "" {
		target = NewTargetRef(event.SubjectID, TargetCustom, "")
	}
	if target.Empty() && event.ObjectID != "" {
		target = NewTargetRef(event.ObjectID, TargetCustom, "")
	}
	if target.Empty() {
		return Perception{}, fmt.Errorf("event has no perception target")
	}
	return w.AdjustPerceptionForTarget(observerID, target, delta)
}

func perceptionKey(observerID ActorID, targetID string) PerceptionKey {
	return PerceptionKey{
		ObserverID: observerID,
		TargetID:   targetID,
	}
}

func clampScore(value int) int {
	if value < -100 {
		return -100
	}
	if value > 100 {
		return 100
	}
	return value
}
