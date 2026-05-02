# Integration Guide

This guide shows one practical way to embed the Living Legacy Engine in a host game. The engine stays intentionally small: your game owns accounts, combat, maps, crafting recipes, economy, networking, storage, and presentation. The engine owns the durable legacy loop around characters, crafted items, death loot, loot dungeons, events, rumors, memory, perception, snapshots, and queries.

## Core Flow

Create a character and a world:

```go
character, err := engine.NewCharacter("char-ash", "Ash")
if err != nil {
	return err
}

world := engine.NewWorld(character)
```

Register additional actors and characters when your game needs them:

```go
crafter, err := engine.NewActor("player-rin", "Rin", engine.ActorCrafter)
if err != nil {
	return err
}
if err := world.AddActor(crafter); err != nil {
	return err
}

other, err := engine.NewCharacter("char-ember", "Ember")
if err != nil {
	return err
}
if err := world.AddCharacter(other); err != nil {
	return err
}
```

`World.Characters` is the multiplayer source of truth. `World.PrimaryCharacter` is a convenience mirror for games that still center one returning character.

## Spawn Dungeons At World Creation

Loot dungeons are world fixtures. They are not created on death. Spawn them when your map, shard, region, or save file is initialized:

```go
_, err = world.SpawnLootDungeon("ash-cache", "Ash Cache", "ash-road", 3)
if err != nil {
	return err
}
```

Spawned dungeons begin dormant. A dormant dungeon cannot be cleared. It becomes active only after eligible crafted loot is deposited into its area.

## Carry Crafted Loot

The engine only accepts crafted items into the dungeon loop:

```go
blade, err := engine.NewCraftedItem("item-001", "Rin's Glassblade", "player-rin", 4, 12, "blade")
if err != nil {
	return err
}
blade.Rarity = 5

if err := world.CarryItem(character.ID, blade); err != nil {
	return err
}
```

Use `Attributes` for game-specific data such as recipe IDs, bind rules, durability, enchantments, season IDs, or audit metadata.

## Death Deposits Loot

When your combat or survival system decides a character died, call the engine with the character, cause, and area:

```go
dungeon, err := world.KillCharacterByID(character.ID, "the ash stair", "ash-road")
if err != nil {
	return err
}
```

Death with eligible crafted loot deposits that loot into the area's spawned dungeon and activates it. Death without eligible crafted loot returns `ErrNoEligibleLoot` and does not unlock the dungeon, which prevents death farming.

You can also deposit loot directly for non-character actors:

```go
dungeon, err = world.DepositDeathLoot("npc-smith", "ash-road", []engine.CraftedItem{blade}, "ambush")
```

All eligible crafted death loot in the same area shares the same dungeon pool. If five players die in the same area, they have an equal default chance of getting each other's item or their own item back when the dungeon is cleared.

## Respawn Characters

Your game decides where, when, and how a character returns. The engine records the alive/dead state:

```go
if err := world.RespawnCharacterByID(character.ID); err != nil {
	return err
}
```

## Clear Dungeons

When your dungeon system says an actor cleared the dungeon, claim one reward:

```go
loot, err := world.LootDungeon(dungeon.ID, engine.ActorID(character.ID))
if err != nil {
	return err
}
```

The dungeon chooses one unclaimed deposited item through `Rules.SelectLoot`, creates a reward through `Rules.LootReward`, applies legacy to the looter if they are a registered character, and locks the dungeon to that looter.

The lock prevents farming. A locked dungeon only reopens when the same actor dies again with eligible crafted loot.

## Customize Rules

Rules are runtime policy. Snapshots store world state, not rule functions. Reapply rules when your server starts or a save is restored:

```go
rules := engine.DefaultRules()

rules.EligibleForDungeon = func(item engine.CraftedItem) bool {
	return item.CrafterID != "" && item.Attributes["no_legacy"] != "true"
}

rules.ItemLegacyValue = func(item engine.CraftedItem) int {
	return item.Power + item.Rarity
}

rules.SelectLoot = func(dungeon engine.LootDungeon, looter engine.ActorID) int {
	return 0 // replace with seeded randomness or server-authoritative selection
}

rules.LootReward = func(source engine.CraftedItem) engine.CraftedItem {
	reward := source
	reward.Power++
	return reward
}

world = engine.NewWorldWithRules(character, rules)
```

`LootReward` may upgrade or transform an item, but it cannot return a reward with lower rarity than the dropped source item. The engine returns `ErrRewardDowngrade` if that floor is violated.

## Events, Rumors, Memory, And Perception

The engine records core events automatically. Host games can add their own:

```go
event, perceptions, err := world.RecordObservedEvent(engine.Event{
	Type:        engine.EventDungeonLooted,
	Description: "Ash cleared the Ash Cache.",
	Subject:     engine.CharacterRef(character),
	Object:      engine.DungeonRef(dungeon),
}, "witness-1", "witness-2")
if err != nil {
	return err
}

_ = event
_ = perceptions
```

Use rumors for uncertain claims:

```go
rumor, err := engine.NewRumor("rumor-001", "witness-1", "Ash returned with Rin's blade.", 0.8, 4, time.Now())
if err != nil {
	return err
}
rumor.Subject = engine.CharacterRef(character)
rumor.Object = engine.ItemRef(blade)

if err := world.AddRumor(rumor); err != nil {
	return err
}
_, err = world.SpreadRumor(rumor.ID, "witness-2", 0.1)
```

Generate setting history when NPC dialogue should come from LLE-owned NPCs, factions, and ties:

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

For immediate NPC chatter, ask for a generated remark. It prefers known rumors, then generated world-history ties, then runtime history, then injected lore:

```go
remark := world.GenerateRemark(engine.TextGenerationContext{
	SourceID: "guard-2",
	TargetID: "npc-rin",
})
```

Use lore hooks when your game wants to add map, quest, biome, or presentation context to LLE-owned history:

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

world.SetRules(rules)
```

Generate rumors from runtime event history when dialogue should track what happened during play:

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
_ = rumors
```

For a single host event, record runtime history and generate the rumor in one call:

```go
event, rumor, err := world.RecordEventWithGeneratedRumor(engine.Event{
	Type:        engine.EventRespawned,
	Description: "Ash returned under the old moon.",
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

Actors learn events and rumors through memory. Rumors and observed events can also change perception through `Rules.RumorPerception` and `Rules.EventPerception`.

## Query State

Queries return cloned data so callers can safely inspect state:

```go
deposits := world.UnclaimedDepositsInArea("ash-road")
history := world.EventsForActor(engine.ActorID(character.ID))
rumors := world.RumorsAbout(string(character.ID))
opinions := world.PerceptionsAbout(string(character.ID))
memory := world.MemoriesOf("witness-2")
```

Use these helpers for UI, audit logs, admin tools, analytics, quest triggers, or periodic simulation jobs.

## Save And Restore

Snapshots are portable data:

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

restored.SetRules(rules)
```

For a simple file-backed save:

```go
store := engine.NewFileSnapshotStore("saves/world.json")
if err := store.Save(world.Snapshot()); err != nil {
	return err
}

loaded, err := store.Load()
if err != nil {
	return err
}
world, err = engine.RestoreWorld(loaded)
```

## Shared Server Access

Wrap the world when multiple goroutines may read or mutate it:

```go
safe := engine.NewSafeWorld(world)

err := safe.Update(func(w *engine.World) error {
	_, err := w.KillCharacterByID("char-ash", "the ash stair", "ash-road")
	return err
})
if err != nil {
	return err
}

snapshot := safe.Snapshot()
_ = snapshot
```

Keep long-running work outside `View` and `Update` closures. Use the closures to touch engine state, then release the lock.

## Error Handling

The engine exposes sentinel errors for common integration decisions:

```go
loot, err := world.LootDungeon("ash-cache", "char-ash")
if err != nil {
	if errors.Is(err, engine.ErrDungeonDormant) {
		return nil // no death loot has activated it yet
	}
	if errors.Is(err, engine.ErrDungeonLocked) {
		return nil // locked until this clearer dies again with crafted loot
	}
	return err
}
_ = loot
```

Common sentinels include `ErrNoEligibleLoot`, `ErrNoLootDungeonInArea`, `ErrDungeonDormant`, `ErrDungeonLocked`, `ErrDungeonDecayed`, `ErrDungeonNotFound`, `ErrCharacterNotFound`, and `ErrRewardDowngrade`.

## Integration Checklist

- Spawn loot dungeons during world, area, shard, or save creation.
- Register characters and actors with stable IDs from your game.
- Convert crafted game items into `CraftedItem` values at carry/death boundaries.
- Call `KillCharacterByID` only after your game has resolved death.
- Treat `ErrNoEligibleLoot` as a non-mutating failed deposit, not as a successful death record.
- Call `RespawnCharacterByID` when your game respawns the character.
- Call `LootDungeon` only after your dungeon gameplay resolves a clear.
- Reapply runtime `Rules` after restoring a snapshot.
- Save snapshots wherever your game already stores world state.
- Use query helpers for read models and UI, not as the only source of truth.
