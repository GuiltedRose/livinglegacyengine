package engine

import "fmt"

// PerceptionKey indexes one observer's view of one target.
type PerceptionKey struct {
	ObserverID ActorID `json:"observer_id"`
	TargetID   string  `json:"target_id"`
}

// Perception stores one actor's opinion of a target.
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

// PerceptionDelta describes score and attribute changes to apply to a perception.
type PerceptionDelta struct {
	Trust      int
	Fear       int
	Respect    int
	Affinity   int
	Notoriety  int
	Confidence int
	Attributes map[string]string
}

// NewPerception creates an empty perception record for an observer and target.
func NewPerception(observerID ActorID, targetID string) Perception {
	return Perception{
		ObserverID: observerID,
		TargetID:   targetID,
		Target:     NewTargetRef(targetID, TargetCustom, ""),
		Attributes: map[string]string{},
	}
}

// Perception returns a cloned perception or an empty default record.
func (w *World) Perception(observerID ActorID, targetID string) Perception {
	key := perceptionKey(observerID, targetID)
	perception, ok := w.Perceptions[key]
	if !ok {
		return NewPerception(observerID, targetID)
	}
	return clonePerception(perception)
}

// AdjustPerception applies a delta to a custom target ID.
func (w *World) AdjustPerception(observerID ActorID, targetID string, delta PerceptionDelta) (Perception, error) {
	return w.AdjustPerceptionForTarget(observerID, NewTargetRef(targetID, TargetCustom, ""), delta)
}

// AdjustPerceptionForTarget applies a delta to a structured target reference.
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

// ApplyRumorToPerception converts a rumor into a perception delta with rules.
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

// ApplyEventToPerception converts an event into a perception delta with rules.
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
