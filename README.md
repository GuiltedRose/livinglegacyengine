# Living Legacy Engine
**A standalone, game-agnostic Go engine for death loops, crafted loot dungeons, and persistent legacy**

*Designed for open innovation — no patents, no restrictions.*

## Overview
The Living Legacy Engine is a small standalone Go module for games where characters die often and the world remembers what those deaths leave behind.

Legacy is no longer inherited through heirs or bloodlines. Characters return again and again, while the durable legacy comes from pre-spawned loot dungeons that open only when crafted loot is lost to death. Those dungeons contain only actor-crafted items, turning the crafting economy into the historical record of the world.

The engine is intentionally agnostic about genre, networking, storage, rendering, combat, economy, and quest structure. A roguelike, MUD, MMO, tabletop campaign manager, survival game, or narrative sim should all be able to embed the same core loop and attach their own systems around it.

## Core Principles
- **Actor Neutrality**: No actor is inherently special. Identity, status, and reputation must be earned through events.
- **Returning Characters**: There is no inheritance chain. Characters respawn after death.
- **Legacy Through Loot**: Death deposits crafted loot into existing area dungeons. Clearing those dungeons feeds legacy.
- **Actor-Crafted Economy**: Loot dungeons are populated only by crafted items made by known actors.
- **Dynamic World State**: Factions, regions, and dungeons can evolve around actor action.
- **Emergent Narrative**: Every world can be seeded, simulated, and integrated into a larger game.
- **Game-Agnostic Core**: The engine exposes plain Go types and does not require a server, database, renderer, or rules framework.

## Go Module

```bash
go test ./...
go run ./cmd/legacy-demo
```

The public API currently lives in `engine/` and is intentionally dependency-free so it can be embedded into a server, local game, simulation worker, command loop, or tool.

For a complete host-game wiring path, see [docs/integration.md](docs/integration.md).

## Integration Boundaries

The engine owns:
- Actor registry
- Engine-owned NPC, faction, and tie generation
- Character registry, plus an explicit `PrimaryCharacter` convenience field
- Target references for actors, items, dungeons, regions, events, and host-game concepts
- Character death/respawn state
- Crafted item eligibility
- Loot dungeon spawning, death loot deposits, clearing, and anti-farm locks
- Legacy scoring
- Rumor creation and propagation
- Generated rumor text from structured world history
- Actor memory records
- Actor-scoped perception records
- Structured event history
- Event observation that teaches memory and updates perception
- Snapshot/restore state
- Timeline and state query helpers
- Optional concurrency wrapper
- Optional file-backed snapshot store

The host game owns:
- Storage backend
- Authentication and player accounts
- Combat, travel, crafting recipes, and economy rules
- User interface
- Multiplayer synchronization
- World map, regions, factions, and narrative presentation

Game-specific data can be attached through `Attributes` maps on items and dungeons, or through structured event `Data`.

## Target References

Events, rumors, and perceptions can point at generic `TargetRef` values instead of relying only on raw IDs.

Target kinds include:
- Actor
- Character
- Item
- Dungeon
- Region
- Faction
- Event
- Custom

Helpers such as `ActorRef`, `CharacterRef`, `ItemRef`, and `DungeonRef` make common targets easy to attach. Legacy string IDs remain available for simple integrations and compatibility.

## Rule Hooks

Host games can keep the default behavior or provide a `Rules` value when creating a world:

```go
rules := engine.DefaultRules()
rules.ItemLegacyValue = func(item engine.CraftedItem) int {
	return item.Power
}
rules.EligibleForDungeon = func(item engine.CraftedItem) bool {
	return item.Attributes["binds_legacy"] == "true"
}
rules.LootReward = func(source engine.CraftedItem) engine.CraftedItem {
	reward := source
	reward.Rarity = source.Rarity + 1
	return reward
}
rules.DepositEvents = func(deposit engine.DepositedLoot) []engine.Event {
	if deposit.Item.Rarity < 5 {
		return nil
	}
	return []engine.Event{{
		Type:        engine.EventLootDeposited,
		Description: "A rare crafted item entered the pool.",
		Object:      engine.ItemRef(deposit.Item),
	}}
}

world := engine.NewWorldWithRules(character, rules)
```

