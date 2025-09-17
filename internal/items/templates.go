package items

import (
	"fmt"
	"log"
	"math/rand"
)

// StatTemplate represents a template for generating item stats
type StatTemplate struct {
	Name           string
	RequiredStats  []StatEntry // Stats that are always present
	OptionalStats  []StatEntry // Stats that may be present with variance
	MaxOptional    int         // Maximum number of optional stats to add
}

// StatEntry represents a stat type and value range
type StatEntry struct {
	StatType    int
	BaseValue   int
	ValueRange  int // Random variance range (0 to ValueRange-1)
	Multiplier  float64 // Multiplier for base stat value
}

// StatTemplateManager handles stat template application
type StatTemplateManager struct {
	debug bool
}

// NewStatTemplateManager creates a new stat template manager
func NewStatTemplateManager(debug bool) *StatTemplateManager {
	return &StatTemplateManager{debug: debug}
}

// ApplySimpleStatTemplate applies a basic stat template to an item for phase 0 scaling
func (stm *StatTemplateManager) ApplySimpleStatTemplate(item *Item) {
	if stm.debug {
		log.Printf("Applying simple stat template for %s (Class: %d, Subclass: %d)", 
			item.Name, *item.Class, *item.Subclass)
	}

	// Clear existing stats
	stm.clearItemStats(item)

	// Get base stat value with variance
	baseStatValue := 6 + rand.Intn(5) // Random between 6-10 for variance

	// Apply template based on item type
	if *item.Class == 2 { // Weapons
		stm.applyWeaponTemplate(item, baseStatValue)
	} else if *item.Class == 4 { // Armor
		stm.applyArmorTemplate(item, baseStatValue)
	}

	// Update stats count
	stm.updateStatsCount(item)

	if stm.debug {
		statsCount, _ := item.GetField("StatsCount")
		log.Printf("Applied simple template: %d stats for %s", statsCount, item.Name)
	}
}

// clearItemStats clears all existing stats on an item
func (stm *StatTemplateManager) clearItemStats(item *Item) {
	for i := 1; i <= 10; i++ {
		item.UpdateField(fmt.Sprintf("StatType%d", i), 0)
		item.UpdateField(fmt.Sprintf("StatValue%d", i), 0)
	}
}

// applyWeaponTemplate applies weapon-specific stat templates
func (stm *StatTemplateManager) applyWeaponTemplate(item *Item, baseStatValue int) {
	classType := item.GetClassUserType()
	
	// Determine weapon type
	isPhysicalWeapon := stm.isPhysicalWeapon(*item.Subclass, classType)
	isTankWeapon := (*item.Subclass == 6) // Shield
	
	if stm.debug {
		log.Printf("Weapon %s (subclass %d, classType %d) determined as physical: %t", 
			item.Name, *item.Subclass, classType, isPhysicalWeapon)
	}

	if isPhysicalWeapon {
		stm.applyPhysicalWeaponTemplate(item, baseStatValue, isTankWeapon)
	} else {
		stm.applyCasterWeaponTemplate(item, baseStatValue, isTankWeapon)
	}
}

// isPhysicalWeapon determines if a weapon should use physical stats
func (stm *StatTemplateManager) isPhysicalWeapon(subclass, classType int) bool {
	// Physical weapon subclasses
	physicalSubclasses := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 13, 15, 16, 17, 18}
	for _, sc := range physicalSubclasses {
		if subclass == sc {
			// Special case: daggers can be physical or caster
			if subclass == 15 { // Dagger
				// If classType is generic (7), assume physical for daggers
				return classType == 7 || classType == 1 || classType == 2
			}
			return true
		}
	}
	
	// Caster weapon subclasses
	casterSubclasses := []int{10, 19, 20} // Staff, Wand, Fishing Pole
	for _, sc := range casterSubclasses {
		if subclass == sc {
			return false
		}
	}
	
	return false // Default to caster if unknown
}

// applyPhysicalWeaponTemplate applies physical DPS weapon stats
func (stm *StatTemplateManager) applyPhysicalWeaponTemplate(item *Item, baseStatValue int, isTankWeapon bool) {
	statSlot := 1
	
	// Primary stat (ALWAYS present) - choose based on weapon type
	primaryStat := 3 // Agility default
	
	// Weapon-specific stat assignment
	if *item.Subclass == 15 { // Dagger - ALWAYS Agility
		primaryStat = 3 // Agility
	} else if *item.Subclass == 0 || *item.Subclass == 1 || *item.Subclass == 8 { // Axe, Sword, Two-handed Sword
		// These can be Strength or Agility, with slight preference for Strength
		if rand.Intn(4) < 3 { // 75% chance for Strength
			primaryStat = 4 // Strength
		} else {
			primaryStat = 3 // Agility
		}
	} else if *item.Subclass == 2 || *item.Subclass == 3 || *item.Subclass == 18 { // Bow, Gun, Crossbow
		primaryStat = 3 // Agility (ranged weapons)
	} else {
		// Other weapons - random choice
		if rand.Intn(2) == 0 {
			primaryStat = 3 // Agility
		} else {
			primaryStat = 4 // Strength
		}
	}
	
	*item.StatType1 = primaryStat
	*item.StatValue1 = baseStatValue + 3 + rand.Intn(2) // 3-4 bonus, consistent
	statSlot++
	
	// Attack Power (ALWAYS present for physical weapons)
	*item.StatType2 = 38 // Attack Power
	*item.StatValue2 = baseStatValue*4 + rand.Intn(2) // Consistent high value with small variance
	statSlot++
	
	// Secondary stats (VARY these) - choose 1-3 randomly
	availableSecondaries := []int{32, 36, 31, 44} // Crit, Haste, Hit, Armor Pen
	stm.addRandomSecondaryStats(item, availableSecondaries, baseStatValue, &statSlot, 1+rand.Intn(3))
	
	// Only add stamina for tank weapons
	if isTankWeapon && statSlot <= 5 {
		item.UpdateField(fmt.Sprintf("StatType%d", statSlot), 7) // Stamina
		item.UpdateField(fmt.Sprintf("StatValue%d", statSlot), baseStatValue+8+rand.Intn(3)) // 8-10 bonus
	}
}

