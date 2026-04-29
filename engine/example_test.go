package engine_test

import (
	"fmt"

	"livinglegacyengine/engine"
)

func ExampleWorld_lootDungeonFlow() {
	character, _ := engine.NewCharacter("ash", "Ash")
	world := engine.NewWorld(character)
	_, _ = world.SpawnLootDungeon("road-cache", "Road Cache", "road", 1)

	item, _ := engine.NewCraftedItem("iron-pin", "Iron Pin", "crafter-1", 2, 3)
	_ = world.CarryItem(character.ID, item)
	dungeon, _ := world.KillCharacterByID(character.ID, "the road", "road")
	_ = world.RespawnCharacterByID(character.ID)
	loot, _ := world.LootDungeon(dungeon.ID, engine.ActorID(character.ID))

	fmt.Println(loot[0].Rarity >= item.Rarity)
	fmt.Println(world.Dungeons[dungeon.ID].Status)
	// Output:
	// true
	// locked
}
