package engine

import (
	"fmt"
	"time"
)

type ItemEligibilityRule func(item CraftedItem) bool
type ItemLegacyRule func(item CraftedItem) int
type DungeonNameRule func(character Character, run int) string
type DungeonDepthRule func(character Character, run int) int
type DungeonAttributesRule func(character Character, items []CraftedItem) map[string]string
type DungeonDecayRule func(dungeon LootDungeon, now time.Time) bool
type RumorPerceptionRule func(rumor Rumor) PerceptionDelta
type EventPerceptionRule func(event Event) PerceptionDelta

type Rules struct {
	EligibleForDungeon ItemEligibilityRule
	ItemLegacyValue    ItemLegacyRule
	DungeonName        DungeonNameRule
	DungeonDepth       DungeonDepthRule
	DungeonAttributes  DungeonAttributesRule
	ShouldDecayDungeon DungeonDecayRule
	RumorPerception    RumorPerceptionRule
	EventPerception    EventPerceptionRule
}

func DefaultRules() Rules {
	return Rules{
		EligibleForDungeon: func(item CraftedItem) bool {
			return item.CrafterID != ""
		},
		ItemLegacyValue: func(item CraftedItem) int {
			return item.LegacyValue()
		},
		DungeonName: func(character Character, run int) string {
			return fmt.Sprintf("%s's Fallen Cache %d", character.Name, run)
		},
		DungeonDepth: func(character Character, _ int) int {
			return max(1, character.Level)
		},
		DungeonAttributes: func(Character, []CraftedItem) map[string]string {
			return map[string]string{}
		},
		ShouldDecayDungeon: func(LootDungeon, time.Time) bool {
			return false
		},
		RumorPerception: func(rumor Rumor) PerceptionDelta {
			confidence := int(rumor.Truth * 100)
			return PerceptionDelta{
				Notoriety:  rumor.Impact,
				Confidence: confidence,
			}
		},
		EventPerception: func(Event) PerceptionDelta {
			return PerceptionDelta{Confidence: 100}
		},
	}
}

func (r Rules) normalized() Rules {
	defaults := DefaultRules()
	if r.EligibleForDungeon == nil {
		r.EligibleForDungeon = defaults.EligibleForDungeon
	}
	if r.ItemLegacyValue == nil {
		r.ItemLegacyValue = defaults.ItemLegacyValue
	}
	if r.DungeonName == nil {
		r.DungeonName = defaults.DungeonName
	}
	if r.DungeonDepth == nil {
		r.DungeonDepth = defaults.DungeonDepth
	}
	if r.DungeonAttributes == nil {
		r.DungeonAttributes = defaults.DungeonAttributes
	}
	if r.ShouldDecayDungeon == nil {
		r.ShouldDecayDungeon = defaults.ShouldDecayDungeon
	}
	if r.RumorPerception == nil {
		r.RumorPerception = defaults.RumorPerception
	}
	if r.EventPerception == nil {
		r.EventPerception = defaults.EventPerception
	}
	return r
}
