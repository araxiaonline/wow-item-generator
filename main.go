package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/araxiaonline/endgame-item-generator/internal/config"
	"github.com/araxiaonline/endgame-item-generator/internal/db/mysql"
	"github.com/araxiaonline/endgame-item-generator/internal/items"
	"github.com/araxiaonline/endgame-item-generator/internal/spells"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

// ItemScaler handles generic item scaling operations
type ItemScaler struct {
	db              *mysql.MySqlDb
	debug           bool
	itemLevel       int
	quality         int
	phase           int
	catchup         float64
	spellAssigner   *spells.ThematicSpellAssigner
	templateManager *items.StatTemplateManager
}

// ScalingResult holds the result of item scaling
type ScalingResult struct {
	Item          *items.Item
	ReferenceItem *items.Item
	Success       bool
	Errors        []string
	Warnings      []string
	SpellSQL      []string // SQL statements for scaled spells
}

// InputSource defines how items are provided to the scaler
type InputSource interface {
	GetItems() ([]mysql.DbItem, error)
	GetName() string
}

// EntryListSource scales items from a list of entry IDs
type EntryListSource struct {
	Entries []int
	DB      *mysql.MySqlDb
}

func (e *EntryListSource) GetItems() ([]mysql.DbItem, error) {
	var items []mysql.DbItem
	for _, entry := range e.Entries {
		item, err := e.DB.GetItem(entry)
		if err != nil {
			log.Printf("Failed to get item %d: %v", entry, err)
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

func (e *EntryListSource) GetName() string {
	return fmt.Sprintf("Entry List (%d items)", len(e.Entries))
}

// CSVSource scales items from a CSV file with entry IDs
type CSVSource struct {
	FilePath string
	DB       *mysql.MySqlDb
}

func (c *CSVSource) GetItems() ([]mysql.DbItem, error) {
	file, err := os.Open(c.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %v", err)
	}

	var items []mysql.DbItem
	for i, record := range records {
		if i == 0 && strings.ToLower(record[0]) == "entry" {
			continue // Skip header row
		}

		if len(record) == 0 {
			continue
		}

		entry, err := strconv.Atoi(record[0])
		if err != nil {
			log.Printf("Invalid entry ID in CSV row %d: %s", i+1, record[0])
			continue
		}

		item, err := c.DB.GetItem(entry)
		if err != nil {
			log.Printf("Failed to get item %d from CSV: %v", entry, err)
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

func (c *CSVSource) GetName() string {
	return fmt.Sprintf("CSV File: %s", c.FilePath)
}

// SQLQuerySource scales items from a custom SQL query
type SQLQuerySource struct {
	Query string
	DB    *mysql.MySqlDb
}

func (s *SQLQuerySource) GetItems() ([]mysql.DbItem, error) {
	// For now, use GetRarePlusItems as a fallback since GetItemsByQuery doesn't exist
	// This can be extended later to support custom queries
	return s.DB.GetRarePlusItems(0, 0)
}

func (s *SQLQuerySource) GetName() string {
	return "Custom SQL Query (using GetRarePlusItems)"
}

// MapItemsSource scales items from map/boss/gameobject data (like Molten Core)
type MapItemsSource struct {
	MapID             int
	BossEntries       []int
	GameObjectEntries []int
	DB                *mysql.MySqlDb
}

func (m *MapItemsSource) GetItems() ([]mysql.DbItem, error) {
	return m.DB.GetBossMapItems(m.MapID, m.BossEntries, m.GameObjectEntries, 0, 0)
}

func (m *MapItemsSource) GetName() string {
	return fmt.Sprintf("Map %d Items", m.MapID)
}

func NewItemScaler(db *mysql.MySqlDb, debug bool, itemLevel, quality, phase int, catchup float64) *ItemScaler {
	return &ItemScaler{
		db:              db,
		debug:           debug,
		itemLevel:       itemLevel,
		quality:         quality,
		phase:           phase,
		catchup:         catchup,
		spellAssigner:   spells.NewThematicSpellAssigner(db),
		templateManager: items.NewStatTemplateManager(debug),
	}
}

// ScaleItem performs the core scaling logic extracted from raid-gear/main.go
func (s *ItemScaler) ScaleItem(dbItem mysql.DbItem) ScalingResult {
	result := ScalingResult{
		Success:  false,
		Errors:   []string{},
		Warnings: []string{},
		SpellSQL: []string{},
	}

	// Create item from database item
	item := items.ItemFromDbItem(dbItem)

	// Set quality if not already epic/legendary
	if *item.Quality < 4 {
		*item.Quality = s.quality
	}

	// Check if item should have a thematic spell assigned
	if s.debug {
		log.Printf("Checking if %s should have thematic spell assigned", item.Name)
	}

	if s.shouldAssignThematicSpell(item) {
		if s.debug {
			log.Printf("Item %s qualifies for thematic spell assignment", item.Name)
		}
		err := s.assignThematicSpell(&item)
		if err != nil {
			if s.debug {
				log.Printf("Failed to assign thematic spell to %s: %v", item.Name, err)
			}
		} else {
			if s.debug {
				log.Printf("Successfully assigned thematic spell to %s", item.Name)
			}
		}
	} else {
		if s.debug {
			log.Printf("Item %s does not qualify for thematic spell assignment", item.Name)
		}
	}

	// Get item class type for reference item matching
	classType := item.GetClassUserType()

	if s.debug {
		log.Printf("Scaling item: %s (Entry: %d) - Class: %d, Subclass: %d, ClassType: %d",
			item.Name, item.Entry, *item.Class, *item.Subclass, classType)
	}

	var referenceItem items.Item

	if s.phase == 0 {
		// Phase 0: Use simple stat templates instead of complex reference matching
		s.templateManager.ApplySimpleStatTemplate(&item)
		if s.debug {
			log.Printf("Using simple stat template for phase 0")
		}
	} else {
		// Phase 1+: Find reference items and use reference item scaling
		referenceItems, err := s.findReferenceItems(&item)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to find reference items: %v", err))
			return result
		}

		if len(referenceItems) == 0 {
			result.Errors = append(result.Errors, "No compatible reference items found")
			result.Warnings = append(result.Warnings, "Manual review required - no reference items available")
			return result
		}

		// Select random reference item
		referenceItem = referenceItems[rand.Intn(len(referenceItems))]
		result.ReferenceItem = &referenceItem

		// Apply stats from reference item
		item.ApplyStats(referenceItem)
	}
	item.ScaleItemWithPhase(s.itemLevel, s.quality, s.phase)

	// Apply tier modifiers if phase is specified
	if s.phase > 0 || s.catchup != 1.0 {
		item.ApplyTierModifiersWithCatchup(s.phase, s.catchup)
	}

	// Collect spell SQL for any scaled spells
	for _, spell := range item.Spells {
		if spell.Scaled {
			spellSQL := spells.SpellToSql(spell, s.quality)
			result.SpellSQL = append(result.SpellSQL, spellSQL)
			if s.debug {
				log.Printf("Generated SQL for scaled spell: %s (ID: %d)", spell.Name, spell.ID)
			}
		}
	}

	if s.debug {
		if s.phase == 0 {
			log.Printf("Successfully scaled %s using simple stat template", item.Name)
		} else {
			log.Printf("Successfully scaled %s using reference item %s", item.Name, referenceItem.Name)
		}
	}

	result.Item = &item
	result.Success = true
	return result
}

// shouldAssignThematicSpell determines if an item should get a thematic spell
func (s *ItemScaler) shouldAssignThematicSpell(item items.Item) bool {
	if s.debug {
		log.Printf("DEBUG: shouldAssignThematicSpell - checking item %s", item.Name)
		log.Printf("DEBUG: SpellId1: %v, SpellId2: %v, SpellId3: %v",
			item.SpellId1, item.SpellId2, item.SpellId3)
	}

	// Check if item already has spells
	if item.SpellId1 != nil && *item.SpellId1 != 0 {
		if s.debug {
			log.Printf("DEBUG: Item already has spell in slot 1: %d", *item.SpellId1)
		}
		return false
	}
	if item.SpellId2 != nil && *item.SpellId2 != 0 {
		if s.debug {
			log.Printf("DEBUG: Item already has spell in slot 2: %d", *item.SpellId2)
		}
		return false
	}
	if item.SpellId3 != nil && *item.SpellId3 != 0 {
		if s.debug {
			log.Printf("DEBUG: Item already has spell in slot 3: %d", *item.SpellId3)
		}
		return false
	}

	if s.debug {
		log.Printf("DEBUG: Item has no existing spells, checking if it should have a proc")
	}

	// Use the thematic spell assigner to determine eligibility
	return s.spellAssigner.ShouldHaveProc(*item.Class, *item.Subclass, *item.ItemLevel, *item.Quality, item.Name)
}

// assignThematicSpell assigns a thematic spell to an item
func (s *ItemScaler) assignThematicSpell(item *items.Item) error {
	spellId, err := s.spellAssigner.AssignThematicSpell(item.Entry, *item.Class, *item.Subclass, *item.ItemLevel, *item.Quality, item.Name)
	if err != nil {
		return err
	}

	// Assign to first available spell slot
	if item.SpellId1 == nil || *item.SpellId1 == 0 {
		*item.SpellId1 = spellId
		*item.SpellTrigger1 = 2 // Chance on hit
		if s.debug {
			log.Printf("Assigned spell %d to %s in slot 1", spellId, item.Name)
		}
	} else if item.SpellId2 == nil || *item.SpellId2 == 0 {
		*item.SpellId2 = spellId
		*item.SpellTrigger2 = 2 // Chance on hit
		if s.debug {
			log.Printf("Assigned spell %d to %s in slot 2", spellId, item.Name)
		}
	} else if item.SpellId3 == nil || *item.SpellId3 == 0 {
		*item.SpellId3 = spellId
		*item.SpellTrigger3 = 2 // Chance on hit
		if s.debug {
			log.Printf("Assigned spell %d to %s in slot 3", spellId, item.Name)
		}
	}

	return nil
}

// findReferenceItems uses the sophisticated reference matching logic from raid-gear
func (s *ItemScaler) findReferenceItems(item *items.Item) ([]items.Item, error) {
	classType := item.GetClassUserType()

	// Handle subclass mapping for weapons (from raid-gear logic)
	subclassToUse := *item.Subclass
	if *item.Subclass == 8 {
		subclassToUse = 1 // two handed axe instead of sword
	}

	// Get high-level reference items
	highLevelItems, err := s.db.GetRaidPhase1Items(*item.Class, subclassToUse, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get reference items: %v", err)
	}

	// Filter items by class type AND inventory type compatibility
	var compatibleChoices []items.Item
	var classOnlyChoices []items.Item
	var anyWeaponChoices []items.Item

	for _, highLevelItem := range highLevelItems {
		refItem := items.ItemFromDbItem(highLevelItem)
		refItem.ScaleItem(*refItem.ItemLevel, *item.Quality)
		refClassType := refItem.GetClassUserType()

		// Check both class type and inventory type compatibility
		classMatch := refClassType == classType
		invMatch := (item.InventoryType != nil && refItem.InventoryType != nil &&
			*item.InventoryType == *refItem.InventoryType)

		if classMatch && invMatch {
			// Perfect match - both class type and inventory type
			compatibleChoices = append(compatibleChoices, refItem)
		} else if classMatch {
			// Class type match only - good fallback
			classOnlyChoices = append(classOnlyChoices, refItem)
		} else if *item.Class == 2 && *refItem.Class == 2 {
			// Any weapon as last resort for weapons
			anyWeaponChoices = append(anyWeaponChoices, refItem)
		}
	}

	if s.debug {
		log.Printf("Reference items found - Perfect: %d, Class-only: %d, Any-weapon: %d",
			len(compatibleChoices), len(classOnlyChoices), len(anyWeaponChoices))
	}

	// Return best available match
	if len(compatibleChoices) > 0 {
		return compatibleChoices, nil
	} else if len(classOnlyChoices) > 0 {
		if s.debug {
			log.Printf("Using class-only match for %s", item.Name)
		}
		return classOnlyChoices, nil
	} else if len(anyWeaponChoices) > 0 {
		if s.debug {
			log.Printf("Using any-weapon fallback for %s", item.Name)
		}
		return anyWeaponChoices, nil
	}

	return compatibleChoices, nil
}

// printScalingComparison shows before/after comparison
func printScalingComparison(original, scaled *items.Item, reference *items.Item) {
	fmt.Printf("\n=== ITEM SCALING: %s (Entry: %d) ===\n", original.Name, original.Entry)
	if reference != nil {
		fmt.Printf("Reference Item: %s (Entry: %d)\n", reference.Name, reference.Entry)
	}

	// Show key stats comparison
	fmt.Printf("Item Level: %d → %d\n", getItemLevel(original), getItemLevel(scaled))
	fmt.Printf("Quality: %d → %d\n", *original.Quality, *scaled.Quality)

	// Show stat changes
	originalStats := extractStats(original)
	scaledStats := extractStats(scaled)

	fmt.Printf("Stats Changes:\n")
	allStats := make(map[int]bool)
	for stat := range originalStats {
		allStats[stat] = true
	}
	for stat := range scaledStats {
		allStats[stat] = true
	}

	for statType := range allStats {
		originalValue := originalStats[statType]
		scaledValue := scaledStats[statType]

		if originalValue != scaledValue {
			statName := getStatName(statType)
			fmt.Printf("  %s: %d → %d\n", statName, originalValue, scaledValue)
		}
	}
	fmt.Printf("=====================================\n\n")
}

// Helper functions
func getItemLevel(item *items.Item) int {
	if item.ItemLevel != nil {
		return *item.ItemLevel
	}
	return 0
}

func extractStats(item *items.Item) map[int]int {
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

func getStatName(statType int) string {
	if name, exists := config.StatModifierNames[statType]; exists {
		return name
	}
	return fmt.Sprintf("Unknown(%d)", statType)
}

func parseIntList(s string) []int {
	if s == "" {
		return []int{}
	}

	parts := strings.Split(s, ",")
	var result []int
	for _, part := range parts {
		if num, err := strconv.Atoi(strings.TrimSpace(part)); err == nil {
			result = append(result, num)
		}
	}
	return result
}

func main() {
	rand.Seed(time.Now().UnixNano())
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	godotenv.Load()

	// Command line flags
	debug := flag.Bool("debug", false, "Enable verbose logging")
	itemLevel := flag.Int("item-level", 325, "Target item level for scaling")
	quality := flag.Int("quality", 4, "Target quality (4=Epic, 5=Legendary)")
	phase := flag.Int("phase", 0, "Tier phase for modifiers (0=no tier bonus, 1-5=tier levels)")
	catchup := flag.Float64("catchup", 1.0, "Catchup bonus multiplier (1.0=no bonus, 1.5=50% bonus)")
	baseLevel := flag.Int("base-level", 80, "Base level requirement for items")
	difficulty := flag.Int("difficulty", 3, "Difficulty level (3=Mythic, 4=Legendary, 5=Ascendant)")
	outputSQL := flag.Bool("sql", false, "Output SQL statements")

	// Input source flags
	entries := flag.String("entries", "", "Comma-separated list of entry IDs")
	csvFile := flag.String("csv", "", "CSV file with entry IDs")
	sqlQuery := flag.String("query", "", "Custom SQL query to get items")
	mapID := flag.Int("map-id", 0, "Map ID for boss/gameobject items")
	bossEntries := flag.String("boss-entries", "", "Comma-separated boss entry IDs")
	gameObjectEntries := flag.String("gameobject-entries", "", "Comma-separated gameobject entry IDs")

	flag.Parse()

	if *debug {
		log.SetOutput(os.Stdout)
	} else {
		log.SetOutput(io.Discard)
	}

	// Connect to database
	mysqlDb, err := mysql.Connect(&mysql.MySqlConfig{
		Host:     os.Getenv("DB_HOST"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Database: os.Getenv("DB_NAME"),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Determine input source
	var source InputSource
	sourceCount := 0

	if *entries != "" {
		entryList := parseIntList(*entries)
		source = &EntryListSource{Entries: entryList, DB: mysqlDb}
		sourceCount++
	}

	if *csvFile != "" {
		source = &CSVSource{FilePath: *csvFile, DB: mysqlDb}
		sourceCount++
	}

	if *sqlQuery != "" {
		source = &SQLQuerySource{Query: *sqlQuery, DB: mysqlDb}
		sourceCount++
	}

	if *mapID > 0 {
		source = &MapItemsSource{
			MapID:             *mapID,
			BossEntries:       parseIntList(*bossEntries),
			GameObjectEntries: parseIntList(*gameObjectEntries),
			DB:                mysqlDb,
		}
		sourceCount++
	}

	if sourceCount == 0 {
		fmt.Println("Error: No input source specified. Use one of:")
		fmt.Println("  --entries \"1234,5678\"")
		fmt.Println("  --csv items.csv")
		fmt.Println("  --query \"SELECT * FROM item_template WHERE ...\"")
		fmt.Println("  --map-id 409 --boss-entries \"11502\" --gameobject-entries \"179703\"")
		os.Exit(1)
	}

	if sourceCount > 1 {
		fmt.Println("Error: Only one input source can be specified at a time")
		os.Exit(1)
	}

	// Initialize scaler
	scaler := NewItemScaler(mysqlDb, *debug, *itemLevel, *quality, *phase, *catchup)

	// Get items from source
	dbItems, err := source.GetItems()
	if err != nil {
		log.Fatal("Failed to get items from source:", err)
	}

	fmt.Printf("🔧 Generic Item Scaler\n")
	fmt.Printf("Source: %s\n", source.GetName())
	fmt.Printf("Target: Item Level %d, Quality %d, Phase %d, Catchup %.1fx\n", *itemLevel, *quality, *phase, *catchup)
	fmt.Printf("Processing %d items...\n\n", len(dbItems))

	// Initialize SQL output if requested
	var sqlFile *os.File
	if *outputSQL {
		sqlFile, err = os.Create("scaled_items.sql")
		if err != nil {
			log.Fatal("Failed to create SQL output file:", err)
		}
		defer sqlFile.Close()

		sqlFile.WriteString("-- Generic Item Scaler Output\n")
		sqlFile.WriteString(fmt.Sprintf("-- Generated: %s\n", time.Now().Format("2006-01-02 15:04:05")))
		sqlFile.WriteString(fmt.Sprintf("-- Item Level: %d, Quality: %d, Phase: %d, Catchup: %.1fx\n\n", *itemLevel, *quality, *phase, *catchup))
	}

	// Process items
	successCount := 0
	for i, dbItem := range dbItems {
		fmt.Printf("[%d/%d] Processing: %s (Entry: %d)\n", i+1, len(dbItems), dbItem.Name, dbItem.Entry)

		// Store original for comparison
		originalItem := items.ItemFromDbItem(dbItem)

		result := scaler.ScaleItem(dbItem)

		if result.Success {
			successCount++
			fmt.Printf("✅ Successfully scaled %s\n", result.Item.Name)

			// Show comparison
			printScalingComparison(&originalItem, result.Item, result.ReferenceItem)

			// Output SQL if requested
			if *outputSQL && sqlFile != nil {
				// Write item SQL
				sqlStatement := items.ItemToSql(*result.Item, *baseLevel, *difficulty, true)
				sqlFile.WriteString(sqlStatement + "\n")

				// Write spell SQL for any scaled spells
				for _, spellSQL := range result.SpellSQL {
					sqlFile.WriteString(spellSQL + "\n")
				}
			}
		} else {
			fmt.Printf("❌ Failed to scale %s\n", dbItem.Name)
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
	fmt.Printf("🏆 Scaling Summary:\n")
	fmt.Printf("Total Items: %d\n", len(dbItems))
	fmt.Printf("Successful: %d\n", successCount)
	fmt.Printf("Failed: %d\n", len(dbItems)-successCount)
	fmt.Printf("Success Rate: %.1f%%\n", float64(successCount)/float64(len(dbItems))*100)

	if *outputSQL {
		fmt.Printf("SQL Output: scaled_items.sql\n")
	}
}
