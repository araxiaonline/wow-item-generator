package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strings"

	"github.com/araxiaonline/endgame-item-generator/internal/config"
	"github.com/araxiaonline/endgame-item-generator/internal/db/mysql"
	"github.com/araxiaonline/endgame-item-generator/internal/items"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

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

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	godotenv.Load("../../.env")

	debug := flag.Bool("debug", false, "Enable verbose logging inside generator")
	difficulty := flag.Int("difficulty", 3, "set the difficulty of the dungeon, defaults to 3 (mythic) 4 (legendary) 5 (ascendant)")
	baselevel := flag.Int("baselevel", 80, "set the base level for items to be used, defaults to 80 this is required for levelUp flag")
	flag.Parse()

	if difficulty == nil || *difficulty < 3 || *difficulty > 5 {
		log.Fatal("difficulty must be between 3-5")
		os.Exit(1)
	}

	if baselevel == nil || *baselevel < 0 {
		log.Fatal("base level must be greater than 80")
		os.Exit(1)
	}

	var itemLevel *int = new(int)
	switch *difficulty {
	case 3:
		*itemLevel = config.MythicItemLevelStart
	case 4:
		*itemLevel = config.LegendaryItemLevelStart
	case 5:
		*itemLevel = config.AscendantItemLevelStart
	}

	if *debug {
		log.SetOutput(os.Stdout)
	} else {
		log.SetOutput(io.Discard)
	}

	// Connect to Mysql
	mysqlDb, err := mysql.Connect(&mysql.MySqlConfig{
		Host:     os.Getenv("DB_HOST"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Database: os.Getenv("DB_NAME"),
	})

	if err != nil {
		log.Fatal(err)
	}

	MAPID := 409
	ITEMLEVEL := 325
	QUALITY := 4

	// Total Hack here to get items from a specific dungeon that does not have stats on them
	rareItems, err := mysqlDb.GetBossMapItems(MAPID, 0, 0)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("<<<< Items to Process: %v >>>>>\n", len(rareItems))

	for _, dbItem := range rareItems {
		log.Printf("Item: %v Entry: %v\n", dbItem.Name, dbItem.Entry)
		item := items.ItemFromDbItem(dbItem)
		item.SetDifficulty(3)
		item.ApplyTierModifiers(1) // this is for phase 1 raids to be released.
		item.ScaleItem(ITEMLEVEL, QUALITY)

		classType := item.GetClassUserType()

		fmt.Printf("Item: %v Entry: %v lookup up Class: %v Subclass: %v\n", item.Name, item.Entry, *item.Class, *item.Subclass)

		subclassToUse := *item.Subclass
		if *item.Subclass == 8 {
			subclassToUse = 1 // two handed axe instead of sword
		}

		highLevelItems, err := mysqlDb.GetRaidPhase1Items(*item.Class, subclassToUse, 0, 0)
		if err != nil {
			log.Fatal(err)
		}

		// create a list for storing choices of items
		var choices []items.Item

		// print all the high level items that matched
		for _, highLevelItem := range highLevelItems {
			highLevelItem := items.ItemFromDbItem(highLevelItem)
			highLevelItem.ScaleItem(*highLevelItem.ItemLevel, *item.Quality)
			highClassType := highLevelItem.GetClassUserType()

			if highClassType == classType {
				choices = append(choices, highLevelItem)
			}

			// fmt.Printf("OriginalItem: %v (%v) Item ClassType: %v vs %v HighLevel Item: %v(%v) \n", item.Name, item.Entry, getClassString(classType), getClassString(highClassType), highLevelItem.Name, highLevelItem.Entry)
		}

		if len(highLevelItems) == 0 {
			fmt.Printf("\033[31mItem: %v Entry: %v has no high level items\033[0m\n", item.Name, item.Entry)
		}

		fmt.Printf("<<<< High Level Items that match class Type: %v %v >>>>>\n", getClassString(classType), len(choices))

		// pick a random high level ite from the choice of item then scale it and display the results.
		if len(choices) > 0 {
			randHighLevelItem := choices[rand.Intn(len(choices))]

			fmt.Printf("New Item: %v Entry: %v \n", item.Name, item.Entry)
			fmt.Printf("ItemStat1: Type: %v (%v) Value: %v\n", *item.StatType1, config.StatModifierNames[*item.StatType1], *item.StatValue1)
			fmt.Printf("ItemStat2: Type: %v (%v) Value: %v\n", *item.StatType2, config.StatModifierNames[*item.StatType2], *item.StatValue2)
			fmt.Printf("ItemStat3: Type: %v (%v) Value: %v\n", *item.StatType3, config.StatModifierNames[*item.StatType3], *item.StatValue3)
			fmt.Printf("ItemStat4: Type: %v (%v) Value: %v\n", *item.StatType4, config.StatModifierNames[*item.StatType4], *item.StatValue4)
			fmt.Printf("ItemStat5: Type: %v (%v) Value: %v\n", *item.StatType5, config.StatModifierNames[*item.StatType5], *item.StatValue5)
			fmt.Printf("ItemStat6: Type: %v (%v) Value: %v\n", *item.StatType6, config.StatModifierNames[*item.StatType6], *item.StatValue6)
			fmt.Printf("ItemStat7: Type: %v (%v) Value: %v\n", *item.StatType7, config.StatModifierNames[*item.StatType7], *item.StatValue7)
			fmt.Printf("ItemStat8: Type: %v (%v) Value: %v\n", *item.StatType8, config.StatModifierNames[*item.StatType8], *item.StatValue8)

			item.ApplyStats(randHighLevelItem)
			item.ScaleItem(ITEMLEVEL, QUALITY)

			// Add fire resistance for fire-themed items
			addFireResistanceIfNeeded(&item)

			// Enforce stat requirements
			enforceStatRequirements(&item)
		} else {
			fmt.Printf("\033[31mItem: %v Entry: %v has no high level items HAS TO BE SCALED MANUALLY\033[0m\n", item.Name, item.Entry)

			fmt.Printf("ItemStat1: Type: %v (%v) Value: %v\n", *item.StatType1, config.StatModifierNames[*item.StatType1], *item.StatValue1)
			fmt.Printf("ItemStat2: Type: %v (%v) Value: %v\n", *item.StatType2, config.StatModifierNames[*item.StatType2], *item.StatValue2)
			fmt.Printf("ItemStat3: Type: %v (%v) Value: %v\n", *item.StatType3, config.StatModifierNames[*item.StatType3], *item.StatValue3)
			fmt.Printf("ItemStat4: Type: %v (%v) Value: %v\n", *item.StatType4, config.StatModifierNames[*item.StatType4], *item.StatValue4)
			fmt.Printf("ItemStat5: Type: %v (%v) Value: %v\n", *item.StatType5, config.StatModifierNames[*item.StatType5], *item.StatValue5)
			fmt.Printf("ItemStat6: Type: %v (%v) Value: %v\n", *item.StatType6, config.StatModifierNames[*item.StatType6], *item.StatValue6)
			fmt.Printf("ItemStat7: Type: %v (%v) Value: %v\n", *item.StatType7, config.StatModifierNames[*item.StatType7], *item.StatValue7)
			fmt.Printf("ItemStat8: Type: %v (%v) Value: %v\n", *item.StatType8, config.StatModifierNames[*item.StatType8], *item.StatValue8)

			item.ScaleItem(ITEMLEVEL, QUALITY)

			// Add fire resistance for fire-themed items
			addFireResistanceIfNeeded(&item)

			// Enforce stat requirements
			enforceStatRequirements(&item)
		}

		// if item.StatsCount == nil || *item.StatsCount == 0 {
		// 	fmt.Printf("Item: %v Entry: %v has no stats\n", item.Name, item.Entry)
		// } else {
		// 	// print all the individual stats
		// 	// fmt.Printf("ItemStat1: Type: %v Value: %v\n", *item.StatType1, *item.StatValue1)
		// 	// fmt.Printf("ItemStat2: Type: %v Value: %v\n", *item.StatType2, *item.StatValue2)
		// 	// fmt.Printf("ItemStat3: Type: %v Value: %v\n", *item.StatType3, *item.StatValue3)
		// 	// fmt.Printf("ItemStat4: Type: %v Value: %v\n", *item.StatType4, *item.StatValue4)
		// 	// fmt.Printf("ItemStat5: Type: %v Value: %v\n", *item.StatType5, *item.StatValue5)
		// 	// fmt.Printf("ItemStat6: Type: %v Value: %v\n", *item.StatType6, *item.StatValue6)
		// 	// fmt.Printf("ItemStat7: Type: %v Value: %v\n", *item.StatType7, *item.StatValue7)
		// }

		// spells, err := item.GetSpells()
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// for _, spell := range spells {
		// 	fmt.Printf("Spell: %v (%v) Description: %v Effect %v BasePoints %v\n", spell.Name, spell.ID, spell.Description, spell.Effect1, spell.EffectBasePoints1)
		// }

		// fmt.Println("\n")
	}
	os.Exit(0)
}

// addFireResistanceIfNeeded adds fire resistance to items with fire-related names
// based on the item's inventory type modifier from the config
func addFireResistanceIfNeeded(item *items.Item) {
	if item.Name == "" {
		return
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