Available hooks:
- `EligibleForDungeon`: decides which carried items become dungeon loot
- `ItemLegacyValue`: scores items for crafter fame and dungeon legacy
- `ShouldDecayDungeon`: decides when active dungeons decay
- `SelectLoot`: chooses which pooled item is returned on clear
- `LootReward`: maps a dropped crafted item to the reward returned by a dungeon clear
- `DepositEvents`: emits extra events when loot enters a dungeon pool
- `ClaimEvents`: emits extra events when deposited loot is claimed
- `DepositRumors`: emits rumors when loot enters a dungeon pool
- `ClaimRumors`: emits rumors when deposited loot is claimed
- `LoreFacts`: injects host-game lore into generated text
- `EventRumorText`: turns runtime events and lore into rumor wording
- `RumorPerception`: maps a rumor into perception changes
- `EventPerception`: maps an event into perception changes

`LootReward` may upgrade a reward, but it may not return an item with lower rarity than the dropped source item.

Rules are runtime policy, not saved state. Snapshots store the world data; host games should restore the world and then apply the rules for that game/version.

## World History And Generated Remarks

The engine can generate NPCs, factions, and ties between them. These ties are stored as `WorldHistory`, not runtime events, so NPC remarks can be grounded in setting history instead of stale hardcoded lines.

```go
history, err := world.GenerateWorldHistory(engine.WorldGenerationOptions{
	Seed:         12,
	NPCCount:     3,
	FactionCount: 2,
	TieCount:     5,
	NPCNames:     []string{"Rin", "Sable", "Oren"},
	FactionNames: []string{"Lantern Compact", "Ash Stair Keepers"},
	Era:          "ash stair founding",
})
if err != nil {
	return err
}

_ = history
```

For lightweight NPC chatter, use `GenerateRemark`. It prefers known rumors, then generated world-history ties, then runtime history, then injected lore:

```go
remark := world.GenerateRemark(engine.TextGenerationContext{
	SourceID: "guard-2",
	TargetID: "npc-rin",
})
```

Host games can still inject extra lore into generated text. The LLE owns generated NPCs, factions, and ties, while `LoreFacts` lets a game add world-map context, quest state, biome details, or presentation flavor.

```go
rules := engine.DefaultRules()
rules.LoreFacts = func(context engine.TextGenerationContext) []engine.LoreFact {
	if context.TargetID != "ash-cache" {
		return nil
	}
	return []engine.LoreFact{{
		Text:   "The stair keepers call this an unpaid oath.",
		Object: engine.NewTargetRef("ash-cache", engine.TargetDungeon, "Ash Cache"),
		Weight: 10,
	}}
}
```

## Event-Derived Rumors

When a host game wants rumors from runtime events, the same text hooks can turn recorded event history into rumor text.

```go
rumors, err := world.GenerateRumorsFromEvents(engine.EventRumorOptions{
	SourceID: "witness-1",
	TargetID: "ash-cache",
	Limit:    3,
	Truth:    0.85,
	Impact:   4,
})
if err != nil {
	return err
}
```

`GenerateRumorsFromEvents` walks structured events, generates rumor descriptions, stores the rumors, and links them back to event targets. Host games can replace `EventRumorText` for their own voice, or keep the default generator and use `LoreFacts` to inject regional lore, faction beliefs, character relationships, prophecies, biome facts, quest state, or anything else the game owns.

For the single-pass path, record a host event and mint its rumor together:

```go
event, rumor, err := world.RecordEventWithGeneratedRumor(engine.Event{
	Type:        engine.EventRespawned,
	Description: "Mara returned under the old moon.",
	Subject:     engine.CharacterRef(character),
}, engine.EventRumorOptions{
	SourceID: "witness-1",
	Truth:    0.9,
	Impact:   3,
})
if err != nil {
	return err
}

_ = event
_ = rumor
```

## Persistence

Snapshots are plain serializable world state:

```go
snapshot := world.Snapshot()

data, err := engine.MarshalSnapshot(snapshot)
if err != nil {
	return err
}

loaded, err := engine.UnmarshalSnapshot(data)
if err != nil {
	return err
}

restored, err := engine.RestoreWorld(loaded)
if err != nil {
	return err
}
```

