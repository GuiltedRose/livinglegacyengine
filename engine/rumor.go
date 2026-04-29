package engine

import (
	"fmt"
	"time"
)

// RumorID is a stable identifier for a rumor.
type RumorID string

// Rumor is an uncertain claim known by one or more actors.
type Rumor struct {
	ID          RumorID           `json:"id"`
	SourceID    ActorID           `json:"source_id"`
	SubjectID   string            `json:"subject_id,omitempty"`
	ObjectID    string            `json:"object_id,omitempty"`
	Subject     TargetRef         `json:"subject,omitempty"`
	Object      TargetRef         `json:"object,omitempty"`
	Description string            `json:"description"`
	Impact      int               `json:"impact"`
	Truth       float64           `json:"truth"`
	Spread      int               `json:"spread"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Attributes  map[string]string `json:"attributes,omitempty"`
}

// NewRumor creates a rumor with validated truth and required identity fields.
func NewRumor(id RumorID, sourceID ActorID, description string, truth float64, impact int, now time.Time) (Rumor, error) {
	if id == "" {
		return Rumor{}, fmt.Errorf("rumor id is required")
	}
	if sourceID == "" {
		return Rumor{}, fmt.Errorf("rumor source id is required")
	}
	if description == "" {
		return Rumor{}, fmt.Errorf("rumor description is required")
	}
	if truth < 0 || truth > 1 {
		return Rumor{}, fmt.Errorf("rumor truth must be between 0 and 1")
	}
	return Rumor{
		ID:          id,
		SourceID:    sourceID,
		Description: description,
		Impact:      impact,
		Truth:       truth,
		CreatedAt:   now,
		UpdatedAt:   now,
		Attributes:  map[string]string{},
	}, nil
}

// AddRumor stores a rumor, indexes it, and teaches it to its source actor.
func (w *World) AddRumor(rumor Rumor) error {
	if rumor.ID == "" {
		return fmt.Errorf("rumor id is required")
	}
	if rumor.SourceID == "" {
		return fmt.Errorf("rumor source id is required")
	}
	if rumor.Description == "" {
		return fmt.Errorf("rumor description is required")
	}
	if rumor.Truth < 0 || rumor.Truth > 1 {
		return fmt.Errorf("rumor truth must be between 0 and 1")
	}
	if rumor.CreatedAt.IsZero() {
		rumor.CreatedAt = w.now()
	}
	if rumor.UpdatedAt.IsZero() {
		rumor.UpdatedAt = rumor.CreatedAt
	}
	rumor.Subject = rumor.Subject.normalized(TargetCustom)
	rumor.Object = rumor.Object.normalized(TargetCustom)
	rumor.Attributes = cloneStringMap(rumor.Attributes)
	if rumor.Attributes == nil {
		rumor.Attributes = map[string]string{}
	}
	w.Rumors = append(w.Rumors, rumor)
	w.indexRumor(len(w.Rumors)-1, rumor)
	w.TeachRumor(rumor.SourceID, rumor)
	return nil
}

// SpreadRumor teaches a rumor to a recipient and applies perception effects.
func (w *World) SpreadRumor(id RumorID, recipient ActorID, distortion float64) (Rumor, error) {
	if recipient == "" {
		return Rumor{}, fmt.Errorf("recipient actor id is required")
	}
	if distortion < 0 {
		return Rumor{}, fmt.Errorf("distortion cannot be negative")
	}
	for index, rumor := range w.Rumors {
		if rumor.ID != id {
			continue
		}
		rumor.Spread++
		rumor.Truth = clamp01(rumor.Truth - distortion)
		rumor.UpdatedAt = w.now()
		w.Rumors[index] = rumor
		w.TeachRumor(recipient, rumor)
		if _, err := w.ApplyRumorToPerception(recipient, rumor); err != nil {
			return Rumor{}, err
		}
		return cloneRumor(rumor), nil
	}
	return Rumor{}, fmt.Errorf("rumor %q does not exist", id)
}

func clamp01(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}
