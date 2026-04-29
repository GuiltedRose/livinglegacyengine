package engine

import "time"

type ItemEligibilityRule func(item CraftedItem) bool
type ItemLegacyRule func(item CraftedItem) int
type DungeonDecayRule func(dungeon LootDungeon, now time.Time) bool
type LootSelectionRule func(dungeon LootDungeon, looter ActorID) int
type LootRewardRule func(source CraftedItem) CraftedItem
type RumorPerceptionRule func(rumor Rumor) PerceptionDelta
type EventPerceptionRule func(event Event) PerceptionDelta

type Rules struct {
	EligibleForDungeon ItemEligibilityRule
	ItemLegacyValue    ItemLegacyRule
	ShouldDecayDungeon DungeonDecayRule
	SelectLoot         LootSelectionRule
	LootReward         LootRewardRule
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
		ShouldDecayDungeon: func(LootDungeon, time.Time) bool {
			return false
		},
		SelectLoot: func(dungeon LootDungeon, looter ActorID) int {
			return selectLootIndex(dungeon, looter)
		},
		LootReward: func(source CraftedItem) CraftedItem {
			return source
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
	if r.ShouldDecayDungeon == nil {
		r.ShouldDecayDungeon = defaults.ShouldDecayDungeon
	}
	if r.SelectLoot == nil {
		r.SelectLoot = defaults.SelectLoot
	}
	if r.LootReward == nil {
		r.LootReward = defaults.LootReward
	}
	if r.RumorPerception == nil {
		r.RumorPerception = defaults.RumorPerception
	}
	if r.EventPerception == nil {
		r.EventPerception = defaults.EventPerception
	}
	return r
}