The JSON helpers use Go's standard `encoding/json` package. Host games can store the bytes anywhere: files, SQL, object storage, save slots, test fixtures, or network messages.

A small file-backed store is included for simple integrations and tests:

```go
store := engine.NewFileSnapshotStore("saves/world.json")
if err := store.Save(world.Snapshot()); err != nil {
	return err
}
loaded, err := store.Load()
```

For shared server access, wrap a world with `NewSafeWorld` and use `View`/`Update` closures to guard concurrent reads and writes.

## Key Systems

### 1. Character Death Loop
The engine tracks a registry of characters:
- Alive/dead state
- Death count
- Inventory
- Recovered items
- Legacy score

`World.Characters` is the multiplayer source of truth. `World.PrimaryCharacter` is an explicit convenience mirror for games that still want a main returning character.

When a character dies in an area, eligible crafted inventory is deposited into a spawned loot dungeon in that area. Death without crafted loot does not open or unlock the dungeon.

### 2. Crafted Item System
Items must have:
- Stable item ID
- Display name
- Actor crafter ID
- Rarity
- Quality
- Power
- Optional tags
- Optional attributes

Only actor-crafted items are accepted into the dungeon/legacy loop. In a multiplayer game the actor may be a player. In another game it may be a settlement, guild, NPC artisan, procedural civilization, or imported campaign entity.

### 3. Loot Dungeon System
Loot dungeons are spawned with the world, usually by area:

```go
dungeon, err := world.SpawnLootDungeon("ash-cache", "Ash Cache", "ash-road", 3)
```

They begin dormant and cannot be cleared. When an actor dies in the area with eligible crafted loot, that loot is deposited into the area's dungeon and the dungeon becomes active:

```go
dungeon, err := world.KillCharacterByID(character.ID, "the ash stair", "ash-road")
```

Other actors can also deposit death loot directly:

```go
dungeon, err := world.DepositDeathLoot("actor-id", "ash-road", craftedItems, "ambush")
```

All crafted death loot in the same area goes into the same dungeon pool. If five actors die in the same area, their eligible crafted items share that pool, so a clearer can recover anyone's item, including their own. The default selection chooses one item from the pool without owner weighting; host games can layer presentation, reward tables, rarity upgrades, or additional randomness around the same pooled state.

Dungeons track:
- Area
- Depth
- Current unclaimed crafted item pool
- Deposit history
- Legacy value
- Status
- Last looter ID
- Lock owner

When a dungeon is cleared, the engine selects one item from the pooled crafted loot and then locks the dungeon to the actor who cleared it. A locked dungeon cannot be farmed. It only unlocks when that same actor dies again with eligible crafted loot, which also deposits that new loot into the area pool.

Reward generation has a rarity floor: the returned reward must have rarity greater than or equal to the dropped source item. The default rule returns the source item unchanged.

Each deposited item is also stored as a `DepositedLoot` record. The record keeps:
- Deposit ID
- Crafted item
- Dropping actor
- Area and dungeon IDs
- Death cause
- Deposit time
- Claiming actor
- Claim time
- Attributes

`LootDungeon.Items` is a compatibility mirror of the current unclaimed pool. `LootDungeon.Deposits` is the durable provenance ledger, including claimed loot.

Deposit query helpers:
- `DepositsByDropper`
- `DepositsByClaimer`
- `DepositsInDungeon`
- `UnclaimedDepositsInDungeon`
- `DepositsInArea`
- `UnclaimedDepositsInArea`
- `DepositsForItem`

Deposit and claim hooks can emit additional events or rumors. This is useful for rare item callouts, crafting fame, regional gossip, audit logs, and world-history UI without forcing every game to use the same rarity scale.

### 4. Event Log
The world records major events:
- Crafted item added
- Character death
- Dungeon spawn
- Loot deposit
- Character respawn
- Dungeon looted
- Dungeon locked/unlocked

Events include readable descriptions plus structured subject/object IDs and string data so host games can build logs, notifications, analytics, admin tools, or persistence without parsing prose.

Host games can also feed their own events into the engine:

```go
event, perceptions, err := world.RecordObservedEvent(engine.Event{
	Type:        engine.EventDungeonLooted,
	Description: "Ash cleared the cache.",
	Subject:     engine.CharacterRef(character),
}, "witness-1", "witness-2")
```

