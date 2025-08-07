package main

import (
	"log"
	"math"

	"github.com/araxiaonline/endgame-item-generator/internal/db/mysql"
	"github.com/araxiaonline/endgame-item-generator/internal/items"
	"github.com/araxiaonline/endgame-item-generator/internal/spells"
)

// ComparePowerStats compares Attack Power and Spell Power stats between old generated items and new items
// and logs the differences. This helps identify cases where new items might have lower power stats
// despite having equal or higher item levels. Also checks for spells that add these stats.
// If oldGenItem is empty (Entry = 0), it will only show stats for the new item.
// Optionally filters by itemClass and itemSubClass (pass -1 to disable filtering)
func ComparePowerStats(newItem mysql.DbItem, oldGenItem mysql.DbItem, itemName string, itemClass int, itemSubClass ...int) {
	// No need to check for nil since we're using value types

	// Check if we should filter by item class/subclass
	if itemClass >= 0 {
		// Skip if item class doesn't match
		if newItem.Class == nil || *newItem.Class != itemClass {
			return
		}

		// Check subclass if provided
		if len(itemSubClass) > 0 && itemSubClass[0] >= 0 {
			if newItem.Subclass == nil || *newItem.Subclass != itemSubClass[0] {
				return
			}
		}
	}

	newItemLevel := 0
	oldItemLevel := 0
	entryID := 0
	originalEntry := 0

	if newItem.ItemLevel != nil {
		newItemLevel = *newItem.ItemLevel
	}

	if oldGenItem.ItemLevel != nil {
		oldItemLevel = *oldGenItem.ItemLevel
	}

	entryID = newItem.Entry

	// Check if we have a valid old item to compare with
	isComparing := oldGenItem.Entry != 0

	// Calculate the original entry ID by subtracting the bump (20000000) from old generated item
	if isComparing {
		originalEntry = oldGenItem.Entry - 20000000
	} else {
		// If we don't have an old item, try to use the new item's entry as the original
		originalEntry = newItem.Entry
	}

	// Try to get the original item to get its original level
	originalItemLevel := 0
	db, err := mysql.GetDb()
	if err == nil {
		originalItem, err := db.GetItem(originalEntry)
		if err == nil && originalItem.ItemLevel != nil {
			originalItemLevel = *originalItem.ItemLevel
			log.Printf("  Original item found: Entry %d, Level %d", originalEntry, originalItemLevel)
		}
	}

	// If we couldn't get the original item level, use the old generated item level
	if originalItemLevel == 0 {
		originalItemLevel = oldItemLevel
	}

	// Find Attack Power and Spell Power in the new item
	newAttackPower := findStatValue(&newItem, items.STAT.AttackPower)
	newSpellPower := findStatValue(&newItem, items.STAT.SpellPower)

	// Find Attack Power and Spell Power from spells in the new item
	newSpellAttackPower, newSpellSpellPower := findPowerFromSpells(&newItem, originalItemLevel)

	// Add spell-based power to direct stats for the new item
	newAttackPower += newSpellAttackPower
	newSpellPower += newSpellSpellPower

	if isComparing {
		// If we have an old item to compare with, get its stats too
		oldAttackPower := findStatValue(&oldGenItem, items.STAT.AttackPower)
		oldSpellPower := findStatValue(&oldGenItem, items.STAT.SpellPower)

		// Find Attack Power and Spell Power from spells in the old item
		oldSpellAttackPower, oldSpellSpellPower := findPowerFromSpells(&oldGenItem)

		// Add spell-based power to direct stats for the old item
		oldAttackPower += oldSpellAttackPower
		oldSpellPower += oldSpellSpellPower

		// Log Attack Power comparison if either item has it
		if oldAttackPower > 0 || newAttackPower > 0 {
			log.Printf("===== ATTACK POWER COMPARISON: %s (Entry: %d) =====", itemName, entryID)
			log.Printf("  Item Level: %d -> %d", oldItemLevel, newItemLevel)
			log.Printf("  Attack Power: %d -> %d", oldAttackPower, newAttackPower)

			// Log spell-based Attack Power if present
			if oldSpellAttackPower > 0 || newSpellAttackPower > 0 {
				log.Printf("  Attack Power from Spells: %d -> %d", oldSpellAttackPower, newSpellAttackPower)
			}

			// Highlight if new item has lower Attack Power despite equal or higher item level
			if newItemLevel >= oldItemLevel && newAttackPower < oldAttackPower && newAttackPower > 0 {
				log.Printf("  WARNING: Attack Power DECREASED despite item level being equal or higher!")
			}
			log.Printf("==========================================")
		}

		// Log Spell Power comparison if either item has it
		if oldSpellPower > 0 || newSpellPower > 0 {
			log.Printf("===== SPELL POWER COMPARISON: %s (Entry: %d) =====", itemName, entryID)
			log.Printf("  Item Level: %d -> %d", oldItemLevel, newItemLevel)
			log.Printf("  Spell Power: %d -> %d", oldSpellPower, newSpellPower)

			// Log spell-based Spell Power if present
			if oldSpellSpellPower > 0 || newSpellSpellPower > 0 {
				log.Printf("  Spell Power from Spells: %d -> %d", oldSpellSpellPower, newSpellSpellPower)
			}

			// Highlight if new item has lower Spell Power despite equal or higher item level
			if newItemLevel >= oldItemLevel && newSpellPower < oldSpellPower && newSpellPower > 0 {
				log.Printf("  WARNING: Spell Power DECREASED despite item level being equal or higher!")
			}
			log.Printf("==========================================")
		}
	} else {
		// If we don't have an old item to compare with, just show the new item's stats

		// Log Attack Power if present
		if newAttackPower > 0 {
			log.Printf("===== NEW ITEM ATTACK POWER: %s (Entry: %d) =====", itemName, entryID)
			log.Printf("  Item Level: %d", newItemLevel)
			log.Printf("  Attack Power: %d", newAttackPower)

			// Log spell-based Attack Power if present
			if newSpellAttackPower > 0 {
				log.Printf("  Attack Power from Spells: %d", newSpellAttackPower)
				log.Printf("  Original Item Level: %d, New Item Level: %d", originalItemLevel, newItemLevel)
			}
			log.Printf("==========================================")
		}

		// Log Spell Power if present
		if newSpellPower > 0 {
			log.Printf("===== NEW ITEM SPELL POWER: %s (Entry: %d) =====", itemName, entryID)
			log.Printf("  Item Level: %d", newItemLevel)
			log.Printf("  Spell Power: %d", newSpellPower)

			// Log spell-based Spell Power if present
			if newSpellSpellPower > 0 {
				log.Printf("  Spell Power from Spells: %d", newSpellSpellPower)
				log.Printf("  Original Item Level: %d, New Item Level: %d", originalItemLevel, newItemLevel)
			}
			log.Printf("==========================================")
		}
	}
}

