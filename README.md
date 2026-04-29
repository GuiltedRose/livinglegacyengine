# Living Legacy Engine
**A standalone, game-agnostic Go engine for death loops, crafted loot dungeons, and persistent legacy**

*Designed for open innovation — no patents, no restrictions.*

## Overview
The Living Legacy Engine is a small standalone Go module for games where one character dies often and the world remembers what those deaths leave behind.

Legacy is no longer inherited through heirs or bloodlines. The same character returns again and again, while the durable legacy comes from loot dungeons created at death. Those dungeons contain only actor-crafted items, turning the crafting economy into the historical record of the world.

The engine is intentionally agnostic about genre, networking, storage, rendering, combat, economy, and quest structure. A roguelike, MUD, MMO, tabletop campaign manager, survival game, or narrative sim should all be able to embed the same core loop and attach their own systems around it.

## Core Principles
- **Actor Neutrality**: No actor is inherently special. Identity, status, and reputation must be earned through events.
- **One Returning Character**: There is no inheritance chain. The same character respawns after death.
- **Legacy Through Loot**: Death creates loot dungeons. Clearing those dungeons feeds the character's legacy score.
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

## Integration Boundaries

The engine owns:
- Actor registry
- Target references for actors, items, dungeons, regions, events, and host-game concepts
- Character death/respawn state
- Crafted item eligibility
- Loot dungeon creation and looting
- Legacy scoring
- Rumor creation and propagation
- Actor memory records
- Actor-scoped perception records
- Structured event history
- Event observation that teaches memory and updates perception
- Snapshot/restore state
- Timeline and state query helpers

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

world := engine.NewWorldWithRules(character, rules)
```

Available hooks:
- `EligibleForDungeon`: decides which carried items become dungeon loot
- `ItemLegacyValue`: scores items for crafter fame and dungeon legacy
- `DungeonName`: names death-created dungeons
- `DungeonDepth`: maps character/run state to dungeon depth
- `DungeonAttributes`: attaches host-game metadata to new dungeons
- `ShouldDecayDungeon`: decides when sealed dungeons decay
- `RumorPerception`: maps a rumor into perception changes
- `EventPerception`: maps an event into perception changes

Rules are runtime policy, not saved state. Snapshots store the world data; host games should restore the world and then apply the rules for that game/version.

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

## Key Systems

### 1. Character Death Loop
The engine tracks one character:
- Alive/dead state
- Death count
- Inventory
- Recovered items
- Legacy score

When the character dies, their current crafted inventory becomes the seed of a loot dungeon.

### 2. Crafted Item System
Items must have:
- Stable item ID
- Display name
- Actor crafter ID
- Quality
- Power
- Optional tags
- Optional attributes

Only actor-crafted items are accepted into the dungeon/legacy loop. In a multiplayer game the actor may be a player. In another game it may be a settlement, guild, NPC artisan, procedural civilization, or imported campaign entity.

### 3. Loot Dungeon System
On death, the engine creates a sealed loot dungeon containing the crafted items carried by the character.

Dungeons track:
- Origin run
- Depth
- Crafted item contents
- Legacy value
- Looted state
- Looter ID

When a dungeon is looted, the same character's legacy score increases.

### 4. Event Log
The world records major events:
- Crafted item added
- Character death
- Dungeon creation
- Character respawn
- Dungeon looted

Events include readable descriptions plus structured subject/object IDs and string data so host games can build logs, notifications, analytics, admin tools, or persistence without parsing prose.

Host games can also feed their own events into the engine:

```go
event, perceptions, err := world.RecordObservedEvent(engine.Event{
	Type:        engine.EventDungeonLooted,
	Description: "Ash cleared the sealed cache.",
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

These helpers are intentionally small and composable. Host games can build richer search, indexing, UI views, or analytics on top.

### 10. Snapshot System
`World.Snapshot` and `RestoreWorld` provide plain serializable state. The engine does not choose a storage backend; host games can encode snapshots as JSON, save them in SQL, sync them through a server, or keep them in memory for tests.

Snapshots include a `version` field. The current version is exposed as `CurrentSnapshotVersion`. Restore runs snapshots through `MigrateSnapshot` and `ValidateSnapshot` before hydrating a world. Version `0` is treated as a legacy pre-version snapshot and migrated to the current version; snapshots newer than the engine supports fail clearly.

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
