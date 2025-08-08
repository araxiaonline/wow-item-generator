package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/araxiaonline/endgame-item-generator/internal/config"
	"github.com/araxiaonline/endgame-item-generator/internal/db/mysql"
	"github.com/araxiaonline/endgame-item-generator/internal/items"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

// Molten Core Configuration
const (
	MOLTEN_CORE_MAP_ID     = 409
	MOLTEN_CORE_ITEM_LEVEL = 325
	MOLTEN_CORE_QUALITY    = 4 // Epic
	MOLTEN_CORE_PHASE      = 1
	MOLTEN_CORE_DIFFICULTY = 3 // Mythic
)

// Trinket spells organized by class type
var spellTrinketSpells = []int{
	60493, // increase spell power
	60485,
	49622,
	33953,
	60063,
}

var meleeTrinketSpells = []int{
	67672,
	58901,
	60436,
	60313,
	60487,
	60442,
}

var tankTrinketSpells = []int{
	67653,
	60180,
	60258,
}

// Fire-themed item names for Molten Core fire resistance
var fireThemeNames = []string{
	"flame", "fire", "salamander", "crimson", "burning",
	"blazing", "infernal", "molten", "ember", "igniting",
	"flamewalker", "flameguard",
}

// Stat priority mappings based on class types
type StatPriority struct {
	Primary   []int // Must have stats
	Secondary []int // Important stats
	Tertiary  []int // Optional stats
}

// StatCountRange defines expected stat counts for item types
type StatCountRange struct {
	Min     int // Minimum required stats
	Max     int // Maximum allowed stats
	Optimal int // Optimal stat count
}

var classStatPriorities = map[int]StatPriority{
	1: { // Strength Melee (Warriors, DKs, Paladins)
		Primary:   []int{4, 7},           // Strength, Stamina
		Secondary: []int{31, 37, 32, 36}, // Hit, Expertise, Crit, Haste
		Tertiary:  []int{44},             // Armor Penetration
	},
	2: { // Agility Melee (Rogues, Hunters, Feral Druids)
		Primary:   []int{3, 7},           // Agility, Stamina
		Secondary: []int{31, 32, 38, 36}, // Hit, Crit, Attack Power, Haste
		Tertiary:  []int{44, 37},         // Armor Penetration, Expertise
	},
	3: { // Ranged (Hunters)
		Primary:   []int{3, 38},          // Agility, Attack Power
		Secondary: []int{31, 32, 44, 36}, // Hit, Crit, Armor Pen, Haste
		Tertiary:  []int{7},              // Stamina
	},
	4: { // Mage (Spell DPS)
		Primary:   []int{45, 5},      // Spell Power, Intellect
		Secondary: []int{31, 32, 36}, // Hit, Crit, Haste
		Tertiary:  []int{6, 43},      // Spirit, Mana Regen
	},
	5: { // Healer
		Primary:   []int{45, 5},      // Spell Power, Intellect
		Secondary: []int{43, 32, 36}, // Crit, Haste, Mana Regen
		Tertiary:  []int{6, 7},       // Spirit, Stamina
	},
	6: { // Tank
		Primary:   []int{7, 12},          // Stamina, Defense
		Secondary: []int{13, 14, 31, 37}, // Dodge, Parry, Hit, Expertise
		Tertiary:  []int{4, 3},           // Strength, Agility
	},
}

// ItemGenerationResult holds the result of item generation
type ItemGenerationResult struct {
	Item          *items.Item
	ReferenceItem *items.Item   // High-level reference item used for scaling
	SpellInfo     *SpellDetails // Information about assigned spells
	Success       bool
	Errors        []string
	Warnings      []string
	Validated     bool
}

// SpellDetails contains information about assigned spells
type SpellDetails struct {
	SpellID     int
	SpellName   string
	Description string
}

// MoltenCoreGenerator handles Molten Core item generation
type MoltenCoreGenerator struct {
	db        *mysql.MySqlDb
	debug     bool
	itemLevel int
	quality   int
}

// isTrinket checks if an item is a trinket
func isTrinket(item *items.Item) bool {
	return item.InventoryType != nil && *item.InventoryType == 0
}

// clearItemSpells removes all spells from an item
func clearItemSpells(item *items.Item) {
	zero := 0
	item.SpellId1 = &zero
	item.SpellTrigger1 = &zero

	item.SpellId2 = &zero
	item.SpellTrigger2 = &zero

	item.SpellId3 = &zero
	item.SpellTrigger3 = &zero
}

// copySpellsFromReference copies all spell data from reference item to target item
func copySpellsFromReference(targetItem, referenceItem *items.Item) {
	// Clear existing spells first
	clearItemSpells(targetItem)

	// Copy spell 1
	if referenceItem.SpellId1 != nil && *referenceItem.SpellId1 > 0 {
		targetItem.SpellId1 = referenceItem.SpellId1
		targetItem.SpellTrigger1 = referenceItem.SpellTrigger1
	}

	// Copy spell 2
	if referenceItem.SpellId2 != nil && *referenceItem.SpellId2 > 0 {
		targetItem.SpellId2 = referenceItem.SpellId2
		targetItem.SpellTrigger2 = referenceItem.SpellTrigger2
	}

	// Copy spell 3
	if referenceItem.SpellId3 != nil && *referenceItem.SpellId3 > 0 {
		targetItem.SpellId3 = referenceItem.SpellId3
		targetItem.SpellTrigger3 = referenceItem.SpellTrigger3
	}
}

// getSimilarWeaponSubclasses returns alternative weapon subclasses to try when no compatible items are found
func getSimilarWeaponSubclasses(originalSubclass int) []int {
	// Weapon subclass mappings for fallback searches
	weaponGroups := map[int][]int{
		// Two-handed weapons
		1:  {1, 5, 8},   // 2H Axe -> 2H Axe, 2H Mace, 2H Sword
		5:  {1, 5, 8},   // 2H Mace -> 2H Axe, 2H Mace, 2H Sword
		8:  {1, 5, 8},   // 2H Sword -> 2H Axe, 2H Mace, 2H Sword
		6:  {6, 17},     // Polearm -> Polearm, Spear
		17: {6, 17},     // Spear -> Polearm, Spear
		9:  {9, 10},     // Staff -> Staff, Stave
		10: {9, 10},     // Stave -> Staff, Stave
		
		// One-handed weapons
		0:  {0, 4, 7, 15}, // 1H Axe -> 1H Axe, 1H Mace, 1H Sword, Fist
		4:  {0, 4, 7, 15}, // 1H Mace -> 1H Axe, 1H Mace, 1H Sword, Fist
		7:  {0, 4, 7, 15}, // 1H Sword -> 1H Axe, 1H Mace, 1H Sword, Fist
		15: {0, 4, 7, 15}, // Fist -> 1H Axe, 1H Mace, 1H Sword, Fist
		13: {13},          // Dagger -> Dagger (unique)
		
		// Ranged weapons
		2:  {2, 3, 18}, // Bow -> Bow, Gun, Crossbow
		3:  {2, 3, 18}, // Gun -> Bow, Gun, Crossbow
		18: {2, 3, 18}, // Crossbow -> Bow, Gun, Crossbow
		16: {16},       // Thrown -> Thrown (unique)
		19: {19},       // Wand -> Wand (unique)
		20: {20},       // Fishing Pole -> Fishing Pole (unique)
	}
	
	if alternatives, exists := weaponGroups[originalSubclass]; exists {
		// Return alternatives excluding the original subclass
		var result []int
		for _, alt := range alternatives {
			if alt != originalSubclass {
				result = append(result, alt)
			}
		}
		return result
	}
	
	return []int{} // No alternatives found
}