// findStatValue searches through all stat slots in an item to find a specific stat type
// and returns its value. Returns 0 if the stat is not found.
func findStatValue(item *mysql.DbItem, statType int) int {
	if item == nil {
		return 0
	}

	// Check all stat slots
	if item.StatType1 != nil && *item.StatType1 == statType && item.StatValue1 != nil {
		return *item.StatValue1
	}
	if item.StatType2 != nil && *item.StatType2 == statType && item.StatValue2 != nil {
		return *item.StatValue2
	}
	if item.StatType3 != nil && *item.StatType3 == statType && item.StatValue3 != nil {
		return *item.StatValue3
	}
	if item.StatType4 != nil && *item.StatType4 == statType && item.StatValue4 != nil {
		return *item.StatValue4
	}
	if item.StatType5 != nil && *item.StatType5 == statType && item.StatValue5 != nil {
		return *item.StatValue5
	}
	if item.StatType6 != nil && *item.StatType6 == statType && item.StatValue6 != nil {
		return *item.StatValue6
	}
	if item.StatType7 != nil && *item.StatType7 == statType && item.StatValue7 != nil {
		return *item.StatValue7
	}
	if item.StatType8 != nil && *item.StatType8 == statType && item.StatValue8 != nil {
		return *item.StatValue8
	}
	if item.StatType9 != nil && *item.StatType9 == statType && item.StatValue9 != nil {
		return *item.StatValue9
	}
	if item.StatType10 != nil && *item.StatType10 == statType && item.StatValue10 != nil {
		return *item.StatValue10
	}

	return 0
}

