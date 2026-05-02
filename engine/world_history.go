package engine

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// TieID is a stable identifier for a generated relationship or allegiance.
type TieID string

// TieKind classifies a generated tie between two world targets.
type TieKind string

const (
	TieMembership TieKind = "membership"
	TieAlliance   TieKind = "alliance"
	TieRivalry    TieKind = "rivalry"
	TieDebt       TieKind = "debt"
	TieOath       TieKind = "oath"
	TieMentorship TieKind = "mentorship"
)

// Tie is an engine-owned relationship between actors, factions, or other targets.
type Tie struct {
	ID          TieID             `json:"id"`
	Subject     TargetRef         `json:"subject"`
	Object      TargetRef         `json:"object"`
	Kind        TieKind           `json:"kind"`
	Strength    int               `json:"strength"`
	Description string            `json:"description"`
	StartedAt   time.Time         `json:"started_at"`
	Attributes  map[string]string `json:"attributes,omitempty"`
}

// WorldHistoryEntry records generated setting history, separate from runtime events.
type WorldHistoryEntry struct {
	ID         string            `json:"id"`
	At         time.Time         `json:"at"`
	Summary    string            `json:"summary"`
	Subject    TargetRef         `json:"subject,omitempty"`
	Object     TargetRef         `json:"object,omitempty"`
	TieID      TieID             `json:"tie_id,omitempty"`
	Tags       []string          `json:"tags,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

// WorldGenerationOptions controls deterministic NPC, faction, and tie generation.
type WorldGenerationOptions struct {
	Seed         int64
	NPCCount     int
	FactionCount int
	TieCount     int
	NPCNames     []string
	FactionNames []string
	Era          string
}

// WorldGenerationResult reports what GenerateWorldHistory added to the world.
type WorldGenerationResult struct {
	Actors  []Actor
	Ties    []Tie
	History []WorldHistoryEntry
}

var defaultNPCNames = []string{"Rin", "Sable", "Oren", "Mira", "Tovan", "Ilyra", "Cass", "Nadim"}
var defaultFactionNames = []string{"Lantern Compact", "Ash Stair Keepers", "Glasswright Circle", "Morrow League"}
var generatedTieKinds = []TieKind{TieAlliance, TieRivalry, TieDebt, TieOath, TieMentorship}

// NewTie validates and creates a relationship between two targets.
func NewTie(id TieID, subject TargetRef, object TargetRef, kind TieKind, strength int, description string, now time.Time) (Tie, error) {
	if id == "" {
		return Tie{}, fmt.Errorf("tie id is required")
	}
	if subject.Empty() {
		return Tie{}, fmt.Errorf("tie subject is required")
	}
	if object.Empty() {
		return Tie{}, fmt.Errorf("tie object is required")
	}
	if subject.ID == object.ID {
		return Tie{}, fmt.Errorf("tie subject and object must differ")
	}
	if kind == "" {
		kind = TieAlliance
	}
	if strength == 0 {
		strength = 1
	}
	if description == "" {
		description = describeTie(subject, object, kind, strength)
	}
	return Tie{
		ID:          id,
		Subject:     subject.normalized(TargetCustom),
		Object:      object.normalized(TargetCustom),
		Kind:        kind,
		Strength:    strength,
		Description: description,
		StartedAt:   now,
		Attributes:  map[string]string{},
	}, nil
}

// AddTie stores a relationship and records generated setting history for it.
func (w *World) AddTie(tie Tie) error {
	if tie.ID == "" {
		return fmt.Errorf("tie id is required")
	}
	if tie.Subject.Empty() {
		return fmt.Errorf("tie subject is required")
	}
	if tie.Object.Empty() {
		return fmt.Errorf("tie object is required")
	}
	if tie.Subject.ID == tie.Object.ID {
		return fmt.Errorf("tie subject and object must differ")
	}
	if tie.Kind == "" {
		tie.Kind = TieAlliance
	}
	if tie.Strength == 0 {
		tie.Strength = 1
	}
	if tie.StartedAt.IsZero() {
		tie.StartedAt = w.now()
	}
	tie.Subject = tie.Subject.normalized(TargetCustom)
	tie.Object = tie.Object.normalized(TargetCustom)
	tie.Attributes = cloneStringMap(tie.Attributes)
	if tie.Attributes == nil {
		tie.Attributes = map[string]string{}
	}
	if tie.Description == "" {
		tie.Description = describeTie(tie.Subject, tie.Object, tie.Kind, tie.Strength)
	}
	w.Ties = append(w.Ties, tie)
	w.WorldHistory = append(w.WorldHistory, WorldHistoryEntry{
		ID:      fmt.Sprintf("history-%s", tie.ID),
		At:      tie.StartedAt,
		Summary: tie.Description,
		Subject: cloneTargetRef(tie.Subject),
		Object:  cloneTargetRef(tie.Object),
		TieID:   tie.ID,
		Tags:    []string{"generated", "tie", string(tie.Kind)},
	})
	return nil
}

// GenerateWorldHistory creates engine-owned NPCs, factions, and ties.
func (w *World) GenerateWorldHistory(options WorldGenerationOptions) (WorldGenerationResult, error) {
	if options.NPCCount == 0 {
		options.NPCCount = 4
	}
	if options.FactionCount == 0 {
		options.FactionCount = 2
	}
	if options.TieCount == 0 {
		options.TieCount = options.NPCCount + options.FactionCount
	}
	if options.NPCCount < 0 || options.FactionCount < 0 || options.TieCount < 0 {
		return WorldGenerationResult{}, fmt.Errorf("world generation counts cannot be negative")
	}
	if options.Seed == 0 {
		options.Seed = w.now().UnixNano()
	}
	rng := rand.New(rand.NewSource(options.Seed))

	result := WorldGenerationResult{}
	factions, err := w.generateActors(options.FactionNames, defaultFactionNames, options.FactionCount, ActorFaction, "faction")
	if err != nil {
		return WorldGenerationResult{}, err
	}
	npcs, err := w.generateActors(options.NPCNames, defaultNPCNames, options.NPCCount, ActorNPC, "npc")
	if err != nil {
		return WorldGenerationResult{}, err
	}
	result.Actors = append(result.Actors, factions...)
	result.Actors = append(result.Actors, npcs...)

	for i, npc := range npcs {
		if len(factions) == 0 || len(result.Ties) >= options.TieCount {
			break
		}
		faction := factions[(i+rng.Intn(len(factions)))%len(factions)]
		tie, err := NewTie(
			TieID(fmt.Sprintf("tie-%s-%s", npc.ID, faction.ID)),
			generatedActorRef(npc),
			generatedActorRef(faction),
			TieMembership,
			1+rng.Intn(5),
			"",
			w.now(),
		)
		if err != nil {
			return WorldGenerationResult{}, err
		}
		tie.Attributes["era"] = options.Era
		if err := w.AddTie(tie); err != nil {
			return WorldGenerationResult{}, err
		}
		result.Ties = append(result.Ties, cloneTie(tie))
		result.History = append(result.History, cloneWorldHistoryEntry(w.WorldHistory[len(w.WorldHistory)-1]))
	}

	targets := append(actorRefs(npcs), actorRefs(factions)...)
	for len(result.Ties) < options.TieCount && len(targets) > 1 {
		first := rng.Intn(len(targets))
		second := rng.Intn(len(targets) - 1)
		if second >= first {
			second++
		}
		subject := targets[first]
		object := targets[second]
		kind := generatedTieKinds[rng.Intn(len(generatedTieKinds))]
		tie, err := NewTie(
			TieID(fmt.Sprintf("tie-%s-%s-%d", subject.ID, object.ID, len(w.Ties)+1)),
			subject,
			object,
			kind,
			1+rng.Intn(5),
			"",
			w.now(),
		)
		if err != nil {
			return WorldGenerationResult{}, err
		}
		tie.Attributes["era"] = options.Era
		if err := w.AddTie(tie); err != nil {
			return WorldGenerationResult{}, err
		}
		result.Ties = append(result.Ties, cloneTie(tie))
		result.History = append(result.History, cloneWorldHistoryEntry(w.WorldHistory[len(w.WorldHistory)-1]))
	}

	return result, nil
}

// TiesForTarget returns cloned ties that reference a target ID.
func (w *World) TiesForTarget(targetID string) []Tie {
	ties := []Tie{}
	for _, tie := range w.Ties {
		if tie.Subject.ID == targetID || tie.Object.ID == targetID {
			ties = append(ties, cloneTie(tie))
		}
	}
	return ties
}

func (w *World) generateActors(names []string, defaults []string, count int, kind ActorKind, prefix string) ([]Actor, error) {
	actors := []Actor{}
	for i := 0; i < count; i++ {
		name := generatedName(names, defaults, i)
		id := ActorID(uniqueActorID(w.Actors, prefix+"-"+slug(name), i))
		actor, err := NewActor(id, name, kind)
		if err != nil {
			return nil, err
		}
		actor.Attributes["generated"] = "true"
		if err := w.AddActor(actor); err != nil {
			return nil, err
		}
		actors = append(actors, actor)
	}
	return actors, nil
}

func generatedName(names []string, defaults []string, index int) string {
	if index < len(names) && names[index] != "" {
		return names[index]
	}
	if len(defaults) == 0 {
		return fmt.Sprintf("Actor %d", index+1)
	}
	return defaults[index%len(defaults)]
}

func uniqueActorID(existing map[ActorID]Actor, base string, index int) string {
	if base == "" || base == "-" {
		base = fmt.Sprintf("actor-%d", index+1)
	}
	id := base
	for suffix := 2; ; suffix++ {
		if _, ok := existing[ActorID(id)]; !ok {
			return id
		}
		id = fmt.Sprintf("%s-%d", base, suffix)
	}
}

func actorRefs(actors []Actor) []TargetRef {
	refs := make([]TargetRef, len(actors))
	for i, actor := range actors {
		refs[i] = generatedActorRef(actor)
	}
	return refs
}

func generatedActorRef(actor Actor) TargetRef {
	switch actor.Kind {
	case ActorFaction:
		return NewTargetRef(string(actor.ID), TargetFaction, actor.Name)
	case ActorNPC, ActorCharacter:
		return NewTargetRef(string(actor.ID), TargetCharacter, actor.Name)
	default:
		return ActorRef(actor)
	}
}

func describeTie(subject TargetRef, object TargetRef, kind TieKind, strength int) string {
	subjectName := targetName(subject, subject.ID, "someone")
	objectName := targetName(object, object.ID, "someone")
	switch kind {
	case TieMembership:
		return fmt.Sprintf("%s is bound to %s by service and reputation.", subjectName, objectName)
	case TieRivalry:
		return fmt.Sprintf("%s and %s have an old rivalry that still shapes local talk.", subjectName, objectName)
	case TieDebt:
		return fmt.Sprintf("%s owes %s a debt strong enough to survive retelling.", subjectName, objectName)
	case TieOath:
		return fmt.Sprintf("%s swore an oath that still pulls them toward %s.", subjectName, objectName)
	case TieMentorship:
		return fmt.Sprintf("%s learned under %s, and neither name travels alone.", subjectName, objectName)
	default:
		if strength >= 4 {
			return fmt.Sprintf("%s and %s are known as close allies.", subjectName, objectName)
		}
		return fmt.Sprintf("%s and %s are tied by a fragile alliance.", subjectName, objectName)
	}
}

func slug(value string) string {
	value = strings.ToLower(generatedRumorIDPattern.ReplaceAllString(value, "-"))
	return strings.Trim(value, "-")
}

func cloneTies(ties []Tie) []Tie {
	if ties == nil {
		return nil
	}
	cloned := make([]Tie, len(ties))
	for i, tie := range ties {
		cloned[i] = cloneTie(tie)
	}
	return cloned
}

func cloneTie(tie Tie) Tie {
	tie.Subject = cloneTargetRef(tie.Subject)
	tie.Object = cloneTargetRef(tie.Object)
	tie.Attributes = cloneStringMap(tie.Attributes)
	return tie
}

func cloneWorldHistory(entries []WorldHistoryEntry) []WorldHistoryEntry {
	if entries == nil {
		return nil
	}
	cloned := make([]WorldHistoryEntry, len(entries))
	for i, entry := range entries {
		cloned[i] = cloneWorldHistoryEntry(entry)
	}
	return cloned
}

func cloneWorldHistoryEntry(entry WorldHistoryEntry) WorldHistoryEntry {
	entry.Subject = cloneTargetRef(entry.Subject)
	entry.Object = cloneTargetRef(entry.Object)
	entry.Tags = append([]string(nil), entry.Tags...)
	entry.Attributes = cloneStringMap(entry.Attributes)
	return entry
}