`RecordEvent` stores an event without observers. `ObserveEvent` teaches one actor about an event and applies `EventPerception`. `RecordObservedEvent` does both in one call.

### 5. Actor System
Actors are generic entities that can make items, start rumors, receive memories, or represent host-game concepts.

Actor kinds include:
- Character
- Crafter
- Faction
- Settlement
- Organization
- Unknown

The kind list is intentionally broad and non-exclusive. Host games can attach their own identifiers and classifications through actor attributes.

### 6. Rumor System
Rumors are uncertain claims that can move between actors.

Rumors track:
- Source actor
- Optional subject/object IDs
- Description
- Impact
- Truth from `0.0` to `1.0`
- Spread count
- Attributes

Spreading a rumor can distort its truth value and teaches the recipient's memory.

### 7. Memory System
Each actor can have a memory record containing known events and known rumors. Memory is intentionally factual-light: it records what an actor has heard or learned, not whether the host game must treat that information as objectively true.

This lets a host game build reputation, unreliable witnesses, regional knowledge, quest hints, history logs, or social simulation on top of the same core data.

### 8. Perception System
Perception records what one actor thinks about another actor, object, dungeon, region, or host-game target ID.

Perception tracks:
- Trust
- Fear
- Respect
- Affinity
- Notoriety
- Confidence
- Attributes

Scores are clamped from `-100` to `100`. Perception is scoped by observer and target, so two actors can hold completely different opinions about the same target.

Rumor spread applies `RumorPerception` automatically for the receiving actor. Host games can also call `ApplyEventToPerception` when an actor observes or learns about an event.

### 9. Query Helpers
The engine exposes clone-returning query helpers for common timeline and state questions:

- `EventsMatching`
- `EventsForTarget`
- `EventsForActor`
- `RumorsMatching`
- `RumorsAbout`
- `MemoriesOf`
- `PerceptionsByObserver`
- `PerceptionsAbout`
- `DepositsByDropper`
- `DepositsByClaimer`
- `DepositsInDungeon`
- `UnclaimedDepositsInDungeon`
- `DepositsInArea`
- `UnclaimedDepositsInArea`
- `DepositsForItem`

These helpers are intentionally small and composable. Host games can build richer search, indexing, UI views, or analytics on top.

### 10. Snapshot System
`World.Snapshot` and `RestoreWorld` provide plain serializable state. The engine does not choose a storage backend; host games can encode snapshots as JSON, save them in SQL, sync them through a server, or keep them in memory for tests.

Snapshots include a `version` field. The current version is exposed as `CurrentSnapshotVersion`. Restore runs snapshots through `MigrateSnapshot` and `ValidateSnapshot` before hydrating a world. Version `0` is treated as a legacy pre-version snapshot and migrated to the current version; snapshots newer than the engine supports fail clearly.

Snapshots write `primary_character` and keep the older `character` field as a compatibility bridge for pre-registry saves.

Query indices are intentionally private runtime state. They are rebuilt when a snapshot is restored, so saved data stays portable and does not depend on internal indexing details.

## Optional Modules

- **Memory Codex**: A timeline of deaths, dungeon clears, and famous crafted items
- **Perception Views**: Different views of the returning character by actor, region, or faction
- **World Event Log**: A dynamic feed of political or historical change
- **Legacy Artifacts**: Items, buildings, or myths that persist across death cycles
- **Dungeon Decay**: Loot dungeons can expire, mutate, or be absorbed by factions
- **Crafting Fame**: Crafters gain renown when their items appear in famous dungeons
- **Rule Hooks**: Host games can replace legacy scoring, dungeon naming, dungeon eligibility, dungeon metadata, and decay policy
- **Timeline Queries**: Host games can inspect events, rumors, memories, and perception without mutating engine state
- **Snapshot Versioning**: Saved worlds can evolve through explicit migration and validation

## Inspirations
- *Outward* – identity and consequence
- *Dwarf Fortress* – world memory and procedural legacy
- Long-running worlds – asynchronous progress, crafting economies, and durable memory

# License & Use
The **Living Legacy Engine** is released under AGPLv3.
> **Game mechanics are not inventions — they are building blocks for everyone. Innovation must remain open.**

---

*Designed by players, for creators. Let's build better worlds together.*
