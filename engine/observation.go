package engine

import "fmt"

func (w *World) RecordEvent(event Event) (Event, error) {
	if event.Type == "" {
		return Event{}, fmt.Errorf("event type is required")
	}
	if event.Description == "" {
		return Event{}, fmt.Errorf("event description is required")
	}
	event = w.normalizeEvent(event)
	w.Events = append(w.Events, event)
	w.indexEvent(len(w.Events)-1, event)
	return cloneEvent(event), nil
}

func (w *World) ObserveEvent(observerID ActorID, event Event) (Perception, error) {
	if observerID == "" {
		return Perception{}, fmt.Errorf("observer actor id is required")
	}
	event = w.normalizeEvent(event)
	perception, err := w.ApplyEventToPerception(observerID, event)
	if err != nil {
		return Perception{}, err
	}
	w.TeachEvent(observerID, event)
	return perception, nil
}

func (w *World) RecordObservedEvent(event Event, observers ...ActorID) (Event, []Perception, error) {
	recorded, err := w.RecordEvent(event)
	if err != nil {
		return Event{}, nil, err
	}

	perceptions := make([]Perception, 0, len(observers))
	for _, observerID := range observers {
		perception, err := w.ObserveEvent(observerID, recorded)
		if err != nil {
			return Event{}, nil, err
		}
		perceptions = append(perceptions, perception)
	}
	return recorded, perceptions, nil
}