// findPowerFromSpells checks for spells on an item that add Attack Power or Spell Power
// and returns the total values for each. If originalItemLevel is provided, it will be used
// to scale the spell values to match what they would be after actual scaling.
func findPowerFromSpells(item *mysql.DbItem, originalItemLevel ...int) (attackPower int, spellPower int) {
	if item == nil {
		return 0, 0
	}

	// Set the spell IDs from the DbItem
	for i := 1; i <= 3; i++ {
		var spellID *int

		switch i {
		case 1:
			spellID = item.SpellId1
		case 2:
			spellID = item.SpellId2
		case 3:
			spellID = item.SpellId3
		}

		if spellID == nil || *spellID == 0 {
			continue
		}

		// Get the spell from the database
		db, err := mysql.GetDb()
		if err != nil {
			log.Printf("Failed to get database connection: %v", err)
			continue
		}

		dbSpell, err := db.GetSpell(*spellID)
		if err != nil {
			log.Printf("Failed to get spell %d: %v", *spellID, err)
			continue
		}

		// Create a Spell struct
		spell := spells.Spell{
			DbSpell:       dbSpell,
			ItemSpellSlot: i,
		}

		// Scale the spell if we have original item level and current item level
		scaledSpell := spell
		if len(originalItemLevel) > 0 && originalItemLevel[0] > 0 && item.ItemLevel != nil && item.Quality != nil {
			// Create a copy of the spell to scale
			scaledSpell = spell

			// Use the new ForceScaleSpell method to scale the spell regardless of effect type
			// This ensures all spells get scaled, not just the ones that qualify for ScaleSpell
			// No tier specified - use base scaling only
			err := scaledSpell.ForceScaleSpell(originalItemLevel[0], *item.ItemLevel, *item.Quality, 3)
			if err == nil {
				log.Printf("  Scaled spell %s (ID: %d) from level %d to %d using ForceScaleSpell",
					scaledSpell.Name, scaledSpell.ID, originalItemLevel[0], *item.ItemLevel)

				// Show before/after values for effect points
				if spell.EffectBasePoints1 != 0 {
					log.Printf("  Effect1: Original value: %d, Scaled value: %d", spell.EffectBasePoints1, scaledSpell.EffectBasePoints1)
				}
				if spell.EffectBasePoints2 != 0 {
					log.Printf("  Effect2: Original value: %d, Scaled value: %d", spell.EffectBasePoints2, scaledSpell.EffectBasePoints2)
				}
				if spell.EffectBasePoints3 != 0 {
					log.Printf("  Effect3: Original value: %d, Scaled value: %d", spell.EffectBasePoints3, scaledSpell.EffectBasePoints3)
				}
			} else {
				log.Printf("  Error scaling spell %s (ID: %d): %v", spell.Name, spell.ID, err)
			}
		}

		// Convert spell to stats
		convStats, err := scaledSpell.ConvertToStats()
		if err != nil {
			log.Printf("  Error converting spell to stats: %v", err)
			continue
		}

		// Check for Attack Power or Spell Power in the converted stats
		for _, stat := range convStats {
			if stat.StatType == items.STAT.AttackPower {
				attackPower += stat.StatValue
				if len(originalItemLevel) > 0 && originalItemLevel[0] > 0 {
					logSpellDetails(spell, "Attack Power", stat.StatValue, item, originalItemLevel[0])
				} else {
					logSpellDetails(spell, "Attack Power", stat.StatValue, item)
				}
			} else if stat.StatType == items.STAT.SpellPower {
				spellPower += stat.StatValue
				if len(originalItemLevel) > 0 && originalItemLevel[0] > 0 {
					logSpellDetails(spell, "Spell Power", stat.StatValue, item, originalItemLevel[0])
				} else {
					logSpellDetails(spell, "Spell Power", stat.StatValue, item)
				}
			}
		}
	}

	return attackPower, spellPower
}

// logSpellDetails logs detailed information about a spell that adds power stats
func logSpellDetails(spell spells.Spell, powerType string, powerValue int, item *mysql.DbItem, originalItemLevel ...int) {
	// Log basic spell information
	log.Printf("  Found %s spell: %s (ID: %d) adding %d %s Description: %s", powerType, spell.Name, spell.ID, powerValue, powerType, spell.Description)

	// Log aura description if available
	if spell.AuraDescription != "" {
		log.Printf("    Aura: %s", spell.AuraDescription)
	}

	// Show what the spell would look like if scaled
	if item.ItemLevel != nil && item.Quality != nil {
		// Calculate what this spell would look like if scaled to a higher item level
		currentLevel := *item.ItemLevel
		startingLevel := currentLevel

		// If original item level is provided, use it as the starting point for scaling
		if len(originalItemLevel) > 0 && originalItemLevel[0] > 0 {
			startingLevel = originalItemLevel[0]
			log.Printf("    Using original item level %d as scaling base (current level: %d)", startingLevel, currentLevel)
		}

		qualModifier := 1.0
		switch *item.Quality {
		case 3: // Rare
			qualModifier = 1.20
		case 4: // Epic
			qualModifier = 1.30
		case 5: // Legendary
			qualModifier = 1.40
		}

		// Determine the appropriate multiplier based on spell effect type
		effectMultiplier := 1.0

		// Check for stat-boosting effects (Attack Power and Spell Power)
		if powerType == "Attack Power" || powerType == "Spell Power" {
			// For stat buffs like Attack Power and Spell Power, use a higher multiplier
			// to account for the significant item level jumps
			effectMultiplier = 2.0
		}

		// Show scaling examples for larger item level jumps
		for _, newLevel := range []int{100, 150, 200, 250} {
			if newLevel <= currentLevel {
				continue // Skip levels that are lower than current
			}

			// Calculate level ratio with a power curve to account for exponential scaling
			// Use the original item level as the starting point if available
			levelRatio := math.Pow(float64(newLevel)/float64(startingLevel), 1.3)

			// Simulate the spell scaling formula with enhanced scaling for large level jumps
			scaledValue := int(float64(powerValue) * levelRatio * qualModifier * effectMultiplier)

			log.Printf("    If scaled to item level %d: ~%d %s (%.2fx increase)",
				newLevel, scaledValue, powerType, float64(scaledValue)/float64(powerValue))
		}
	}
}