// addMissingKeyStats automatically adds missing key stats (SPELL_POWER/ATTACK_POWER) when validation fails
func (g *MoltenCoreGenerator) addMissingKeyStats(item *items.Item, classType int) bool {
	var addedStats []string
	
	// Determine what key stat should be added based on class type
	var targetStatType int
	var statName string
	
	switch classType {
	case 4, 5: // Mage, Healer - need SPELL_POWER
		targetStatType = 45 // SPELL_POWER
		statName = "SPELL_POWER"
	case 1, 2, 3, 6: // Melee DPS, Ranged, Tank - need ATTACK_POWER
		targetStatType = 38 // ATTACK_POWER
		statName = "ATTACK_POWER"
	default:
		return false // Unknown class type, can't determine what stat to add
	}
	
	// Check if the item already has this key stat
	for i := 1; i <= 8; i++ {
		statTypeField := fmt.Sprintf("StatType%d", i)
		if statType, err := item.GetField(statTypeField); err == nil && statType == targetStatType {
			return false // Already has the key stat
		}
	}
	
	// Find an empty stat slot to add the missing key stat
	for i := 1; i <= 8; i++ {
		statTypeField := fmt.Sprintf("StatType%d", i)
		statValueField := fmt.Sprintf("StatValue%d", i)
		
		if statType, err := item.GetField(statTypeField); err == nil && statType == 0 {
			// Found empty slot, calculate appropriate stat value
			baseValue := rand.Intn(151) + 350 // Random between 350-500
			
			// Apply scaling based on item level and quality
			scaledValue := g.calculateScaledStatValue(baseValue, targetStatType)
			
			// Set the stat
			item.UpdateField(statTypeField, targetStatType)
			item.UpdateField(statValueField, scaledValue)
			
			addedStats = append(addedStats, fmt.Sprintf("%s: %d", statName, scaledValue))
			
			if g.debug {
				log.Printf("Auto-added missing key stat %s (%d) = %d to %s", statName, targetStatType, scaledValue, item.Name)
			}
			
			return true
		}
	}
	
	return false // No empty slots available
}

// calculateScaledStatValue calculates an appropriate stat value based on item level, quality, and stat type
func (g *MoltenCoreGenerator) calculateScaledStatValue(baseValue, statType int) int {
	// Get scaling factor for this stat type
	scalingFactor := 1.0
	if factor, exists := config.ScalingFactor[statType]; exists {
		scalingFactor = factor
	}
	
	// Apply item level and quality scaling
	itemLevelModifier := float64(g.itemLevel) / 100.0
	qualityModifier := 1.0
	if modifier, exists := config.QualityModifiers[g.quality]; exists {
		qualityModifier = modifier
	}
	
	// Calculate final value
	finalValue := float64(baseValue) * scalingFactor * itemLevelModifier * qualityModifier
	
	return int(finalValue)
}

// getSimilarArmorSubclasses returns alternative armor subclasses for fallback searches
func getSimilarArmorSubclasses(originalSubclass int) []int {
	// For armor, we can be more flexible with material types
	armorGroups := map[int][]int{
		1: {1},    // Cloth -> Cloth only
		2: {2, 3}, // Leather -> Leather, Mail
		3: {2, 3}, // Mail -> Leather, Mail
		4: {4},    // Plate -> Plate only
		6: {6},    // Shield -> Shield only
		0: {0},    // Miscellaneous -> Miscellaneous only
	}
	
	if alternatives, exists := armorGroups[originalSubclass]; exists {
		// Return alternatives excluding the original subclass
		var result []int
		for _, alt := range alternatives {
			if alt != originalSubclass {
				result = append(result, alt)
			}
		}
		return result
	}
	
	return []int{} // No alternatives found
}

// getSpellDetails looks up spell information from the database
func (g *MoltenCoreGenerator) getSpellDetails(spellID int) *SpellDetails {
	if spellID == 0 {
		return nil
	}

	spell, err := g.db.GetSpell(spellID)
	if err != nil {
		log.Printf("Failed to lookup spell %d: %v", spellID, err)
		return nil
	}

	return &SpellDetails{
		SpellID:     spell.ID,
		SpellName:   spell.Name,
		Description: spell.Description,
	}
}

func NewMoltenCoreGenerator(db *mysql.MySqlDb, debug bool) *MoltenCoreGenerator {
	return &MoltenCoreGenerator{
		db:        db,
		debug:     debug,
		itemLevel: MOLTEN_CORE_ITEM_LEVEL,
		quality:   MOLTEN_CORE_QUALITY,
	}
}

