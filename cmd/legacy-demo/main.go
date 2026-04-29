package main

import (
	"fmt"
	"log"

	"livinglegacyengine/engine"
)

func main() {
	character, err := engine.NewCharacter("one-who-returns", "Mara")
	if err != nil {
		log.Fatal(err)
	}

	world := engine.NewWorld(character)
	dungeon, err := world.SpawnLootDungeon("ash-stair-cache", "Ash Stair Cache", "ash-stair", 3)
	if err != nil {
		log.Fatal(err)
	}

	blade, err := engine.NewCraftedItem("iron-song", "Iron Song", "actor-ada", 4, 12, "sword", "crafted")
	if err != nil {
		log.Fatal(err)
	}
	charm, err := engine.NewCraftedItem("lamp-charm", "Lamp Charm", "actor-bryn", 2, 6, "charm", "crafted")
	if err != nil {
		log.Fatal(err)
	}

	if err := world.CarryItem(character.ID, blade); err != nil {
		log.Fatal(err)
	}
	if err := world.CarryItem(character.ID, charm); err != nil {
		log.Fatal(err)
	}

	if _, err := world.KillCharacterByID(character.ID, "the ash stair", "ash-stair"); err != nil {
		log.Fatal(err)
	}
	if err := world.RespawnCharacterByID(character.ID); err != nil {
		log.Fatal(err)
	}

	loot, err := world.LootDungeon(dungeon.ID, engine.ActorID(character.ID))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s cleared for %d crafted item\n", dungeon.Name, len(loot))
	fmt.Printf("Dungeon status: %s\n", world.Dungeons[dungeon.ID].Status)
	fmt.Printf("Legacy score: %d\n", world.PrimaryCharacter.LegacyScore)
}
