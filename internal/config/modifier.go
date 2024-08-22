package config

var InvTypeModifiers = map[int]float64{
	1:  0.813, // Head
	2:  1.0,   // Neck
	3:  0.75,  // Shoulder
	5:  1.0,   // Chest
	6:  0.562, // Waist
	7:  0.875, // Legs
	8:  0.688, // Feet
	9:  0.437, // Wrists
	10: 0.625, // Hands
	11: 1.0,   // Finger
	13: 0.42,  // One-Hand (not to confuse with Off-Hand = 22)
	14: 0.56,  // Shield (class = armor, not weapon even if in weapon slot)
	15: 0.32,  // Ranged (Bows) (see also Ranged right = 26)
	16: 0.56,  // Back
	17: 1.0,   // Two-Hand
	18: 1.0,   // Bag (assuming same as Chest for simplicity)
	19: 1.0,   // Tabard (assuming same as Chest for simplicity)
	20: 1.0,   // Robe (see also Chest = 5)
	21: 1.0,   // Main hand
	22: 0.42,  // Off Hand weapons (see also One-Hand = 13)
	23: 0.56,  // Held in Off-Hand (class = armor, not weapon even if in weapon slot)
	24: 1.0,   // Ammo (assuming same as Chest for simplicity)
	25: 0.32,  // Thrown
	26: 0.32,  // Ranged right (Wands, Guns) (see also Ranged = 15)
	27: 1.0,   // Quiver (assuming same as Chest for simplicity)
}

var QualityModifiers = map[int]float64{
	0: 1.0, // Common
	1: 1.1, // Uncommon
	2: 1.2, // Rare
	3: 1.3, // Epic
	4: 1.5, // Legendary
	5: 1.7, // Artifact
}

var MaterialModifiers = map[int]float64{
	1: 1.2,  // Cloth
	2: 2.2,  // Leather
	3: 4.75, // Mail
	4: 9.0,  // Plate
	6: 20.0, // Plate
}

var StatModifiers = map[int]float64{
	0:  1.0,  // ITEM_MOD_MANA
	1:  1.0,  // ITEM_MOD_HEALTH
	3:  1.0,  // ITEM_MOD_AGILITY
	4:  1.0,  // ITEM_MOD_STRENGTH
	5:  1.0,  // ITEM_MOD_INTELLECT
	6:  1.0,  // ITEM_MOD_SPIRIT
	7:  1.0,  // ITEM_MOD_STAMINA
	12: 1.0,  // ITEM_MOD_DEFENSE_SKILL_RATING
	13: 1.0,  // ITEM_MOD_DODGE_RATING
	14: 1.0,  // ITEM_MOD_PARRY_RATING
	15: 1.0,  // ITEM_MOD_BLOCK_RATING
	16: 1.0,  // ITEM_MOD_HIT_MELEE_RATING
	17: 1.0,  // ITEM_MOD_HIT_RANGED_RATING
	18: 1.0,  // ITEM_MOD_HIT_SPELL_RATING
	19: 1.0,  // ITEM_MOD_CRIT_MELEE_RATING
	20: 1.0,  // ITEM_MOD_CRIT_RANGED_RATING
	21: 1.0,  // ITEM_MOD_CRIT_SPELL_RATING
	22: 1.0,  // ITEM_MOD_HIT_TAKEN_MELEE_RATING
	23: 1.0,  // ITEM_MOD_HIT_TAKEN_RANGED_RATING
	24: 1.0,  // ITEM_MOD_HIT_TAKEN_SPELL_RATING
	25: 1.0,  // ITEM_MOD_CRIT_TAKEN_MELEE_RATING
	26: 1.0,  // ITEM_MOD_CRIT_TAKEN_RANGED_RATING
	27: 1.0,  // ITEM_MOD_CRIT_TAKEN_SPELL_RATING
	28: 1.0,  // ITEM_MOD_HASTE_MELEE_RATING
	29: 1.0,  // ITEM_MOD_HASTE_RANGED_RATING
	30: 1.0,  // ITEM_MOD_HASTE_SPELL_RATING
	31: 1.0,  // ITEM_MOD_HIT_RATING
	32: 1.0,  // ITEM_MOD_CRIT_RATING
	33: 1.0,  // ITEM_MOD_HIT_TAKEN_RATING
	34: 1.0,  // ITEM_MOD_CRIT_TAKEN_RATING
	35: 1.0,  // ITEM_MOD_RESILIENCE_RATING
	36: 1.0,  // ITEM_MOD_HASTE_RATING
	37: 1.0,  // ITEM_MOD_EXPERTISE_RATING
	38: 0.5,  // ITEM_MOD_ATTACK_POWER
	39: 0.5,  // ITEM_MOD_RANGED_ATTACK_POWER
	40: 0.5,  // ITEM_MOD_FERAL_ATTACK_POWER (not used as of 3.3)
	41: 0.5,  // ITEM_MOD_SPELL_HEALING_DONE
	42: 0.5,  // ITEM_MOD_SPELL_DAMAGE_DONE
	43: 2.5,  // ITEM_MOD_MANA_REGENERATION
	44: 1.0,  // ITEM_MOD_ARMOR_PENETRATION_RATING
	45: 0.5,  // ITEM_MOD_SPELL_POWER
	46: 1.0,  // ITEM_MOD_HEALTH_REGEN
	47: 2.0,  // ITEM_MOD_SPELL_PENETRATION
	48: 0.65, // ITEM_MOD_BLOCK_VALUE
}
