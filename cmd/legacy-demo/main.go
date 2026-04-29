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

	blade, err := engine.NewCraftedItem("iron-song", "Iron Song", "actor-ada", 4, 12, "sword", "crafted")
	if err != nil {
		log.Fatal(err)
	}
	charm, err := engine.NewCraftedItem("lamp-charm", "Lamp Charm", "actor-bryn", 2, 6, "charm", "crafted")
	if err != nil {
		log.Fatal(err)
	}

	if err := world.AddCraftedItem(blade); err != nil {
		log.Fatal(err)
	}
	if err := world.AddCraftedItem(charm); err != nil {
		log.Fatal(err)
	}

	dungeon, err := world.KillCharacter("the ash stair")
	if err != nil {
		log.Fatal(err)
	}
	world.RespawnCharacter()

	loot, err := world.LootDungeon(dungeon.ID, "actor-cora")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s created with %d crafted items\n", dungeon.Name, len(loot))
	fmt.Printf("Legacy score: %d\n", world.Character.LegacyScore)
}
