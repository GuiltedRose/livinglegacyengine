package engine

import (
	"fmt"
	"hash/fnv"
	"regexp"
	"strings"
)

// LoreFact is host-game knowledge that can be woven into generated text.
type LoreFact struct {
	ID         string            `json:"id,omitempty"`
	Text       string            `json:"text"`
	Subject    TargetRef         `json:"subject,omitempty"`
	Object     TargetRef         `json:"object,omitempty"`
	Weight     int               `json:"weight,omitempty"`
	Tags       []string          `json:"tags,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

// GeneratedText is a piece of generated presentation text and its provenance.
type GeneratedText struct {
	Text       string            `json:"text"`
	Sources    []TargetRef       `json:"sources,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

// TextGenerationContext describes the facts available to host lore hooks.
type TextGenerationContext struct {
	SourceID   ActorID
	ListenerID ActorID
	TargetID   string
	Events     []Event
	Rumors     []Rumor
	Ties       []Tie
	History    []WorldHistoryEntry
}

// HistoryRumorTextContext is passed to the history rumor text rule.
type HistoryRumorTextContext struct {
	SourceID ActorID
	Event    Event
	History  []Event
	Lore     []LoreFact
}

// EventRumorOptions controls rumor creation from recorded runtime events.
type EventRumorOptions struct {
	SourceID ActorID
	TargetID string
	Limit    int
	Truth    float64
	Impact   int
	Lore     []LoreFact
}

// HistoryRumorOptions is kept as an alias for older integrations.
type HistoryRumorOptions = EventRumorOptions

var generatedRumorIDPattern = regexp.MustCompile(`[^a-zA-Z0-9_-]+`)

// RecordEventWithGeneratedRumor records an event and immediately creates one
// rumor from that event using the configured lore and text hooks.
func (w *World) RecordEventWithGeneratedRumor(event Event, options EventRumorOptions) (Event, Rumor, error) {
	recorded, err := w.RecordEvent(event)
	if err != nil {
		return Event{}, Rumor{}, err
	}
	if options.TargetID != "" && !eventReferences(recorded, options.TargetID) {
		return recorded, Rumor{}, nil
	}
	truth, impact, err := historyRumorDefaults(options)
	if err != nil {
		return Event{}, Rumor{}, err
	}

	history := w.EventsMatching(EventFilter{TargetID: options.TargetID})
	lore := w.historyRumorLore(options, history)
	rumor, err := w.createHistoryRumor(len(w.Events)-1, recorded, history, lore, options.SourceID, truth, impact)
	if err != nil {
		return Event{}, Rumor{}, err
	}
	return recorded, rumor, nil
}

// GenerateRumorsFromEvents creates fresh rumors from existing runtime events.
func (w *World) GenerateRumorsFromEvents(options EventRumorOptions) ([]Rumor, error) {
	truth, impact, err := historyRumorDefaults(options)
	if err != nil {
		return nil, err
	}

	history := []Event{}
	for _, event := range w.Events {
		if options.TargetID != "" && !eventReferences(event, options.TargetID) {
			continue
		}
		history = append(history, cloneEvent(event))
	}

	lore := w.historyRumorLore(options, history)

	created := []Rumor{}
	for eventIndex, event := range w.Events {
		if options.TargetID != "" && !eventReferences(event, options.TargetID) {
			continue
		}
		rumor, err := w.createHistoryRumor(eventIndex, event, history, lore, options.SourceID, truth, impact)
		if err != nil {
			return nil, err
		}
		if rumor.ID == "" {
			continue
		}
		created = append(created, rumor)
		if options.Limit > 0 && len(created) >= options.Limit {
			break
		}
	}
	return created, nil
}

// GenerateRumorsFromHistory is kept as an alias for older integrations.
func (w *World) GenerateRumorsFromHistory(options EventRumorOptions) ([]Rumor, error) {
	return w.GenerateRumorsFromEvents(options)
}

func historyRumorDefaults(options EventRumorOptions) (float64, int, error) {
	if options.SourceID == "" {
		return 0, 0, fmt.Errorf("source actor id is required")
	}
	truth := options.Truth
	if truth == 0 {
		truth = 1
	}
	if truth < 0 || truth > 1 {
		return 0, 0, fmt.Errorf("rumor truth must be between 0 and 1")
	}
	impact := options.Impact
	if impact == 0 {
		impact = 1
	}
	return truth, impact, nil
}

func (w *World) historyRumorLore(options EventRumorOptions, history []Event) []LoreFact {
	lore := cloneLoreFacts(options.Lore)
	lore = append(lore, w.rules.LoreFacts(TextGenerationContext{
		SourceID: options.SourceID,
		TargetID: options.TargetID,
		Events:   cloneEvents(history),
		Rumors:   w.RumorsMatching(RumorFilter{TargetID: options.TargetID}),
		Ties:     w.TiesForTarget(options.TargetID),
		History:  cloneWorldHistory(w.WorldHistory),
	})...)
	return lore
}

func (w *World) createHistoryRumor(eventIndex int, event Event, history []Event, lore []LoreFact, sourceID ActorID, truth float64, impact int) (Rumor, error) {
	id := generatedHistoryRumorID(sourceID, eventIndex, event)
	if w.hasRumor(id) {
		return Rumor{}, nil
	}
	text := w.rules.EventRumorText(HistoryRumorTextContext{
		SourceID: sourceID,
		Event:    cloneEvent(event),
		History:  cloneEvents(history),
		Lore:     cloneLoreFacts(lore),
	})
	description := strings.TrimSpace(text.Text)
	if description == "" {
		description = event.Description
	}
	rumor, err := NewRumor(id, sourceID, description, truth, impact, w.now())
	if err != nil {
		return Rumor{}, err
	}
	rumor.SubjectID = event.SubjectID
	rumor.ObjectID = event.ObjectID
	rumor.Subject = cloneTargetRef(event.Subject)
	rumor.Object = cloneTargetRef(event.Object)
	rumor.Attributes = cloneStringMap(text.Attributes)
	if rumor.Attributes == nil {
		rumor.Attributes = map[string]string{}
	}
	rumor.Attributes["generated_from"] = "history"
	rumor.Attributes["history_event_type"] = string(event.Type)
	for i, source := range text.Sources {
		if source.ID == "" {
			continue
		}
		rumor.Attributes[fmt.Sprintf("source_%d", i+1)] = source.ID
	}
	if err := w.AddRumor(rumor); err != nil {
		return Rumor{}, err
	}
	return cloneRumor(rumor), nil
}

// GenerateRemark returns a generated line from known rumors, events, and lore.
func (w *World) GenerateRemark(context TextGenerationContext) GeneratedText {
	events := cloneEvents(context.Events)
	if len(events) == 0 {
		events = w.EventsMatching(EventFilter{TargetID: context.TargetID})
	}
	rumors := cloneRumors(context.Rumors)
	if len(rumors) == 0 {
		rumors = w.RumorsMatching(RumorFilter{TargetID: context.TargetID})
	}
	lore := w.rules.LoreFacts(TextGenerationContext{
		SourceID:   context.SourceID,
		ListenerID: context.ListenerID,
		TargetID:   context.TargetID,
		Events:     events,
		Rumors:     rumors,
		Ties:       w.TiesForTarget(context.TargetID),
		History:    cloneWorldHistory(w.WorldHistory),
	})
	if len(rumors) > 0 {
		rumor := rumors[stableIndex(context.TargetID+string(context.SourceID), len(rumors))]
		return GeneratedText{
			Text:       rumor.Description,
			Sources:    []TargetRef{rumor.Subject, rumor.Object},
			Attributes: map[string]string{"kind": "rumor"},
		}
	}
	ties := cloneTies(context.Ties)
	if len(ties) == 0 && context.TargetID != "" {
		ties = w.TiesForTarget(context.TargetID)
	}
	if len(ties) > 0 {
		tie := ties[stableIndex(context.TargetID+string(context.SourceID), len(ties))]
		return GeneratedText{
			Text:       tie.Description,
			Sources:    []TargetRef{tie.Subject, tie.Object},
			Attributes: map[string]string{"kind": "world_history_tie", "tie_kind": string(tie.Kind)},
		}
	}
	if len(events) > 0 {
		event := events[stableIndex(context.TargetID+string(context.SourceID), len(events))]
		return w.rules.EventRumorText(HistoryRumorTextContext{
			SourceID: context.SourceID,
			Event:    event,
			History:  events,
			Lore:     lore,
		})
	}
	if fact, ok := bestLoreFact(lore, context.TargetID); ok {
		return GeneratedText{
			Text:       fact.Text,
			Sources:    []TargetRef{fact.Subject, fact.Object},
			Attributes: map[string]string{"kind": "lore"},
		}
	}
	return GeneratedText{
		Text:       "No one has heard anything worth trusting yet.",
		Attributes: map[string]string{"kind": "empty"},
	}
}

// DefaultHistoryRumorText is the baseline deterministic text generator.
func DefaultHistoryRumorText(context HistoryRumorTextContext) GeneratedText {
	event := context.Event
	subject := targetName(event.Subject, event.SubjectID, "someone")
	object := targetName(event.Object, event.ObjectID, "something")
	loreText, loreSource := matchingLore(context.Lore, event)

	var variants []string
	switch event.Type {
	case EventItemCrafted:
		variants = []string{
			fmt.Sprintf("They say %s first entered the record in %s's hands.", object, subject),
			fmt.Sprintf("The old talk says %s was made famous by %s.", object, subject),
			fmt.Sprintf("Ask around and %s's name keeps returning with %s.", subject, object),
		}
	case EventCharacterDied:
		variants = []string{
			fmt.Sprintf("People still lower their voices about %s's death.", subject),
			fmt.Sprintf("The story goes that %s fell, and the place remembered.", subject),
			fmt.Sprintf("There is old fear around the day %s did not come back whole.", subject),
		}
	case EventLootDeposited:
		variants = []string{
			fmt.Sprintf("Word is %s left crafted work sealed inside %s.", subject, object),
			fmt.Sprintf("The tale says %s fed %s with lost craft.", subject, object),
			fmt.Sprintf("Travelers claim %s became richer after %s fell.", object, subject),
		}
	case EventDungeonSpawned:
		variants = []string{
			fmt.Sprintf("They say %s was waiting before anyone dared name it.", object),
			fmt.Sprintf("Old maps started marking %s before the first clear.", object),
			fmt.Sprintf("The earliest tale puts %s in the world already hungry.", object),
		}
	case EventDungeonLooted:
		variants = []string{
			fmt.Sprintf("The talk is that %s came out of %s changed.", subject, object),
			fmt.Sprintf("They say %s broke into %s and legacy followed.", subject, object),
			fmt.Sprintf("If the story is true, %s took more than loot from %s.", subject, object),
		}
	case EventDungeonLocked:
		variants = []string{
			fmt.Sprintf("For now, %s answers only to %s.", object, subject),
			fmt.Sprintf("They say %s shut its doors behind %s.", object, subject),
			fmt.Sprintf("The warning on %s is simple: %s has the claim.", object, subject),
		}
	case EventDungeonUnlocked:
		variants = []string{
			fmt.Sprintf("The lock on %s broke after %s died again.", object, subject),
			fmt.Sprintf("People say %s reopened when %s's claim failed.", object, subject),
			fmt.Sprintf("There is fresh talk that %s no longer belongs to %s.", object, subject),
		}
	case EventDungeonDecayed:
		variants = []string{
			fmt.Sprintf("The last word on %s is decay.", object),
			fmt.Sprintf("They say %s lost its old hunger and collapsed into memory.", object),
			fmt.Sprintf("No one brings back clear directions to %s anymore.", object),
		}
	case EventRespawned:
		variants = []string{
			fmt.Sprintf("They say %s returned again, which is never a small thing.", subject),
			fmt.Sprintf("The rumor is that death failed to keep %s.", subject),
			fmt.Sprintf("Some swear %s came back carrying less, but meaning more.", subject),
		}
	default:
		variants = []string{event.Description}
	}

	text := variants[stableIndex(event.Description+string(event.Type), len(variants))]
	attributes := map[string]string{"kind": "history_rumor"}
	sources := []TargetRef{event.Subject, event.Object}
	if loreText != "" {
		text = fmt.Sprintf("%s %s", text, loreText)
		attributes["lore"] = "true"
		if !loreSource.Empty() {
			sources = append(sources, loreSource)
		}
	}
	return GeneratedText{
		Text:       text,
		Sources:    nonEmptyTargets(sources),
		Attributes: attributes,
	}
}

func generatedHistoryRumorID(sourceID ActorID, index int, event Event) RumorID {
	seed := fmt.Sprintf("%s-%03d-%s-%s-%s", sourceID, index+1, event.Type, event.SubjectID, event.ObjectID)
	seed = strings.ToLower(generatedRumorIDPattern.ReplaceAllString(seed, "-"))
	seed = strings.Trim(seed, "-")
	return RumorID("history-rumor-" + seed)
}

func (w *World) hasRumor(id RumorID) bool {
	for _, rumor := range w.Rumors {
		if rumor.ID == id {
			return true
		}
	}
	return false
}

func targetName(target TargetRef, fallbackID string, fallback string) string {
	if target.Name != "" {
		return target.Name
	}
	if fallbackID != "" {
		return fallbackID
	}
	if target.ID != "" {
		return target.ID
	}
	return fallback
}

func matchingLore(lore []LoreFact, event Event) (string, TargetRef) {
	for _, targetID := range []string{event.SubjectID, event.Subject.ID, event.ObjectID, event.Object.ID} {
		if targetID == "" {
			continue
		}
		if fact, ok := bestLoreFact(lore, targetID); ok {
			if fact.Subject.ID == targetID {
				return fact.Text, fact.Subject
			}
			return fact.Text, fact.Object
		}
	}
	return "", TargetRef{}
}

func bestLoreFact(lore []LoreFact, targetID string) (LoreFact, bool) {
	var best LoreFact
	found := false
	for _, fact := range lore {
		if strings.TrimSpace(fact.Text) == "" {
			continue
		}
		if targetID != "" && fact.Subject.ID != targetID && fact.Object.ID != targetID {
			continue
		}
		if !found || fact.Weight > best.Weight {
			best = fact
			found = true
		}
	}
	return cloneLoreFact(best), found
}

func stableIndex(seed string, size int) int {
	if size <= 0 {
		return 0
	}
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(seed))
	return int(hash.Sum32() % uint32(size))
}

func cloneLoreFacts(facts []LoreFact) []LoreFact {
	if facts == nil {
		return nil
	}
	cloned := make([]LoreFact, len(facts))
	for i, fact := range facts {
		cloned[i] = cloneLoreFact(fact)
	}
	return cloned
}

func cloneLoreFact(fact LoreFact) LoreFact {
	fact.Subject = cloneTargetRef(fact.Subject)
	fact.Object = cloneTargetRef(fact.Object)
	fact.Tags = append([]string(nil), fact.Tags...)
	fact.Attributes = cloneStringMap(fact.Attributes)
	return fact
}

func nonEmptyTargets(targets []TargetRef) []TargetRef {
	filtered := []TargetRef{}
	for _, target := range targets {
		if target.Empty() {
			continue
		}
		filtered = append(filtered, cloneTargetRef(target))
	}
	return filtered
}
