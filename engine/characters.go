package engine

import "fmt"

func (w *World) AddCharacter(character Character) error {
	if character.ID == "" {
		return fmt.Errorf("%w", ErrCharacterRequired)
	}
	if character.Name == "" {
		return fmt.Errorf("character name is required")
	}
	w.Characters[character.ID] = cloneCharacter(character)
	return nil
}

func (w *World) CharacterByID(id CharacterID) (Character, bool) {
	character, ok := w.Characters[id]
	if !ok {
		return Character{}, false
	}
	return cloneCharacter(character), true
}

func (w *World) CarryItem(characterID CharacterID, item CraftedItem) error {
	itemValue := w.rules.ItemLegacyValue(item)
	if itemValue < 0 {
		return fmt.Errorf("item %q has negative legacy value", item.ID)
	}
	character, ok := w.Characters[characterID]
	if !ok {
		return fmt.Errorf("%w: %s", ErrCharacterNotFound, characterID)
	}
	if err := character.Carry(item); err != nil {
		return err
	}
	w.Characters[characterID] = character
	if characterID == w.PrimaryCharacter.ID {
		w.PrimaryCharacter = character
	}
	w.Crafters[item.CrafterID] += itemValue
	w.record(Event{
		Type:        EventItemCrafted,
		Description: fmt.Sprintf("%s carries %s, crafted by %s", character.Name, item.Name, item.CrafterID),
		SubjectID:   string(item.CrafterID),
		ObjectID:    string(item.ID),
		Subject:     NewTargetRef(string(item.CrafterID), TargetActor, ""),
		Object:      ItemRef(item),
		Data: map[string]string{
			"character_id": string(character.ID),
			"item_name":    item.Name,
		},
	})
	return nil
}

func (w *World) KillCharacterByID(characterID CharacterID, cause string, areaID AreaID) (LootDungeon, error) {
	character, ok := w.Characters[characterID]
	if !ok {
		return LootDungeon{}, fmt.Errorf("%w: %s", ErrCharacterNotFound, characterID)
	}
	if !character.Alive {
		return LootDungeon{}, fmt.Errorf("%s is already dead", character.Name)
	}
	if cause == "" {
		cause = "unknown danger"
	}

	dungeon, err := w.DepositDeathLoot(ActorID(character.ID), areaID, character.Inventory, cause)
	if err != nil {
		return LootDungeon{}, err
	}

	character.Alive = false
	character.Deaths++
	w.Characters[characterID] = character
	if characterID == w.PrimaryCharacter.ID {
		w.PrimaryCharacter = character
	}
	w.record(Event{
		Type:        EventCharacterDied,
		Description: fmt.Sprintf("%s died to %s", character.Name, cause),
		SubjectID:   string(character.ID),
		Subject:     CharacterRef(character),
		Data: map[string]string{
			"cause": cause,
		},
	})
	return dungeon, nil
}

func (w *World) RespawnCharacterByID(characterID CharacterID) error {
	character, ok := w.Characters[characterID]
	if !ok {
		return fmt.Errorf("%w: %s", ErrCharacterNotFound, characterID)
	}
	character.Respawn()
	w.Characters[characterID] = character
	if characterID == w.PrimaryCharacter.ID {
		w.PrimaryCharacter = character
	}
	w.record(Event{
		Type:        EventRespawned,
		Description: fmt.Sprintf("%s returns for another run", character.Name),
		SubjectID:   string(character.ID),
		Subject:     CharacterRef(character),
	})
	return nil
}

func (w *World) applyLootToCharacter(looter ActorID, loot []CraftedItem, legacyValue int) {
	characterID := CharacterID(looter)
	character, ok := w.Characters[characterID]
	if !ok {
		return
	}
	character.LegacyScore += legacyValue
	character.RecoveredItems = append(character.RecoveredItems, cloneItems(loot)...)
	w.Characters[characterID] = character
	if characterID == w.PrimaryCharacter.ID {
		w.PrimaryCharacter = character
	}
}
