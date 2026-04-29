package engine

import "errors"

var (
	ErrActorRequired       = errors.New("actor id is required")
	ErrCharacterRequired   = errors.New("character id is required")
	ErrCharacterNotFound   = errors.New("character not found")
	ErrDungeonDecayed      = errors.New("dungeon is decayed")
	ErrDungeonDormant      = errors.New("dungeon is dormant")
	ErrDungeonLocked       = errors.New("dungeon is locked")
	ErrDungeonNotFound     = errors.New("dungeon not found")
	ErrDuplicateDungeon    = errors.New("dungeon already exists")
	ErrNoEligibleLoot      = errors.New("no eligible crafted loot")
	ErrNoLootDungeonInArea = errors.New("no loot dungeon in area")
	ErrRewardDowngrade     = errors.New("loot reward downgraded rarity")
)
