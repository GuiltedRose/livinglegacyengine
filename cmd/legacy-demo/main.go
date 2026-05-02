package main

import (
	"fmt"
	"log"

	"github.com/GuiltedRose/livinglegacyengine/engine"
)

func main() {
	character, err := engine.NewCharacter("one-who-returns", "Mara")
	if err != nil {
		log.Fatal(err)
	}

	rules := engine.DefaultRules()
	rules.LoreFacts = func(context engine.TextGenerationContext) []engine.LoreFact {
		if context.TargetID != "ash-stair-cache" {
			return nil
		}
		return []engine.LoreFact{{
			ID:     "ash-stair-warning",
			Text:   "Locals say the stair keeps the names of everyone who paid it in iron.",
			Object: engine.NewTargetRef("ash-stair-cache", engine.TargetDungeon, "Ash Stair Cache"),
			Weight: 10,
		}}
	}

	world := engine.NewWorldWithRules(character, rules)
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
		log.Fatal(err)
	}

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
	fmt.Printf("Generated %d actors and %d ties\n", len(history.Actors), len(history.Ties))

	remark := world.GenerateRemark(engine.TextGenerationContext{
		SourceID: "ash-stair-witness",
		TargetID: "npc-rin",
	})
	fmt.Printf("Remark: %s\n", remark.Text)

	rumors, err := world.GenerateRumorsFromEvents(engine.EventRumorOptions{
		SourceID: "ash-stair-witness",
		TargetID: string(dungeon.ID),
		Limit:    2,
		Truth:    0.85,
		Impact:   3,
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, rumor := range rumors {
		fmt.Printf("Rumor: %s\n", rumor.Description)
	}
}