// applyCasterWeaponTemplate applies caster weapon stats
func (stm *StatTemplateManager) applyCasterWeaponTemplate(item *Item, baseStatValue int, isTankWeapon bool) {
	statSlot := 1
	
	// Intellect (ALWAYS present for casters)
	*item.StatType1 = 5 // Intellect
	*item.StatValue1 = baseStatValue + 3 + rand.Intn(2) // 3-4 bonus, consistent
	statSlot++
	
	// Spell Power (ALWAYS present for caster weapons)
	*item.StatType2 = 45 // Spell Power
	*item.StatValue2 = baseStatValue*3 + rand.Intn(2) // Consistent high value with small variance
	statSlot++
	
	// Secondary stats (VARY these) - choose 1-3 randomly
	availableSecondaries := []int{32, 36, 31, 43, 6} // Crit, Haste, Hit, MP5, Spirit
	stm.addRandomSecondaryStats(item, availableSecondaries, baseStatValue, &statSlot, 1+rand.Intn(3))
	
	// Only add stamina for tank weapons (rare for casters)
	if isTankWeapon && statSlot <= 5 {
		item.UpdateField(fmt.Sprintf("StatType%d", statSlot), 7) // Stamina
		item.UpdateField(fmt.Sprintf("StatValue%d", statSlot), baseStatValue+8+rand.Intn(3)) // 8-10 bonus
	}
}

// applyArmorTemplate applies armor-specific stat templates
func (stm *StatTemplateManager) applyArmorTemplate(item *Item, baseStatValue int) {
	if *item.Subclass == 0 { // Trinkets
		stm.applyTrinketTemplate(item, baseStatValue)
	} else {
		stm.applyGenericArmorTemplate(item, baseStatValue)
	}
}

// applyTrinketTemplate applies trinket stats (minimal stats, power from spells)
func (stm *StatTemplateManager) applyTrinketTemplate(item *Item, baseStatValue int) {
	*item.StatType1 = 7  // Stamina
	*item.StatValue1 = baseStatValue + 3 + rand.Intn(3) // 3-5 bonus with variance
}

// applyGenericArmorTemplate applies generic armor stats
func (stm *StatTemplateManager) applyGenericArmorTemplate(item *Item, baseStatValue int) {
	classType := item.GetClassUserType()
	statSlot := 1
	
	// Stamina (ALWAYS present on armor)
	*item.StatType1 = 7  // Stamina
	*item.StatValue1 = baseStatValue + 4 + rand.Intn(2) // 4-5 bonus, consistent
	statSlot++
	
	// Primary stat (ALWAYS present) based on class type
	primaryStat := 5 // Default to Intellect
	if classType == 1 { // Strength
		primaryStat = 4 // Strength
	} else if classType == 2 { // Agility
		primaryStat = 3 // Agility
	} else if classType == 7 { // Generic/Unknown - random choice
		primaryStats := []int{3, 4, 5} // Agi, Str, Int
		primaryStat = primaryStats[rand.Intn(len(primaryStats))]
	}
	// classType 3,4,5,6 (casters) get Intellect by default
	
	*item.StatType2 = primaryStat
	*item.StatValue2 = baseStatValue + 2 + rand.Intn(2) // 2-3 bonus, consistent
	statSlot++
	
	// Secondary stats (VARY these) - choose 0-2 randomly
	availableSecondaries := []int{32, 36, 31, 6} // Crit, Haste, Hit, Spirit
	stm.addRandomSecondaryStats(item, availableSecondaries, baseStatValue-3, &statSlot, rand.Intn(3))
}

// addRandomSecondaryStats adds random secondary stats to an item
func (stm *StatTemplateManager) addRandomSecondaryStats(item *Item, availableStats []int, baseValue int, statSlot *int, count int) {
	// Shuffle available stats
	rand.Shuffle(len(availableStats), func(i, j int) { 
		availableStats[i], availableStats[j] = availableStats[j], availableStats[i] 
	})
	
	// Add the specified number of secondary stats
	for i := 0; i < count && i < len(availableStats) && *statSlot <= 5; i++ {
		item.UpdateField(fmt.Sprintf("StatType%d", *statSlot), availableStats[i])
		item.UpdateField(fmt.Sprintf("StatValue%d", *statSlot), baseValue+rand.Intn(4)) // 0 to +3 variance
		(*statSlot)++
	}
}

// updateStatsCount updates the StatsCount field based on non-zero stats
func (stm *StatTemplateManager) updateStatsCount(item *Item) {
	statsCount := 0
	for i := 1; i <= 10; i++ {
		statType, _ := item.GetField(fmt.Sprintf("StatType%d", i))
		if statType > 0 {
			statsCount++
		}
	}
	*item.StatsCount = statsCount
}
