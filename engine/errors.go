package engine

import "errors"

var (
	// ErrActorRequired indicates an operation needs a non-empty actor ID.
	ErrActorRequired = errors.New("actor id is required")
	// ErrCharacterRequired indicates an operation needs a non-empty character ID.
	ErrCharacterRequired = errors.New("character id is required")
	// ErrCharacterNotFound indicates a character ID is not registered.
	ErrCharacterNotFound = errors.New("character not found")
	// ErrDungeonDecayed indicates the dungeon can no longer accept or return loot.
	ErrDungeonDecayed = errors.New("dungeon is decayed")
	// ErrDungeonDormant indicates the dungeon has not been activated by death loot.
	ErrDungeonDormant = errors.New("dungeon is dormant")
	// ErrDungeonLocked indicates the dungeon is locked to the actor who last cleared it.
	ErrDungeonLocked = errors.New("dungeon is locked")
	// ErrDungeonNotFound indicates a dungeon ID is not registered.
	ErrDungeonNotFound = errors.New("dungeon not found")
	// ErrDuplicateDungeon indicates a dungeon ID is already registered.
	ErrDuplicateDungeon = errors.New("dungeon already exists")
	// ErrNoEligibleLoot indicates no carried items passed the eligibility rule.
	ErrNoEligibleLoot = errors.New("no eligible crafted loot")
	// ErrNoLootDungeonInArea indicates an area has no available spawned loot dungeon.
	ErrNoLootDungeonInArea = errors.New("no loot dungeon in area")
	// ErrRewardDowngrade indicates LootReward returned lower rarity than the source item.
	ErrRewardDowngrade = errors.New("loot reward downgraded rarity")
)
