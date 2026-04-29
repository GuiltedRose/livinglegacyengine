package engine

import "time"

// ItemEligibilityRule decides whether a carried crafted item enters a death loot pool.
type ItemEligibilityRule func(item CraftedItem) bool

// ItemLegacyRule scores an item for dungeon legacy value and crafter fame.
type ItemLegacyRule func(item CraftedItem) int

// DungeonDecayRule decides whether an active dungeon should decay.
type DungeonDecayRule func(dungeon LootDungeon, now time.Time) bool

// LootSelectionRule selects an unclaimed item index from a dungeon pool.
type LootSelectionRule func(dungeon LootDungeon, looter ActorID) int

// LootRewardRule maps a deposited item to the reward returned by a clear.
type LootRewardRule func(source CraftedItem) CraftedItem

// DepositEventsRule emits extra events when loot enters a dungeon pool.
type DepositEventsRule func(deposit DepositedLoot) []Event

// ClaimEventsRule emits extra events when deposited loot is claimed.
type ClaimEventsRule func(deposit DepositedLoot, reward CraftedItem) []Event

// DepositRumorsRule emits rumors when loot enters a dungeon pool.
type DepositRumorsRule func(deposit DepositedLoot) []Rumor

// ClaimRumorsRule emits rumors when deposited loot is claimed.
type ClaimRumorsRule func(deposit DepositedLoot, reward CraftedItem) []Rumor

// RumorPerceptionRule maps a rumor to a perception change.
type RumorPerceptionRule func(rumor Rumor) PerceptionDelta

// EventPerceptionRule maps an event to a perception change.
type EventPerceptionRule func(event Event) PerceptionDelta

// Rules contains host-game policy hooks. Nil hooks are filled from DefaultRules.
type Rules struct {
	EligibleForDungeon ItemEligibilityRule
	ItemLegacyValue    ItemLegacyRule
	ShouldDecayDungeon DungeonDecayRule
	SelectLoot         LootSelectionRule
	LootReward         LootRewardRule
	DepositEvents      DepositEventsRule
	ClaimEvents        ClaimEventsRule
	DepositRumors      DepositRumorsRule
	ClaimRumors        ClaimRumorsRule
	RumorPerception    RumorPerceptionRule
	EventPerception    EventPerceptionRule
}

// DefaultRules returns the engine's baseline policy set.
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
		DepositEvents: func(DepositedLoot) []Event {
			return nil
		},
		ClaimEvents: func(DepositedLoot, CraftedItem) []Event {
			return nil
		},
		DepositRumors: func(DepositedLoot) []Rumor {
			return nil
		},
		ClaimRumors: func(DepositedLoot, CraftedItem) []Rumor {
			return nil
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
	if r.DepositEvents == nil {
		r.DepositEvents = defaults.DepositEvents
	}
	if r.ClaimEvents == nil {
		r.ClaimEvents = defaults.ClaimEvents
	}
	if r.DepositRumors == nil {
		r.DepositRumors = defaults.DepositRumors
	}
	if r.ClaimRumors == nil {
		r.ClaimRumors = defaults.ClaimRumors
	}
	if r.RumorPerception == nil {
		r.RumorPerception = defaults.RumorPerception
	}
	if r.EventPerception == nil {
		r.EventPerception = defaults.EventPerception
	}
	return r
}