func getClassString(class int) string {
	switch class {
	case 1:
		return "Strength Melee"
	case 2:
		return "Agility Melee"
	case 3:
		return "Ranged"
	case 4:
		return "Mage"
	case 5:
		return "Healer"
	case 6:
		return "Tank"
	default:
		return "Unknown"
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	godotenv.Load("../../.env")

	debug := flag.Bool("debug", false, "Enable verbose logging inside generator")
	outputSql := flag.Bool("sql", false, "Output SQL statements for generated items")
	validateOnly := flag.Bool("validate", false, "Only validate items without generating")
	flag.Parse()

	if *debug {
		log.SetOutput(os.Stdout)
	} else {
		log.SetOutput(io.Discard)
	}

	// Connect to MySQL
	mysqlDb, err := mysql.Connect(&mysql.MySqlConfig{
		Host:     os.Getenv("DB_HOST"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Database: os.Getenv("DB_NAME"),
	})

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Initialize Molten Core generator
	generator := NewMoltenCoreGenerator(mysqlDb, *debug)

	fmt.Printf("🔥 Molten Core Item Generator v2.0\n")
	fmt.Printf("Target Item Level: %d, Quality: %d (Epic)\n\n", MOLTEN_CORE_ITEM_LEVEL, MOLTEN_CORE_QUALITY)

	// Get Molten Core items from database
	// Boss entries and GameObject entries for Molten Core
	bossEntries := []int{11502}        // Previously hardcoded boss entry
	gameObjectEntries := []int{179703} // Previously hardcoded GameObject entry
	rareItems, err := mysqlDb.GetBossMapItems(MOLTEN_CORE_MAP_ID, bossEntries, gameObjectEntries, 0, 0)
	if err != nil {
		log.Fatal("Failed to get Molten Core items:", err)
	}

	fmt.Printf("📋 Processing %d Molten Core items...\n\n", len(rareItems))

	var results []ItemGenerationResult
	successCount := 0

	// Process each item
	for i, dbItem := range rareItems {
		fmt.Printf("[%d/%d] Processing: %s (Entry: %d)\n", i+1, len(rareItems), dbItem.Name, dbItem.Entry)

		// Store original item for comparison (before any modifications)
		// Create a deep copy to preserve original values
		originalItem := items.ItemFromDbItem(dbItem)
		// Scale original item to show actual baseline stats at original item level
		if originalItem.ItemLevel != nil && *originalItem.ItemLevel > 0 {
			originalItem.ScaleItem(*originalItem.ItemLevel, *originalItem.Quality)
		}

		// Store true original values before any scaling for comparison
		var originalArmor int
		var originalFireRes int
		var originalItemLevel int

		if dbItem.Armor != nil {
			originalArmor = *dbItem.Armor
		}
		if dbItem.FireRes != nil {
			originalFireRes = *dbItem.FireRes
		}
		if dbItem.ItemLevel != nil {
			originalItemLevel = *dbItem.ItemLevel
		}

		result := generator.GenerateItem(dbItem, *validateOnly)
		results = append(results, result)

		if result.Success {
			successCount++
			fmt.Printf("✅ Successfully generated %s\n", result.Item.Name)

			// Show clean before/after comparison
			if !*validateOnly {
				classType := result.Item.GetClassUserType()
				printThreeWayComparison(&originalItem, result.ReferenceItem, result.Item, result.SpellInfo, classType, originalArmor, originalFireRes, originalItemLevel)
			}

			if *outputSql && result.Item != nil {
				sqlStatement := items.ItemToSql(*result.Item, 80, MOLTEN_CORE_DIFFICULTY)
				fmt.Printf("SQL: %s\n", sqlStatement)
			}
		} else {
			fmt.Printf("❌ Failed to generate %s\n", dbItem.Name)
			for _, errMsg := range result.Errors {
				fmt.Printf("   Error: %s\n", errMsg)
			}
		}

		for _, warning := range result.Warnings {
			fmt.Printf("⚠️  Warning: %s\n", warning)
		}
		fmt.Println()
	}

	// Print summary
	fmt.Printf("\n🏆 Generation Summary:\n")
	fmt.Printf("Total Items: %d\n", len(rareItems))
	fmt.Printf("Successful: %d\n", successCount)
	fmt.Printf("Failed: %d\n", len(rareItems)-successCount)
	fmt.Printf("Success Rate: %.1f%%\n", float64(successCount)/float64(len(rareItems))*100)
}

// addFireResistanceIfNeeded scales existing fire resistance by 1.5x and adds fire resistance to fire-themed items
func addFireResistanceIfNeeded(item *items.Item) {
	if item.Name == "" {
		return
	}

	// First, scale existing fire resistance by 1.5x if present
	if item.FireRes != nil && *item.FireRes > 0 {
		originalFireRes := *item.FireRes
		scaledFireRes := int(float64(*item.FireRes) * 1.5)
		item.FireRes = &scaledFireRes
		fmt.Printf("\033[33mFire Resistance Scaled: %v fire resistance increased from %d to %d (1.5x scaling)\033[0m\n", item.Name, originalFireRes, scaledFireRes)
		return // Don't add additional fire resistance if it already exists
	}

	itemName := strings.ToLower(item.Name)

	// Check for fire-related keywords in the item name
	fireKeywords := []string{"flame", "fire", "salamander", "crimson", "burning", "blazing", "infernal", "molten", "ember", "igniting", "flamewalker", "flameguard"}

	hasFireName := false
	for _, keyword := range fireKeywords {
		if strings.Contains(itemName, keyword) {
			hasFireName = true
			break
		}
	}

	if !hasFireName {
		return
	}

	// Get the inventory type modifier for this item
	var invTypeModifier float64 = 1.0 // default
	if item.InventoryType != nil {
		if modifier, exists := config.InvTypeModifiers[*item.InventoryType]; exists {
			invTypeModifier = modifier
		}
	}

	// Calculate fire resistance based on modifier (10-25 range)
	// Higher modifier = higher fire resistance
	baseFireRes := 10.0
	maxFireRes := 25.0

	// Scale the fire resistance based on the modifier
	// Normalize the modifier to a 0-1 range (assuming modifiers range from ~0.3 to 1.0)
	normalizedModifier := (invTypeModifier - 0.3) / 0.7
	if normalizedModifier < 0 {
		normalizedModifier = 0
	}
	if normalizedModifier > 1 {
		normalizedModifier = 1
	}

	fireResistance := int(baseFireRes + (maxFireRes-baseFireRes)*normalizedModifier)

	// Set the fire resistance on the item
	item.FireRes = &fireResistance

	fmt.Printf("\033[32mFire Resistance Applied: %v gets %v fire resistance (modifier: %.2f)\033[0m\n", item.Name, fireResistance, invTypeModifier)
}

// enforceStatRequirements applies the user's stat requirements
func enforceStatRequirements(item *items.Item) {
	if item.InventoryType == nil {
		return
	}

	// Check if item is ring (11) or trinket (0)
	isRingOrTrinket := *item.InventoryType == 11 || *item.InventoryType == 0

	// Collect current stats
	stats := make(map[int]int)
	statCount := 0

	// Read current stats
	for i := 1; i <= 8; i++ {
		statType := getStatType(item, i)
		statValue := getStatValue(item, i)

		if statType != nil && statValue != nil && *statType != 0 && *statValue != 0 {
			stats[*statType] = *statValue
			statCount++
		}
	}

	// Apply mana regen cap (stat type 43)
	if manaRegen, exists := stats[43]; exists && manaRegen > 60 {
		stats[43] = 60
		fmt.Printf("\033[33mMana Regen Capped: %v mana regen reduced to 60\033[0m\n", item.Name)
	}

	// Find spell power (45) or attack power (38) and make it the highest stat
	maxStatValue := 0
	for _, value := range stats {
		if value > maxStatValue {
			maxStatValue = value
		}
	}

	// Boost spell power or attack power to be the max stat
	if spellPower, hasSpellPower := stats[45]; hasSpellPower {
		if spellPower < maxStatValue {
			stats[45] = maxStatValue + 10 // Make it slightly higher
			fmt.Printf("\033[36mSpell Power Boosted: %v spell power increased to %v\033[0m\n", item.Name, stats[45])
		}
	} else if attackPower, hasAttackPower := stats[38]; hasAttackPower {
		if attackPower < maxStatValue {
			stats[38] = maxStatValue + 10 // Make it slightly higher
			fmt.Printf("\033[36mAttack Power Boosted: %v attack power increased to %v\033[0m\n", item.Name, stats[38])
		}
	}

	// Ensure minimum 4 stats for non-rings/trinkets
	if !isRingOrTrinket && statCount < 4 {
		// Add additional stats to reach minimum of 4
		neededStats := 4 - statCount

		// Common stats to add based on item type
		commonStats := []int{7, 5, 3, 4} // Stamina, Intellect, Agility, Strength

		for i := 0; i < neededStats && i < len(commonStats); i++ {
			statType := commonStats[i]
			if _, exists := stats[statType]; !exists {
				// Add a reasonable stat value based on item level
				statValue := 50
				if item.ItemLevel != nil {
					statValue = *item.ItemLevel / 2 // Simple scaling
				}
				stats[statType] = statValue
				statCount++
			}
		}

		if statCount >= 4 {
			fmt.Printf("\033[35mStats Added: %v now has %v stats (minimum 4 enforced)\033[0m\n", item.Name, statCount)
		}
	}

	// Write stats back to item
	writeStatsToItem(item, stats)
}

// Helper function to get stat type
func getStatType(item *items.Item, index int) *int {
	switch index {
	case 1:
		return item.StatType1
	case 2:
		return item.StatType2
	case 3:
		return item.StatType3
	case 4:
		return item.StatType4
	case 5:
		return item.StatType5
	case 6:
		return item.StatType6
	case 7:
		return item.StatType7
	case 8:
		return item.StatType8
	default:
		return nil
	}
}

// Helper function to get stat value
func getStatValue(item *items.Item, index int) *int {
	switch index {
	case 1:
		return item.StatValue1
	case 2:
		return item.StatValue2
	case 3:
		return item.StatValue3
	case 4:
		return item.StatValue4
	case 5:
		return item.StatValue5
	case 6:
		return item.StatValue6
	case 7:
		return item.StatValue7
	case 8:
		return item.StatValue8
	default:
		return nil
	}
}

// Helper function to write stats back to item
func writeStatsToItem(item *items.Item, stats map[int]int) {
	i := 1
	for statType, statValue := range stats {
		if i > 8 {
			break
		}

		switch i {
		case 1:
			item.StatType1 = &statType
			item.StatValue1 = &statValue
		case 2:
			item.StatType2 = &statType
			item.StatValue2 = &statValue
		case 3:
			item.StatType3 = &statType
			item.StatValue3 = &statValue
		case 4:
			item.StatType4 = &statType
			item.StatValue4 = &statValue
		case 5:
			item.StatType5 = &statType
			item.StatValue5 = &statValue
		case 6:
			item.StatType6 = &statType
			item.StatValue6 = &statValue
		case 7:
			item.StatType7 = &statType
			item.StatValue7 = &statValue
		case 8:
			item.StatType8 = &statType
			item.StatValue8 = &statValue
		}
		i++
	}

	// Update stats count
	statCount := len(stats)
	item.StatsCount = &statCount
}

// GenerateItem generates and validates a Molten Core item
func (g *MoltenCoreGenerator) GenerateItem(dbItem mysql.DbItem, validateOnly bool) ItemGenerationResult {
	result := ItemGenerationResult{
		Success:   false,
		Errors:    []string{},
		Warnings:  []string{},
		Validated: false,
	}

	// Create item from database item
	item := items.ItemFromDbItem(dbItem)
	item.SetDifficulty(MOLTEN_CORE_DIFFICULTY)
	item.ApplyTierModifiers(MOLTEN_CORE_PHASE)

	// Initial scaling
	item.ScaleItem(g.itemLevel, g.quality)

	classType := item.GetClassUserType()
	if g.debug {
		log.Printf("Item: %s (Entry: %d) - Class: %d, Subclass: %d, ClassType: %s",
			item.Name, item.Entry, *item.Class, *item.Subclass, getClassString(classType))
	}

	// Handle subclass mapping for weapons
	subclassToUse := *item.Subclass
	if *item.Subclass == 8 {
		subclassToUse = 1 // two handed axe instead of sword
	}

	// Get high-level reference items
	highLevelItems, err := g.db.GetRaidPhase1Items(*item.Class, subclassToUse, 0, 0)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to get reference items: %v", err))
		return result
	}

	// Filter items by class type compatibility
	var compatibleChoices []items.Item
	for _, highLevelItem := range highLevelItems {
		highLevelItem := items.ItemFromDbItem(highLevelItem)
		highLevelItem.ScaleItem(*highLevelItem.ItemLevel, *item.Quality)
		highClassType := highLevelItem.GetClassUserType()

		if highClassType == classType {
			compatibleChoices = append(compatibleChoices, highLevelItem)
		}
	}

	if g.debug {
		log.Printf("Found %d compatible reference items for class type %s",
			len(compatibleChoices), getClassString(classType))
	}
	
	// If no compatible items found, try similar subclasses as fallback
	if len(compatibleChoices) == 0 {
		var similarSubclasses []int
		if *item.Class == 2 { // Weapons
			similarSubclasses = getSimilarWeaponSubclasses(subclassToUse)
		} else if *item.Class == 4 { // Armor
			similarSubclasses = getSimilarArmorSubclasses(subclassToUse)
		}
		
		if g.debug && len(similarSubclasses) > 0 {
			log.Printf("No compatible items found for subclass %d, trying similar subclasses: %v", subclassToUse, similarSubclasses)
		}
		
		// Try each similar subclass until we find compatible items
		for _, altSubclass := range similarSubclasses {
			altHighLevelItems, err := g.db.GetRaidPhase1Items(*item.Class, altSubclass, 0, 0)
			if err != nil {
				if g.debug {
					log.Printf("Failed to get reference items for subclass %d: %v", altSubclass, err)
				}
				continue
			}
			
			// Filter these alternative items by class type compatibility
			for _, highLevelItem := range altHighLevelItems {
				highLevelItem := items.ItemFromDbItem(highLevelItem)
				highLevelItem.ScaleItem(*highLevelItem.ItemLevel, *item.Quality)
				highClassType := highLevelItem.GetClassUserType()

				if highClassType == classType {
					compatibleChoices = append(compatibleChoices, highLevelItem)
				}
			}
			
			if len(compatibleChoices) > 0 {
				if g.debug {
					log.Printf("Found %d compatible fallback reference items using subclass %d for class type %s",
						len(compatibleChoices), altSubclass, getClassString(classType))
				}
				break // Found compatible items, stop searching
			}
		}
	}

	// Apply stats from reference item or generate manually
	var selectedReferenceItem *items.Item
	if len(compatibleChoices) > 0 {
		// Use random compatible reference item
		randHighLevelItem := compatibleChoices[rand.Intn(len(compatibleChoices))]
		// Store reference item for comparison
		selectedReferenceItem = &randHighLevelItem
		item.ApplyStats(randHighLevelItem)

		// Handle spell assignment: clear original spells and use reference item spells (except for trinkets)
		if !isTrinket(&item) {
			// For non-trinkets: clear original spells and copy from reference item
			copySpellsFromReference(&item, selectedReferenceItem)
			if g.debug {
				log.Printf("Copied spells from reference item %s to %s", selectedReferenceItem.Name, item.Name)
			}
		} else {
			// For trinkets: clear original spells (will be replaced by applyTrinketSpells later)
			clearItemSpells(&item)
			if g.debug {
				log.Printf("Cleared original spells from trinket %s (will be replaced by applyTrinketSpells)", item.Name)
			}
		}

		item.ScaleItem(g.itemLevel, g.quality)
	} else {
		// Manual scaling required - clear original spells for consistency
		if !isTrinket(&item) {
			clearItemSpells(&item)
			if g.debug {
				log.Printf("Cleared original spells from %s (no reference item available)", item.Name)
			}
		} else {
			clearItemSpells(&item)
			if g.debug {
				log.Printf("Cleared original spells from trinket %s (manual scaling)", item.Name)
			}
		}
		result.Warnings = append(result.Warnings, "No compatible reference items found, using manual scaling")
		item.ScaleItem(g.itemLevel, g.quality)
	}

	// Store reference item in result
	result.ReferenceItem = selectedReferenceItem

	// Apply Molten Core specific enhancements
	addFireResistanceIfNeeded(&item)

	// Apply trinket spells if applicable
	g.applyTrinketSpells(&item, classType)

	// Capture spell information for display
	var spellInfo *SpellDetails
	if item.SpellId1 != nil && *item.SpellId1 > 0 {
		spellInfo = g.getSpellDetails(*item.SpellId1)
	} else if item.SpellId2 != nil && *item.SpellId2 > 0 {
		spellInfo = g.getSpellDetails(*item.SpellId2)
	} else if item.SpellId3 != nil && *item.SpellId3 > 0 {
		spellInfo = g.getSpellDetails(*item.SpellId3)
	}
	result.SpellInfo = spellInfo

	// Enforce stat requirements and validation
	enforceStatRequirements(&item)

	// Validate the final item using enhanced validation
	validationErrors, validationWarnings, validationScore := g.validateItemAdvanced(&item, classType)
	result.Warnings = append(result.Warnings, validationWarnings...)

	if len(validationErrors) > 0 {
		result.Errors = append(result.Errors, validationErrors...)
		if !validateOnly {
			// Try to fix validation errors
			g.fixValidationErrors(&item, classType, validationErrors)
			
			// If still failing validation, try adding missing key stats
			fixedErrors, fixedWarnings, newScore := g.validateItemAdvanced(&item, classType)
			if len(fixedErrors) > 0 && classType != 7 { // Don't try to fix Generic class type
				if g.addMissingKeyStats(&item, classType) {
					// Re-validate after adding missing key stats
					finalErrors, finalWarnings, finalScore := g.validateItemAdvanced(&item, classType)
					if len(finalErrors) == 0 {
						result.Warnings = append(result.Warnings, fmt.Sprintf("Item auto-corrected with missing key stat (Score: %d → %d → %d)", validationScore, newScore, finalScore))
						validationErrors = finalErrors
						validationScore = finalScore
						result.Warnings = append(result.Warnings, finalWarnings...)
					} else {
						result.Errors = finalErrors
						result.Warnings = append(result.Warnings, finalWarnings...)
					}
				} else {
					result.Errors = fixedErrors
					result.Warnings = append(result.Warnings, fixedWarnings...)
				}
			} else if len(fixedErrors) == 0 {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Item auto-corrected (Score: %d → %d)", validationScore, newScore))
				validationErrors = fixedErrors
				validationScore = newScore
				result.Warnings = append(result.Warnings, fixedWarnings...)
			} else {
				result.Errors = fixedErrors
				result.Warnings = append(result.Warnings, fixedWarnings...)
			}
		}
	}

	// Add validation score to warnings for transparency
	if validationScore < 100 {
		result.Warnings = append(result.Warnings, fmt.Sprintf("Validation Score: %d/100", validationScore))
	}

	result.Item = &item
	result.Success = len(validationErrors) == 0
	result.Validated = true

	return result
}

// applyTrinketSpells applies appropriate trinket spells based on class type
// This is MANDATORY for all trinkets - they must have a spell
func (g *MoltenCoreGenerator) applyTrinketSpells(item *items.Item, classType int) {
	if item.InventoryType == nil || *item.InventoryType != 0 { // Not a trinket
		return
	}

	var spellOptions []int
	var classDescription string

	switch classType {
	case 4, 5: // Mage, Healer
		spellOptions = spellTrinketSpells
		classDescription = "spell caster"
	case 1, 2, 3: // Melee DPS, Ranged
		spellOptions = meleeTrinketSpells
		classDescription = "melee/ranged DPS"
	case 6: // Tank
		spellOptions = tankTrinketSpells
		classDescription = "tank"
	default:
		// Fallback to melee spells for unknown class types
		spellOptions = meleeTrinketSpells
		classDescription = "unknown (defaulting to melee)"
		if g.debug {
			log.Printf("Unknown class type %d for trinket %s, defaulting to melee spells", classType, item.Name)
		}
	}

	// MANDATORY: All trinkets must have a spell
	if len(spellOptions) > 0 {
		selectedSpell := spellOptions[rand.Intn(len(spellOptions))]
		item.SpellId1 = &selectedSpell
		if g.debug {
			log.Printf("Applied %s trinket spell %d to %s", classDescription, selectedSpell, item.Name)
		}
	} else {
		// This should never happen, but provide fallback
		log.Printf("ERROR: No spell options available for class type %d, assigning default spell", classType)
		defaultSpell := 60493 // Default to spell power trinket
		item.SpellId1 = &defaultSpell
	}
}

// validateItem validates an item against stat priorities and requirements
func (g *MoltenCoreGenerator) validateItem(item *items.Item, classType int) []string {
	var errors []string

	// Get stat priorities for this class type
	priorities, exists := classStatPriorities[classType]
	if !exists {
		errors = append(errors, fmt.Sprintf("Unknown class type: %d", classType))
		return errors
	}

	// Check for required primary stats
	currentStats := g.getCurrentStats(item)
	for _, requiredStat := range priorities.Primary {
		if _, hasstat := currentStats[requiredStat]; !hasstat {
			statName := config.StatModifierNames[requiredStat]
			errors = append(errors, fmt.Sprintf("Missing required primary stat: %s (%d)", statName, requiredStat))
		}
	}

	// Validate stat limits based on plan requirements
	for statType, statValue := range currentStats {
		switch statType {
		case 43: // Mana Regeneration
			if statValue > 60 {
				errors = append(errors, fmt.Sprintf("Mana regeneration (%d) exceeds limit of 60", statValue))
			}
		case 12: // Defense Rating
			if statValue > 80 {
				errors = append(errors, fmt.Sprintf("Defense rating (%d) exceeds limit of 80", statValue))
			}
		case 38, 45: // Attack Power, Spell Power
			if (classType == 1 || classType == 2 || classType == 3) && statType == 38 {
				// Attack Power should be highest for DPS
				if !g.isHighestStat(currentStats, statType) {
					errors = append(errors, "Attack Power should be the highest stat for DPS items")
				}
			} else if (classType == 4 || classType == 5) && statType == 45 {
				// Spell Power should be highest for casters
				if !g.isHighestStat(currentStats, statType) {
					errors = append(errors, "Spell Power should be the highest stat for caster items")
				}
			}
		}
	}

	return errors
}

// getCurrentStats extracts current stats from an item
func (g *MoltenCoreGenerator) getCurrentStats(item *items.Item) map[int]int {
	stats := make(map[int]int)

	for i := 1; i <= 8; i++ {
		statType := getStatType(item, i)
		statValue := getStatValue(item, i)
		if statType != nil && statValue != nil && *statType > 0 && *statValue > 0 {
			stats[*statType] = *statValue
		}
	}

	return stats
}

// isHighestStat checks if a stat type has the highest value among all stats
func (g *MoltenCoreGenerator) isHighestStat(stats map[int]int, statType int) bool {
	statValue, exists := stats[statType]
	if !exists {
		return false
	}

	for _, value := range stats {
		if value > statValue {
			return false
		}
	}
	return true
}

// fixValidationErrors attempts to fix common validation errors
func (g *MoltenCoreGenerator) fixValidationErrors(item *items.Item, classType int, errors []string) {
	// This is a simplified fix implementation
	// In a production system, this would be more sophisticated
	for _, err := range errors {
		if strings.Contains(err, "Mana regeneration") && strings.Contains(err, "exceeds limit") {
			// Cap mana regeneration at 60
			g.capStatValue(item, 43, 60)
		} else if strings.Contains(err, "Defense rating") && strings.Contains(err, "exceeds limit") {
			// Cap defense rating at 80
			g.capStatValue(item, 12, 80)
		}
	}
}

// capStatValue caps a specific stat at a maximum value
func (g *MoltenCoreGenerator) capStatValue(item *items.Item, statType, maxValue int) {
	for i := 1; i <= 8; i++ {
		currentStatType := getStatType(item, i)
		currentStatValue := getStatValue(item, i)
		if currentStatType != nil && currentStatValue != nil &&
			*currentStatType == statType && *currentStatValue > maxValue {
			*currentStatValue = maxValue
			if g.debug {
				log.Printf("Capped stat %d to %d for item %s", statType, maxValue, item.Name)
			}
			break
		}
	}
}

// printThreeWayComparison shows before, reference item, and after scaling comparison with spell information
func printThreeWayComparison(originalItem, referenceItem, scaledItem *items.Item, spellInfo *SpellDetails, classType int, originalArmor, originalFireRes, originalItemLevel int) {
	fmt.Printf("\n" + strings.Repeat("=", 100) + "\n")
	fmt.Printf("📊 ITEM SCALING COMPARISON: %s (Entry: %d)\n", originalItem.Name, originalItem.Entry)
	fmt.Printf("Class Type: %s | Item Level: %d → %d | Quality: %d\n",
		getClassString(classType),
		originalItemLevel,
		getItemLevel(scaledItem),
		*scaledItem.Quality)

	// Display item type information
	slotName := getItemSlotName(scaledItem)
	materialType := getMaterialTypeName(scaledItem)
	weaponSubclass := getWeaponSubclassName(scaledItem)

	if materialType != "" {
		fmt.Printf("Item Type: %s %s\n", materialType, slotName)
	} else if weaponSubclass != "" {
		fmt.Printf("Item Type: %s %s\n", weaponSubclass, slotName)
	} else {
		fmt.Printf("Item Type: %s\n", slotName)
	}

	// Display spell information if available
	if spellInfo != nil {
		fmt.Printf("🪄 Assigned Spell: [%d] %s\n", spellInfo.SpellID, spellInfo.SpellName)
		if spellInfo.Description != "" {
			fmt.Printf("📜 Description: %s\n", spellInfo.Description)
		}
	}

	fmt.Printf(strings.Repeat("=", 100) + "\n")

	// Print three-way comparison header
	if referenceItem != nil {
		fmt.Printf("🔶 Reference Item: %s (Entry: %d)\n", referenceItem.Name, referenceItem.Entry)
		fmt.Printf("%-30s | %-30s | %-30s\n", "🔸 BEFORE SCALING", "🔶 REFERENCE ITEM", "🔹 AFTER SCALING")
	} else {
		fmt.Printf("%-45s | %-45s\n", "🔸 BEFORE SCALING", "🔹 AFTER SCALING (Manual)")
	}
	fmt.Printf(strings.Repeat("-", 100) + "\n")

	originalStats := extractItemStats(originalItem)
	scaledStats := extractItemStats(scaledItem)
	var referenceStats map[int]int
	if referenceItem != nil {
		referenceStats = extractItemStats(referenceItem)
	}

	// Get all unique stat types
	allStatTypes := make(map[int]bool)
	for statType := range originalStats {
		allStatTypes[statType] = true
	}
	for statType := range scaledStats {
		allStatTypes[statType] = true
	}
	if referenceStats != nil {
		for statType := range referenceStats {
			allStatTypes[statType] = true
		}
	}

	// Print stats comparison
	for statType := range allStatTypes {
		originalValue := 0
		scaledValue := 0
		referenceValue := 0
		statName := getStatName(statType)

		if value, exists := originalStats[statType]; exists {
			originalValue = value
		}
		if value, exists := scaledStats[statType]; exists {
			scaledValue = value
		}
		if referenceStats != nil {
			if value, exists := referenceStats[statType]; exists {
				referenceValue = value
			}
		}

		change := ""
		if scaledValue > originalValue {
			if originalValue == 0 {
				change = "✨ NEW"
			} else {
				change = fmt.Sprintf("📈 +%d", scaledValue-originalValue)
			}
		} else if scaledValue < originalValue {
			change = fmt.Sprintf("📉 -%d", originalValue-scaledValue)
		} else if scaledValue > 0 {
			change = "➡️ SAME"
		}

		if change != "" {
			if referenceItem != nil {
				fmt.Printf("%-25s %3d | %-25s %3d | %-25s %3d %s\n",
					statName, originalValue,
					statName, referenceValue,
					statName, scaledValue, change)
			} else {
				fmt.Printf("%-25s %3d | %-25s %3d %s\n",
					statName, originalValue,
					statName, scaledValue, change)
			}
		}
	}

	// Show armor comparison using stored original value
	if scaledItem.Armor != nil && *scaledItem.Armor > 0 {
		scaledArmorValue := *scaledItem.Armor
		referenceArmorValue := 0
		if referenceItem != nil && referenceItem.Armor != nil {
			referenceArmorValue = *referenceItem.Armor
		}

		change := ""
		if scaledArmorValue > originalArmor {
			change = fmt.Sprintf("📈 +%d", scaledArmorValue-originalArmor)
		} else if scaledArmorValue < originalArmor {
			change = fmt.Sprintf("📉 -%d", originalArmor-scaledArmorValue)
		} else {
			change = "➡️ SAME"
		}

		if referenceItem != nil {
			fmt.Printf("%-25s %3d | %-25s %3d | %-25s %3d %s\n",
				"Armor", originalArmor,
				"Armor", referenceArmorValue,
				"Armor", scaledArmorValue, change)
		} else {
			fmt.Printf("%-25s %3d | %-25s %3d %s\n",
				"Armor", originalArmor,
				"Armor", scaledArmorValue, change)
		}
	}

	// Show Fire Resistance comparison using stored original value
	scaledFireRes := getFireResistance(scaledItem)
	referenceFireRes := 0
	if referenceItem != nil {
		referenceFireRes = getFireResistance(referenceItem)
	}
	if originalFireRes > 0 || scaledFireRes > 0 || referenceFireRes > 0 {
		change := ""
		if scaledFireRes > originalFireRes {
			if originalFireRes == 0 {
				change = "✨ NEW"
			} else {
				change = fmt.Sprintf("📈 +%d", scaledFireRes-originalFireRes)
			}
		} else if scaledFireRes < originalFireRes {
			change = fmt.Sprintf("📉 -%d", originalFireRes-scaledFireRes)
		} else {
			change = "➡️ SAME"
		}

		if referenceItem != nil {
			fmt.Printf("%-25s %3d | %-25s %3d | %-25s %3d %s\n",
				"Fire Resistance", originalFireRes,
				"Fire Resistance", referenceFireRes,
				"Fire Resistance", scaledFireRes, change)
		} else {
			fmt.Printf("%-25s %3d | %-25s %3d %s\n",
				"Fire Resistance", originalFireRes,
				"Fire Resistance", scaledFireRes, change)
		}
	}

	fmt.Printf(strings.Repeat("=", 100) + "\n")
}

// printItemComparison shows clean before/after item scaling comparison
func printItemComparisonWithOriginals(originalItem, scaledItem *items.Item, classType int, originalArmor, originalFireRes, originalItemLevel int) {
	fmt.Printf("\n" + strings.Repeat("=", 80) + "\n")
	fmt.Printf("📊 ITEM SCALING COMPARISON: %s (Entry: %d)\n", originalItem.Name, originalItem.Entry)
	fmt.Printf("Class Type: %s | Item Level: %d → %d | Quality: %d\n",
		getClassString(classType),
		originalItemLevel,
		getItemLevel(scaledItem),
		*scaledItem.Quality)

	// Display item type information
	slotName := getItemSlotName(scaledItem)
	materialType := getMaterialTypeName(scaledItem)
	weaponSubclass := getWeaponSubclassName(scaledItem)

	if materialType != "" {
		fmt.Printf("Item Type: %s %s\n", materialType, slotName)
	} else if weaponSubclass != "" {
		fmt.Printf("Item Type: %s %s\n", weaponSubclass, slotName)
	} else {
		fmt.Printf("Item Type: %s\n", slotName)
	}
	fmt.Printf(strings.Repeat("=", 80) + "\n")

	// Print side-by-side comparison
	fmt.Printf("%-40s | %-40s\n", "🔸 BEFORE SCALING", "🔹 AFTER SCALING")
	fmt.Printf(strings.Repeat("-", 80) + "\n")

	originalStats := extractItemStats(originalItem)
	scaledStats := extractItemStats(scaledItem)

	// Get all unique stat types
	allStatTypes := make(map[int]bool)
	for statType := range originalStats {
		allStatTypes[statType] = true
	}
	for statType := range scaledStats {
		allStatTypes[statType] = true
	}

	// Print stats comparison
	for statType := range allStatTypes {
		originalValue := 0
		scaledValue := 0
		statName := getStatName(statType)

		if value, exists := originalStats[statType]; exists {
			originalValue = value
		}
		if value, exists := scaledStats[statType]; exists {
			scaledValue = value
		}

		change := ""
		if scaledValue > originalValue {
			if originalValue == 0 {
				change = "✨ NEW"
			} else {
				change = fmt.Sprintf("📈 +%d", scaledValue-originalValue)
			}
		} else if scaledValue < originalValue {
			change = fmt.Sprintf("📉 -%d", originalValue-scaledValue)
		} else if scaledValue > 0 {
			change = "➡️ SAME"
		}

		if change != "" {
			fmt.Printf("%-25s %3d | %-25s %3d %s\n",
				statName, originalValue,
				statName, scaledValue, change)
		}
	}

	// Show armor comparison if applicable using stored original value
	if scaledItem.Armor != nil && *scaledItem.Armor > 0 {
		scaledArmorValue := *scaledItem.Armor

		change := ""
		if scaledArmorValue > originalArmor {
			change = fmt.Sprintf("📈 +%d", scaledArmorValue-originalArmor)
		} else if scaledArmorValue < originalArmor {
			change = fmt.Sprintf("📉 -%d", originalArmor-scaledArmorValue)
		} else {
			change = "➡️ SAME"
		}
		fmt.Printf("%-25s %3d | %-25s %3d %s\n",
			"Armor", originalArmor,
			"Armor", scaledArmorValue, change)
	}

	// Show DPS/Damage comparison for weapons
	if isWeapon(originalItem) || isWeapon(scaledItem) {
		originalDPS := calculateDPS(originalItem)
		scaledDPS := calculateDPS(scaledItem)
		if originalDPS > 0 || scaledDPS > 0 {
			change := ""
			if scaledDPS > originalDPS {
				change = fmt.Sprintf("📈 +%.1f", scaledDPS-originalDPS)
			} else if scaledDPS < originalDPS {
				change = fmt.Sprintf("📉 -%.1f", originalDPS-scaledDPS)
			} else if scaledDPS > 0 {
				change = "➡️ SAME"
			} else {
				change = "✨ NEW"
			}
			fmt.Printf("%-25s %3.1f | %-25s %3.1f %s\n",
				"DPS", originalDPS,
				"DPS", scaledDPS, change)
		}

		// Show damage range
		originalMinDmg := getDamageMin(originalItem)
		originalMaxDmg := getDamageMax(originalItem)
		scaledMinDmg := getDamageMin(scaledItem)
		scaledMaxDmg := getDamageMax(scaledItem)
		if originalMinDmg > 0 || originalMaxDmg > 0 || scaledMinDmg > 0 || scaledMaxDmg > 0 {
			fmt.Printf("%-25s %d-%d | %-25s %d-%d\n",
				"Damage Range", originalMinDmg, originalMaxDmg,
				"Damage Range", scaledMinDmg, scaledMaxDmg)
		}
	}

	// Show Fire Resistance comparison using stored original value
	scaledFireRes := getFireResistance(scaledItem)
	if originalFireRes > 0 || scaledFireRes > 0 {
		change := ""
		if scaledFireRes > originalFireRes {
			if originalFireRes == 0 {
				change = "✨ NEW"
			} else {
				change = fmt.Sprintf("📈 +%d", scaledFireRes-originalFireRes)
			}
		} else if scaledFireRes < originalFireRes {
			change = fmt.Sprintf("📉 -%d", originalFireRes-scaledFireRes)
		} else {
			change = "➡️ SAME"
		}
		fmt.Printf("%-25s %3d | %-25s %3d %s\n",
			"Fire Resistance", originalFireRes,
			"Fire Resistance", scaledFireRes, change)
	}

	fmt.Printf(strings.Repeat("=", 80) + "\n")
}

func printItemComparison(originalItem, scaledItem *items.Item, classType int) {
	fmt.Printf("\n" + strings.Repeat("=", 80) + "\n")
	fmt.Printf("📊 ITEM SCALING COMPARISON: %s (Entry: %d)\n", originalItem.Name, originalItem.Entry)
	fmt.Printf("Class Type: %s | Item Level: %d → %d | Quality: %d\n",
		getClassString(classType),
		getItemLevel(originalItem),
		getItemLevel(scaledItem),
		*scaledItem.Quality)

	// Display item type information
	slotName := getItemSlotName(scaledItem)
	materialType := getMaterialTypeName(scaledItem)
	weaponSubclass := getWeaponSubclassName(scaledItem)

	if materialType != "" {
		fmt.Printf("Item Type: %s %s\n", materialType, slotName)
	} else if weaponSubclass != "" {
		fmt.Printf("Item Type: %s %s\n", weaponSubclass, slotName)
	} else {
		fmt.Printf("Item Type: %s\n", slotName)
	}
	fmt.Printf(strings.Repeat("=", 80) + "\n")

	// Print side-by-side comparison
	fmt.Printf("%-40s | %-40s\n", "🔸 BEFORE SCALING", "🔹 AFTER SCALING")
	fmt.Printf(strings.Repeat("-", 80) + "\n")

	originalStats := extractItemStats(originalItem)
	scaledStats := extractItemStats(scaledItem)

	// Get all unique stat types
	allStatTypes := make(map[int]bool)
	for statType := range originalStats {
		allStatTypes[statType] = true
	}
	for statType := range scaledStats {
		allStatTypes[statType] = true
	}

	// Print stats comparison
	for statType := range allStatTypes {
		statName := getStatName(statType)
		originalValue := originalStats[statType]
		scaledValue := scaledStats[statType]

		if originalValue > 0 || scaledValue > 0 {
			change := ""
			if originalValue == 0 {
				change = "✨ NEW"
			} else if scaledValue > originalValue {
				change = fmt.Sprintf("📈 +%d", scaledValue-originalValue)
			} else if scaledValue < originalValue {
				change = fmt.Sprintf("📉 -%d", originalValue-scaledValue)
			} else {
				change = "➡️ SAME"
			}

			fmt.Printf("%-25s %3d | %-25s %3d %s\n",
				statName, originalValue,
				statName, scaledValue, change)
		}
	}

	// Show armor comparison if applicable
	if scaledItem.Armor != nil && *scaledItem.Armor > 0 {
		// Calculate what the armor would be at the original item level
		originalArmorValue := calculateOriginalArmor(scaledItem, getItemLevel(originalItem))
		scaledArmorValue := *scaledItem.Armor

		change := ""
		if scaledArmorValue > originalArmorValue {
			change = fmt.Sprintf("📈 +%d", scaledArmorValue-originalArmorValue)
		} else if scaledArmorValue < originalArmorValue {
			change = fmt.Sprintf("📉 -%d", originalArmorValue-scaledArmorValue)
		} else {
			change = "➡️ SAME"
		}
		fmt.Printf("%-25s %3d | %-25s %3d %s\n",
			"Armor", originalArmorValue,
			"Armor", scaledArmorValue, change)
	}

	// Show DPS/Damage comparison for weapons
	if isWeapon(originalItem) || isWeapon(scaledItem) {
		originalDPS := calculateDPS(originalItem)
		scaledDPS := calculateDPS(scaledItem)
		if originalDPS > 0 || scaledDPS > 0 {
			change := ""
			if scaledDPS > originalDPS {
				change = fmt.Sprintf("📈 +%.1f", scaledDPS-originalDPS)
			} else if scaledDPS < originalDPS {
				change = fmt.Sprintf("📉 -%.1f", originalDPS-scaledDPS)
			} else if scaledDPS > 0 {
				change = "➡️ SAME"
			} else {
				change = "✨ NEW"
			}
			fmt.Printf("%-25s %3.1f | %-25s %3.1f %s\n",
				"DPS", originalDPS,
				"DPS", scaledDPS, change)
		}

		// Show damage range
		originalMinDmg := getDamageMin(originalItem)
		originalMaxDmg := getDamageMax(originalItem)
		scaledMinDmg := getDamageMin(scaledItem)
		scaledMaxDmg := getDamageMax(scaledItem)
		if originalMinDmg > 0 || originalMaxDmg > 0 || scaledMinDmg > 0 || scaledMaxDmg > 0 {
			fmt.Printf("%-25s %d-%d | %-25s %d-%d\n",
				"Damage Range", originalMinDmg, originalMaxDmg,
				"Damage Range", scaledMinDmg, scaledMaxDmg)
		}
	}

	// Show Fire Resistance comparison
	originalFireRes := getFireResistance(originalItem)
	scaledFireRes := getFireResistance(scaledItem)
	if originalFireRes > 0 || scaledFireRes > 0 {
		change := ""
		if scaledFireRes > originalFireRes {
			if originalFireRes == 0 {
				change = "✨ NEW"
			} else {
				change = fmt.Sprintf("📈 +%d", scaledFireRes-originalFireRes)
			}
		} else if scaledFireRes < originalFireRes {
			change = fmt.Sprintf("📉 -%d", originalFireRes-scaledFireRes)
		} else {
			change = "➡️ SAME"
		}
		fmt.Printf("%-25s %3d | %-25s %3d %s\n",
			"Fire Resistance", originalFireRes,
			"Fire Resistance", scaledFireRes, change)
	}

	fmt.Printf(strings.Repeat("=", 80) + "\n")
}

// extractItemStats extracts all stats from an item into a map
func extractItemStats(item *items.Item) map[int]int {
	stats := make(map[int]int)
	for i := 1; i <= 8; i++ {
		statType := getStatType(item, i)
		statValue := getStatValue(item, i)
		if statType != nil && statValue != nil && *statType > 0 && *statValue > 0 {
			stats[*statType] = *statValue
		}
	}
	return stats
}

// getStatName returns human-readable stat name
func getStatName(statType int) string {
	if name, exists := config.StatModifierNames[statType]; exists {
		return name
	}
	return fmt.Sprintf("Unknown(%d)", statType)
}

// getItemLevel safely gets item level
func getItemLevel(item *items.Item) int {
	if item.ItemLevel != nil {
		return *item.ItemLevel
	}
	return 0
}

// Enhanced validation with detailed scoring and constraint checking
func (g *MoltenCoreGenerator) validateItemAdvanced(item *items.Item, classType int) ([]string, []string, int) {
	var errors []string
	var warnings []string
	var score int = 100 // Start with perfect score

	// Get stat priorities for this class type
	priorities, exists := classStatPriorities[classType]
	if !exists {
		errors = append(errors, fmt.Sprintf("Unknown class type: %d", classType))
		return errors, warnings, 0
	}

	currentStats := g.getCurrentStats(item)
	itemType := g.getItemType(item)

	// 1. PRIMARY STAT VALIDATION (Critical - 40 points)
	missingPrimary := 0
	for _, requiredStat := range priorities.Primary {
		if _, hasstat := currentStats[requiredStat]; !hasstat {
			// Check if this stat should be on this item type
			if g.shouldHaveStat(itemType, requiredStat, classType) {
				statName := getStatName(requiredStat)
				errors = append(errors, fmt.Sprintf("Missing CRITICAL primary stat: %s", statName))
				missingPrimary++
				score -= 20 // Heavy penalty
			}
		}
	}

	// 2. STAT LIMIT VALIDATION (Critical - 30 points)
	for statType, statValue := range currentStats {
		switch statType {
		case 43: // Mana Regeneration
			if statValue > 60 {
				errors = append(errors, fmt.Sprintf("Mana Regen (%d) exceeds limit of 60", statValue))
				score -= 15
			} else if statValue > 45 {
				warnings = append(warnings, fmt.Sprintf("Mana Regen (%d) is high, consider reducing", statValue))
				score -= 5
			}
		case 12: // Defense Rating
			if statValue > 80 {
				errors = append(errors, fmt.Sprintf("Defense Rating (%d) exceeds limit of 80", statValue))
				score -= 15
			}
		}
	}

	// 3. POWER STAT PRIORITY VALIDATION (Important - 20 points)
	highestStatValue := g.getHighestStatValue(currentStats)
	for statType, statValue := range currentStats {
		switch statType {
		case 38: // Attack Power
			if classType == 1 || classType == 2 || classType == 3 {
				if statValue < highestStatValue {
					errors = append(errors, "Attack Power should be the highest stat for DPS items")
					score -= 10
				}
			}
		case 45: // Spell Power
			if classType == 4 || classType == 5 {
				if statValue < highestStatValue {
					errors = append(errors, "Spell Power should be the highest stat for caster items")
					score -= 10
				}
			}
		}
	}

	// 4. STAMINA VALIDATION (Context-dependent - 10 points)
	staminaValue := currentStats[7] // Stamina
	if g.shouldHaveStamina(itemType) {
		if staminaValue == 0 {
			warnings = append(warnings, "Item should have Stamina but doesn't")
			score -= 5
		}
	} else {
		if staminaValue > 0 {
			warnings = append(warnings, "Item has Stamina but typically shouldn't (weapon/ring/trinket/neck)")
			score -= 2
		}
	}

	// 5. STAT COUNT VALIDATION (Important - 15 points)
	statCount := len(currentStats)
	expectedStatRange := g.getExpectedStatCount(itemType)
	if statCount < expectedStatRange.Min {
		errors = append(errors, fmt.Sprintf("%s items should have at least %d stats, but has %d",
			itemType, expectedStatRange.Min, statCount))
		score -= 10
	} else if statCount > expectedStatRange.Max {
		warnings = append(warnings, fmt.Sprintf("%s items should have at most %d stats, but has %d",
			itemType, expectedStatRange.Max, statCount))
		score -= 5
	} else if statCount < expectedStatRange.Optimal {
		warnings = append(warnings, fmt.Sprintf("%s items optimally have %d stats, but has %d",
			itemType, expectedStatRange.Optimal, statCount))
		score -= 3
	}

	// 6. TRINKET SPELL VALIDATION (Critical for trinkets - 20 points)
	if itemType == "trinket" {
		if !g.hasTrinketSpell(item) {
			errors = append(errors, "Trinkets MUST have a spell assigned based on class type")
			score -= 20 // Heavy penalty for missing trinket spell
		} else {
			// Validate spell is appropriate for class type
			if !g.isTrinketSpellAppropriate(item, classType) {
				warnings = append(warnings, "Trinket spell may not be optimal for this class type")
				score -= 5
			}
		}
	}

	return errors, warnings, score
}

// Helper functions for advanced validation
func (g *MoltenCoreGenerator) getItemType(item *items.Item) string {
	if item.InventoryType == nil {
		return "unknown"
	}
	switch *item.InventoryType {
	case 0:
		return "trinket"
	case 2:
		return "neck"
	case 11:
		return "ring"
	case 13, 17, 21, 22:
		return "weapon"
	default:
		return "armor"
	}
}

func (g *MoltenCoreGenerator) shouldHaveStat(itemType string, statType, classType int) bool {
	// Stamina logic is handled separately
	if statType == 7 {
		return g.shouldHaveStamina(itemType)
	}
	// Most other primary stats should be on most items
	return true
}

func (g *MoltenCoreGenerator) shouldHaveStamina(itemType string) bool {
	// From plan: "Most items should have stamina unless weapon, ring, trinket, necklace"
	return itemType != "weapon" && itemType != "ring" && itemType != "trinket" && itemType != "neck"
}

func (g *MoltenCoreGenerator) getHighestStatValue(stats map[int]int) int {
	highest := 0
	for _, value := range stats {
		if value > highest {
			highest = value
		}
	}
	return highest
}

// getExpectedStatCount returns the expected stat count range for different item types
func (g *MoltenCoreGenerator) getExpectedStatCount(itemType string) StatCountRange {
	switch itemType {
	case "trinket":
		return StatCountRange{Min: 2, Max: 2, Optimal: 2}
	case "ring":
		return StatCountRange{Min: 3, Max: 4, Optimal: 4}
	case "neck":
		return StatCountRange{Min: 3, Max: 4, Optimal: 4}
	case "weapon":
		return StatCountRange{Min: 3, Max: 5, Optimal: 4}
	case "armor":
		return StatCountRange{Min: 5, Max: 6, Optimal: 6}
	default:
		// Default to armor requirements for unknown types
		return StatCountRange{Min: 5, Max: 6, Optimal: 6}
	}
}

// hasTrinketSpell checks if a trinket has any spell assigned
func (g *MoltenCoreGenerator) hasTrinketSpell(item *items.Item) bool {
	return (item.SpellId1 != nil && *item.SpellId1 > 0) ||
		(item.SpellId2 != nil && *item.SpellId2 > 0) ||
		(item.SpellId3 != nil && *item.SpellId3 > 0)
}

// isTrinketSpellAppropriate checks if the assigned trinket spell matches the class type
func (g *MoltenCoreGenerator) isTrinketSpellAppropriate(item *items.Item, classType int) bool {
	// Get the first assigned spell ID
	var assignedSpell int
	if item.SpellId1 != nil && *item.SpellId1 > 0 {
		assignedSpell = *item.SpellId1
	} else if item.SpellId2 != nil && *item.SpellId2 > 0 {
		assignedSpell = *item.SpellId2
	} else if item.SpellId3 != nil && *item.SpellId3 > 0 {
		assignedSpell = *item.SpellId3
	} else {
		return false // No spell assigned
	}

	// Check if the spell is in the appropriate category for the class type
	switch classType {
	case 4, 5: // Mage, Healer - should have spell trinket spells
		for _, spellId := range spellTrinketSpells {
			if assignedSpell == spellId {
				return true
			}
		}
	case 1, 2, 3: // Melee DPS, Ranged - should have melee trinket spells
		for _, spellId := range meleeTrinketSpells {
			if assignedSpell == spellId {
				return true
			}
		}
	case 6: // Tank - should have tank trinket spells
		for _, spellId := range tankTrinketSpells {
			if assignedSpell == spellId {
				return true
			}
		}
	default:
		// For unknown class types, any spell is considered appropriate
		return true
	}

	return false
}

// Helper functions for enhanced item comparison output

// isWeapon checks if an item is a weapon based on inventory type
func isWeapon(item *items.Item) bool {
	if item.InventoryType == nil {
		return false
	}
	// Weapon inventory types: 13, 17, 21, 22, 15, 25, 26
	weaponTypes := []int{13, 17, 21, 22, 15, 25, 26}
	for _, weaponType := range weaponTypes {
		if *item.InventoryType == weaponType {
			return true
		}
	}
	return false
}

// calculateDPS calculates the DPS of a weapon
func calculateDPS(item *items.Item) float64 {
	if !isWeapon(item) {
		return 0.0
	}

	minDmg := getDamageMin(item)
	maxDmg := getDamageMax(item)
	delay := getWeaponDelay(item)

	if delay == 0 {
		return 0.0
	}

	avgDamage := float64(minDmg+maxDmg) / 2.0
	return (avgDamage * 1000.0) / float64(delay) // Convert delay from ms to seconds
}

// getDamageMin gets minimum damage from item
func getDamageMin(item *items.Item) int {
	if item.MinDmg1 != nil {
		return int(*item.MinDmg1)
	}
	return 0
}

// getDamageMax gets maximum damage from item
func getDamageMax(item *items.Item) int {
	if item.MaxDmg1 != nil {
		return int(*item.MaxDmg1)
	}
	return 0
}

// getWeaponDelay gets weapon delay/speed
func getWeaponDelay(item *items.Item) int {
	if item.Delay != nil {
		return int(*item.Delay)
	}
	return 0
}

// getFireResistance gets fire resistance from item
func getFireResistance(item *items.Item) int {
	if item.FireRes != nil {
		return *item.FireRes
	}
	return 0
}

// getMaterialTypeName returns the material type name based on subclass for armor
func getMaterialTypeName(item *items.Item) string {
	if item.Class == nil || *item.Class != 4 { // Not armor
		return ""
	}
	if item.Subclass == nil {
		return "Unknown"
	}

	materialTypes := map[int]string{
		1: "Cloth",
		2: "Leather",
		3: "Mail",
		4: "Plate",
		6: "Shield",
		0: "Miscellaneous",
	}

	if materialName, exists := materialTypes[*item.Subclass]; exists {
		return materialName
	}
	return fmt.Sprintf("Unknown(%d)", *item.Subclass)
}

// getItemSlotName returns the item slot/type name based on inventory type
func getItemSlotName(item *items.Item) string {
	if item.InventoryType == nil {
		return "Unknown"
	}

	slotNames := map[int]string{
		1:  "Head",
		2:  "Neck",
		3:  "Shoulder",
		4:  "Shirt",
		5:  "Chest",
		6:  "Waist",
		7:  "Legs",
		8:  "Feet",
		9:  "Wrists",
		10: "Hands",
		11: "Finger",
		12: "Trinket",
		13: "One-Hand",
		14: "Shield",
		15: "Ranged",
		16: "Back",
		17: "Two-Hand",
		18: "Bag",
		19: "Tabard",
		20: "Robe",
		21: "Main Hand",
		22: "Off Hand",
		23: "Holdable",
		24: "Ammo",
		25: "Thrown",
		26: "Ranged Right",
		28: "Relic",
	}

	if slotName, exists := slotNames[*item.InventoryType]; exists {
		return slotName
	}
	return fmt.Sprintf("Unknown(%d)", *item.InventoryType)
}

// getWeaponSubclassName returns weapon subclass name
func getWeaponSubclassName(item *items.Item) string {
	if item.Class == nil || *item.Class != 2 { // Not a weapon
		return ""
	}
	if item.Subclass == nil {
		return "Unknown"
	}

	weaponSubclasses := map[int]string{
		0:  "Axe",
		1:  "Axe",
		2:  "Bow",
		3:  "Gun",
		4:  "Mace",
		5:  "Mace",
		6:  "Polearm",
		7:  "Sword",
		8:  "Sword",
		9:  "Obsolete",
		10: "Staff",
		11: "Exotic",
		12: "Exotic",
		13: "Fist Weapon",
		14: "Miscellaneous",
		15: "Dagger",
		16: "Thrown",
		17: "Spear",
		18: "Crossbow",
		19: "Wand",
		20: "Fishing Pole",
	}

	if weaponName, exists := weaponSubclasses[*item.Subclass]; exists {
		return weaponName
	}
	return fmt.Sprintf("Unknown(%d)", *item.Subclass)
}

// calculateOriginalArmor calculates what the armor value would be at the original item level
// using the same formula as ScaleArmor but with the original item level
func calculateOriginalArmor(item *items.Item, originalItemLevel int) int {
	// Only calculate for armor items
	if item.Class == nil || *item.Class != 4 {
		return 0
	}

	// Need quality and subclass for the calculation
	if item.Quality == nil || item.Subclass == nil {
		return 0
	}

	// Use the same modifiers as ScaleArmor
	qualityModifier, qOk := config.QualityModifiers[*item.Quality]
	materialModifier, mOk := config.MaterialModifiers[*item.Subclass]

	if !qOk || !mOk {
		return 0
	}

	// Calculate armor at original item level using the same formula
	originalArmorValue := math.Ceil(float64(originalItemLevel) * qualityModifier * materialModifier)
	return int(originalArmorValue)
}
