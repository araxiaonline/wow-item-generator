package items

import (
	"errors"
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/araxiaonline/endgame-item-generator/internal/config"
	"github.com/araxiaonline/endgame-item-generator/internal/db/mysql"
	"github.com/araxiaonline/endgame-item-generator/internal/spells"
)

/**
 * For details about values of item int values use link below
 * @link https://www.azerothcore.org/wiki/item_template
 */
type Item struct {
	mysql.DbItem
	StatsMap      map[int]*ItemStat
	ConvStatCount int
	Spells        []spells.Spell
	Difficulty    int
}

// Use for storing item stats for all stats that will be scaled.
type ItemStat struct {
	Value    int
	Percent  float64
	Type     string
	AdjValue float64
}

type StatScaleParams struct {
	ItemLevel    int
	NewItemLevel int
	Quality      int
	ItemType     int
	StatTypeId   int
	StatValue    int
}

// Create a new item from the database item with deep copy to prevent shared pointer issues
func ItemFromDbItem(dbItem mysql.DbItem) Item {
	// Create a deep copy to avoid shared pointer references
	copy := dbItem

	// Create new pointers for all StatType fields
	if dbItem.StatType1 != nil {
		val := *dbItem.StatType1
		copy.StatType1 = &val
	}
	if dbItem.StatType2 != nil {
		val := *dbItem.StatType2
		copy.StatType2 = &val
	}
	if dbItem.StatType3 != nil {
		val := *dbItem.StatType3
		copy.StatType3 = &val
	}
	if dbItem.StatType4 != nil {
		val := *dbItem.StatType4
		copy.StatType4 = &val
	}
	if dbItem.StatType5 != nil {
		val := *dbItem.StatType5
		copy.StatType5 = &val
	}
	if dbItem.StatType6 != nil {
		val := *dbItem.StatType6
		copy.StatType6 = &val
	}
	if dbItem.StatType7 != nil {
		val := *dbItem.StatType7
		copy.StatType7 = &val
	}
	if dbItem.StatType8 != nil {
		val := *dbItem.StatType8
		copy.StatType8 = &val
	}
	if dbItem.StatType9 != nil {
		val := *dbItem.StatType9
		copy.StatType9 = &val
	}
	if dbItem.StatType10 != nil {
		val := *dbItem.StatType10
		copy.StatType10 = &val
	}

	// Create new pointers for all StatValue fields
	if dbItem.StatValue1 != nil {
		val := *dbItem.StatValue1
		copy.StatValue1 = &val
	}
	if dbItem.StatValue2 != nil {
		val := *dbItem.StatValue2
		copy.StatValue2 = &val
	}
	if dbItem.StatValue3 != nil {
		val := *dbItem.StatValue3
		copy.StatValue3 = &val
	}
	if dbItem.StatValue4 != nil {
		val := *dbItem.StatValue4
		copy.StatValue4 = &val
	}
	if dbItem.StatValue5 != nil {
		val := *dbItem.StatValue5
		copy.StatValue5 = &val
	}
	if dbItem.StatValue6 != nil {
		val := *dbItem.StatValue6
		copy.StatValue6 = &val
	}
	if dbItem.StatValue7 != nil {
		val := *dbItem.StatValue7
		copy.StatValue7 = &val
	}
	if dbItem.StatValue8 != nil {
		val := *dbItem.StatValue8
		copy.StatValue8 = &val
	}
	if dbItem.StatValue9 != nil {
		val := *dbItem.StatValue9
		copy.StatValue9 = &val
	}
	if dbItem.StatValue10 != nil {
		val := *dbItem.StatValue10
		copy.StatValue10 = &val
	}

	// Create new pointers for other critical fields
	if dbItem.Class != nil {
		val := *dbItem.Class
		copy.Class = &val
	}
	if dbItem.Subclass != nil {
		val := *dbItem.Subclass
		copy.Subclass = &val
	}
	if dbItem.Quality != nil {
		val := *dbItem.Quality
		copy.Quality = &val
	}
	if dbItem.ItemLevel != nil {
		val := *dbItem.ItemLevel
		copy.ItemLevel = &val
	}
	if dbItem.Armor != nil {
		val := *dbItem.Armor
		copy.Armor = &val
	}
	if dbItem.FireRes != nil {
		val := *dbItem.FireRes
		copy.FireRes = &val
	}
	if dbItem.Material != nil {
		val := *dbItem.Material
		copy.Material = &val
	}
	if dbItem.InventoryType != nil {
		val := *dbItem.InventoryType
		copy.InventoryType = &val
	}
	if dbItem.MinDmg1 != nil {
		val := *dbItem.MinDmg1
		copy.MinDmg1 = &val
	}
	if dbItem.MaxDmg1 != nil {
		val := *dbItem.MaxDmg1
		copy.MaxDmg1 = &val
	}
	if dbItem.Delay != nil {
		val := *dbItem.Delay
		copy.Delay = &val
	}
	if dbItem.SpellId1 != nil {
		val := *dbItem.SpellId1
		copy.SpellId1 = &val
	}
	if dbItem.SpellTrigger1 != nil {
		val := *dbItem.SpellTrigger1
		copy.SpellTrigger1 = &val
	}
	if dbItem.SpellId2 != nil {
		val := *dbItem.SpellId2
		copy.SpellId2 = &val
	}
	if dbItem.SpellTrigger2 != nil {
		val := *dbItem.SpellTrigger2
		copy.SpellTrigger2 = &val
	}
	if dbItem.SpellId3 != nil {
		val := *dbItem.SpellId3
		copy.SpellId3 = &val
	}
	if dbItem.SpellTrigger3 != nil {
		val := *dbItem.SpellTrigger3
		copy.SpellTrigger3 = &val
	}

	return Item{
		DbItem: copy,
	}
}

func (item Item) GetDifficulty() int {
	return item.Difficulty
}

func (item *Item) SetDifficulty(difficulty int) {
	item.Difficulty = difficulty
}

// scaleArmor calculates and updates the item's armor value based on its level, quality, and material subclass.
// It checks for nil pointers for critical scaling fields and valid map keys before performing calculations.
func (item *Item) ScaleArmor(itemLevel int) {
	// Ensure critical pointer fields for scaling are non-nil
	// Entry and Name are value types from the embedded DbItem and used for logging.
	if item.Class == nil || item.Armor == nil || item.Quality == nil || item.Subclass == nil || item.Material == nil {
		log.Printf("Item (Entry: %d, Name: '%s'): Cannot scale armor: one or more required pointer fields (Class, Armor, Quality, Subclass, Material) are nil.", item.Entry, item.Name)
		return
	}

	// Scale Armor Stats only if Class is 4 (ITEM_CLASS_ARMOR) and Armor > 0
	if *item.Class == 4 && *item.Armor > 0 {
		qualityModifier, qOk := config.QualityModifiers[*item.Quality]
		// Assuming item.Subclass is the correct key for MaterialModifiers as per original logic
		materialModifier, mOk := config.MaterialModifiers[*item.Subclass]

		if qOk && mOk {
			// preArmor := *item.Armor
			scaledArmorValue := math.Ceil(float64(itemLevel) * qualityModifier * materialModifier)
			*item.Armor = int(scaledArmorValue)

			// log.Printf("Item (Entry: %d, Name: '%s'): Scaled armor to %d (was %d). ItemLevel: %d, Quality: %d (Mod: %.2f), Subclass for MaterialMod: %d (Mod: %.2f). Actual Material field: %d",
			// 	item.Entry, item.Name, *item.Armor, preArmor, itemLevel, *item.Quality, qualityModifier, *item.Subclass, materialModifier, *item.Material)
		} else {
			var errorMessages []string
			if !qOk {
				errorMessages = append(errorMessages, fmt.Sprintf("invalid Quality key: %d", *item.Quality))
			}
			if !mOk {
				errorMessages = append(errorMessages, fmt.Sprintf("invalid Subclass key for MaterialModifier: %d", *item.Subclass))
			}
			log.Printf("Item (Entry: %d, Name: '%s'): Could not scale armor. Issues: %s. Original Armor: %d",
				item.Entry, item.Name, strings.Join(errorMessages, "; "), *item.Armor)
		}
	}
}

// Get the primary stat for an item (strength, agility, intellect, spirit, stamina)
func (item Item) GetPrimaryStat() (int, int, error) {
	var primaryStat int64
	var primaryVal int64

	for i := 1; i < 11; i++ {

		statType, err := item.GetField(fmt.Sprintf("StatType%v", i))
		if err != nil {
			log.Printf("Failed to get stat type %v for item: %v", i, item.Name)
			continue
		}
		if statType < 3 || statType > 7 {
			continue
		}

		val, err := item.GetField(fmt.Sprintf("StatValue%v", i))

		if err != nil {
			log.Printf("Failed to get stat value %v for item: %v", i, item.Name)
			continue
		}
		if val == 0 {
			continue
		}

		if int64(val) > primaryVal {
			primaryVal = int64(val)
			primaryStat = int64(statType)
		}
	}

	return int(primaryStat), int(primaryVal), nil
}

/**
 * Get the statIds for anitem as a slice of integers
 * @return []int
 */
func (item Item) GetStatList() ([]int, error) {

	statList := []int{}

	// Also need to get spells that on the item that convert to stats
	spells, err := item.GetSpells()
	if err != nil {
		log.Printf("Failed to get spells for item: %v", err)
		return nil, err
	}

	for _, spell := range spells {
		convStats, err := spell.ConvertToStats()
		if err != nil {
			log.Printf("Failed to convert spell to stats: %v for spell %v", err, spell.Name)
			continue
		}

		if len(convStats) == 0 {
			continue
		}

		for _, convStat := range convStats {
			statList = append(statList, convStat.StatType)
		}

	}

	for i := 1; i < 11; i++ {
		val, err := item.GetField(fmt.Sprintf("StatValue%v", i))

		if err != nil {
			log.Printf("Failed to get stat value %v for item: %v", i, item.Name)
			continue
		}
		if val == 0 {
			continue
		}

		statId, err := item.GetField(fmt.Sprintf("StatType%v", i))
		if err != nil {
			log.Printf("Failed to get stat type %v for item: %s", i, item.Name)
			continue
		}
		statList = append(statList, statId)
		slices.Sort(statList)

	}
	return statList, nil
}

func (i Item) GetDpsModifier() (float64, error) {
	if i.Subclass == nil {
		return 0, fmt.Errorf("subclass on the item is not set")
	}

	if i.Quality == nil {
		return 0, fmt.Errorf("quality is not set")
	}

	typeModifier := 0.0
	// Is a One-Handed Weapon
	if *i.Subclass == 0 || *i.Subclass == 4 || *i.Subclass == 13 || *i.Subclass == 15 || *i.Subclass == 7 {
		typeModifier = 0.58
	}

	// Is a Two-Handed Weapon
	if *i.Subclass == 1 || *i.Subclass == 5 || *i.Subclass == 6 || *i.Subclass == 8 || *i.Subclass == 10 || *i.Subclass == 17 {
		typeModifier = 0.85
	}

	// Ranged Weapons
	if *i.Subclass == 2 || *i.Subclass == 3 || *i.Subclass == 16 || *i.Subclass == 18 {
		typeModifier = 0.70
	}

	// Wands
	if *i.Subclass == 19 {
		typeModifier = 0.70
	}

	qualityModifier := 1.0

	// Add the quality modifier for the DPS calculation
	qualityModifier = config.QualityModifiers[*i.Quality]

	if typeModifier == 0 {
		return 0, fmt.Errorf("Item subclass is not a weapon %v", *i.Subclass)
	}

	return (qualityModifier * typeModifier), nil
}

// Get the current expected DPS of the item based on the min and max damage and delay
func (item Item) GetDPS() (float64, error) {

	if item.MinDmg1 == nil || item.MaxDmg1 == nil {
		return 0, fmt.Errorf("MinDmg1 or MaxDmg1 is not set")
	}

	if item.Delay == nil {
		return 0, fmt.Errorf("delay is not set")
	}

	dps := math.Round(((float64(*item.MinDmg1)+float64(*item.MaxDmg1))/2.0)/(float64(*item.Delay)/1000.0)*100) / 100

	return dps, nil
}

// Scales and items dps damage numbers based on a desired item level.
func (item *Item) ScaleDPS(oldLevel, level int) (float64, error) {
	if item.ItemLevel == nil {
		return 0, fmt.Errorf("ItemLevel is not set")
	}

	if item.Delay == nil {
		return 0, fmt.Errorf("delay is not set")
	}

	modifier, err := item.GetDpsModifier()
	if err != nil {
		log.Fatalf("Error getting DPS modifier: %v", err)
		return 0.0, err
	}

	scalingFactor := math.Pow(float64(level)/float64(oldLevel), 1.012)

	dps := modifier * float64(level) * scalingFactor
	adjDps := (dps * (*item.Delay / 1000) / 100)

	// Use deterministic values based on item entry instead of random values
	// We'll use the item entry to derive consistent min/max modifiers
	minMod := 70  // Default mid-range value (was 100-(rand.IntN(15)+22) which is ~70)
	maxMod := 135 // Default mid-range value (was 100+(rand.IntN(15)+28) which is ~135)

	minimum := adjDps * float64(minMod)
	maximum := adjDps * float64(maxMod)

	// If the weapon has secondary damage, scale that as well based on the ratio of the primary damage
	if *item.MinDmg2 != 0 && *item.MaxDmg2 != 0 {
		ratioMin := float64(*item.MinDmg2) / float64(*item.MinDmg1)
		ratioMax := float64(*item.MaxDmg2) / float64(*item.MaxDmg1)
		minimum2 := ratioMin * float64(minimum)
		maximum2 := ratioMax * float64(maximum)

		minimum2 = math.Ceil(minimum2)
		maximum2 = math.Ceil(maximum2)

		item.MinDmg2 = &minimum2
		item.MaxDmg2 = &maximum2

		// In order to balance the original scale of the secondary damage from primary
		minimum = minimum - float64(minimum2)*0.85
		maximum = maximum - float64(maximum2)*0.85
	}

	minimum = math.Ceil(minimum)
	maximum = math.Ceil(maximum)

	item.MinDmg1 = &minimum
	item.MaxDmg1 = &maximum

	dps, _ = item.GetDPS()

	return dps, nil
}

// Create a Map of stat percentages based on the current stat and how budgets are caluated
func (item Item) GetStatPercents(spellStats []spells.ConvItemStat) map[int]*ItemStat {

	statMap := make(map[int]*ItemStat)
	statBudget := 0.0

	values := reflect.ValueOf(item)
	for i := 1; i < 11; i++ {
		var statValue = values.FieldByName(fmt.Sprintf("StatValue%v", i)).Elem().Int()
		var statType = values.FieldByName(fmt.Sprintf("StatType%v", i)).Elem().Int()
		if statValue == 0 {
			continue
		}

		adjValue := float64(statValue) / config.StatModifiers[int(statType)]
		statBudget += adjValue
		statMap[int(statType)] = &ItemStat{
			Value:    int(statValue),
			Percent:  0.0,
			Type:     "Item",
			AdjValue: adjValue,
		}
	}

	// Calculate the total budget for the spell stats if we have some
	for _, spellStat := range spellStats {
		statBudget += float64(spellStat.Budget)
		statMap[spellStat.StatType] = &ItemStat{
			Value:    spellStat.StatValue,
			Percent:  0.0,
			Type:     "Spell",
			AdjValue: float64(spellStat.Budget),
		}
	}

	// Combine all stats and calculate percentages for each stat
	for statId, stat := range statMap {
		statMap[statId].Percent = math.Round(float64(stat.AdjValue)/statBudget*100) / 100
	}

	return statMap
}

// UpdateSpellID updates a spell ID in the item's spell slots
// It replaces oldSpellId with newSpellId in any of the item's spell slots
func (item *Item) UpdateSpellID(oldSpellId int, newSpellId int) bool {
	updated := false

	// Check and update each spell slot
	if item.SpellId1 != nil && *item.SpellId1 == oldSpellId {
		*item.SpellId1 = newSpellId
		updated = true
	}

	if item.SpellId2 != nil && *item.SpellId2 == oldSpellId {
		*item.SpellId2 = newSpellId
		updated = true
	}

	if item.SpellId3 != nil && *item.SpellId3 == oldSpellId {
		*item.SpellId3 = newSpellId
		updated = true
	}

	// If we updated any spell IDs, clear the cached spells so they'll be reloaded
	if updated {
		item.Spells = nil
	}

	return updated
}

// get an array of all the spells set on the item
func (item *Item) GetSpells() ([]spells.Spell, error) {
	// dont reload for the same item .
	if len(item.Spells) > 0 {
		return item.Spells, nil
	}

	spellList := []spells.Spell{}
	values := reflect.ValueOf(item)
	for i := 1; i < 4; i++ {
		spellId := values.Elem().FieldByName(fmt.Sprintf("SpellId%v", i)).Elem().Int()
		if spellId == 0 {
			continue
		}

		if spellId == -1 {
			continue
		}

		db, err := mysql.GetDb()
		if err != nil {
			return nil, err
		}

		dbspell, err := db.GetSpell(int(spellId))
		if err != nil {
			log.Printf("failed to get the spell: %v error: %v", spellId, err)
			continue
		}
		spell := spells.Spell{
			DbSpell:       dbspell,
			ItemSpellSlot: i,
		}

		spellList = append(spellList, spell)
	}
	item.Spells = spellList
	return spellList, nil
}

func (item *Item) GetNonStatSpells() ([]spells.Spell, error) {
	nonStatSpells := []spells.Spell{}
	for i := 1; i < 4; i++ {
		spellId, err := item.GetField(fmt.Sprintf("SpellId%v", i))

		if err != nil {
			log.Printf("Failed to get spell id %v", i)
			continue
		}

		if spellId == 0 {
			continue
		}

		db, err := mysql.GetDb()
		if err != nil {
			return nil, err
		}

		dbSpell, err := db.GetSpell(spellId)
		if err != nil {
			log.Printf("Failed to get spell %v", spellId)
			continue
		}

		spell := spells.Spell{
			DbSpell: dbSpell,
		}

		// Need to handle extended spell casts basically when a spell casts another spell and the base points are there
		// instead of with the item itself.
		// Can just create a new spell with base points, type and remove triggerspell and see what happens?
		// For now just skip anything not in our list.
		if spell.EffectAura1 == 42 || spell.EffectAura2 == 42 || spell.EffectAura3 == 42 {
			continue
		}

		spell.ItemSpellSlot = i
		nonStatSpells = append(nonStatSpells, spell)
	}
	return nonStatSpells, nil
}

// Applies status of one item to another overwriting the current stats
func (item *Item) ApplyStats(otherItem Item) (success bool, err error) {

	for i := 1; i < 11; i++ {
		statType, err := otherItem.GetField(fmt.Sprintf("StatType%v", i))
		if err != nil {
			return false, err
		}

		statValue, err := otherItem.GetField(fmt.Sprintf("StatValue%v", i))
		if err != nil {
			return false, err
		}

		item.UpdateField(fmt.Sprintf("StatType%v", i), statType)
		item.UpdateField(fmt.Sprintf("StatValue%v", i), statValue)
	}

	if otherItem.SocketColor1 != nil {
		item.SocketColor1 = otherItem.SocketColor1
		item.SocketContent1 = otherItem.SocketContent1
	}

	if otherItem.SocketColor2 != nil {
		item.SocketColor2 = otherItem.SocketColor2
		item.SocketContent2 = otherItem.SocketContent2
	}

	if otherItem.SocketColor3 != nil {
		item.SocketColor3 = otherItem.SocketColor3
		item.SocketContent3 = otherItem.SocketContent3
	}

	item.ItemLevel = otherItem.ItemLevel

	if otherItem.Armor != nil {
		item.Armor = otherItem.Armor
	}

	return true, nil
}

// Stat Formula scaler
// Ceiling of ((ItemLevel * QualityModifier * ItemTypeModifier)^1.7095 * %ofStats) ^ (1/1.7095)) / StatModifier
// i.e)   Green Strength Helmet  (((100 * 1.1 * 1.0)^1.705) * 1)^(1/1.7095) / 1.0 = 110 Strength on item
func (item *Item) ScaleItem(itemLevel int, itemQuality int) (bool, error) {
	var allSpellStats []spells.ConvItemStat
	if item.ItemLevel == nil {
		return false, errors.New("field itemLevel is not set")
	}

	if item.Quality == nil {
		return false, errors.New("field quality is not set")
	}

	fromItemLevel := *item.ItemLevel
	*item.ItemLevel = itemLevel

	// if an item quality is being forced than use it intead
	if *item.Quality < itemQuality {
		*item.Quality = itemQuality
	}

	log.Printf("Scaling item %v %v to item level %v and quality %v", item.Name, item.Entry, itemLevel, *item.Quality)

	// Get all the spell Stats on the item we can convert
	spellList, err := item.GetSpells()
	if err != nil {
		log.Printf("Failed to get spells for item: %v", err)
		return false, err
	}

	for i := 0; i < len(spellList); i++ {

		log.Printf("Spell %v (%v) Effect %v AuraEffect %v Spell Desc: %v basePoints %v", spellList[i].Name, spellList[i].ID, spellList[i].Effect1, spellList[i].EffectAura1, spellList[i].Description, spellList[i].EffectBasePoints1)

		convStats, err := spellList[i].ConvertToStats()
		if err != nil {
			log.Printf("Failed to convert spell to stats: %v for spell %v", err, spellList[i].Name)
			continue
		}

		if len(convStats) != 0 {
			item.UpdateField(fmt.Sprintf("SpellId%v", i+1), 0)
			item.UpdateField(fmt.Sprintf("SpellTrigger%v", i+1), 0)
		}

		allSpellStats = append(allSpellStats, convStats...)
	}

	allStats := item.GetStatPercents(allSpellStats)

	for statId, stat := range allStats {
		origValue := stat.Value

		scaleParams := StatScaleParams{
			ItemLevel:    fromItemLevel,
			NewItemLevel: *item.ItemLevel,
			Quality:      *item.Quality,
			ItemType:     *item.InventoryType,
			StatTypeId:   statId,
			StatValue:    stat.Value,
		}

		stat.Value = scaleStatv3(scaleParams, item.GetDifficulty())
		// stat.Value = scaleStatv2(itemLevel, *item.InventoryType, *item.Quality, stat.Percent, config.StatModifiers[statId])

		if statId == STAT.SpellPower && stat.Value < 100 {
			stat.Value = int(math.Round(float64(stat.Value) * 2.3785))
		}

		correctSpellAttackPower(item, allStats)

		log.Printf(">>>>>> Scaled : StatId: %v Type: %s Orig: %v - New Value: %v Percent: %v", statId, stat.Type, origValue, stat.Value, stat.Percent)
	}

	item.addStats(allStats)
	*item.StatsCount = len(allStats)

	// Scale Armor Stats
	item.ScaleArmor(itemLevel)

	// If the item is a weapon scale the DPS
	if *item.Class == 2 && *item.MinDmg1 > 0 {
		predps, err := item.GetDPS()
		if err != nil {
			log.Printf("Failed to get DPS: %v", err)
		}

		dps, err := item.ScaleDPS(fromItemLevel, itemLevel)
		if err != nil {
			log.Printf("Failed to scale DPS: %v", err)
			return false, err
		}
		log.Printf("DPS: %.1f scaled up from previous dps %v: Min %v - Max %v", dps, predps, *item.MinDmg1, *item.MaxDmg1)
	}

	item.cleanSpells()

	// Item is scaled now we have to determine if there are additional spell effects that need scaled.
	// this will be as simple as possible as the effects will just be a percentage of the item stats.
	// This could lead to some OP weapons that will need tuned down later. But for now, we will just scale at a
	// An example of this might on hit do $s1 nature damage over $d seconds.  We would just scale the $s1 value
	// based on the formula below. This assumes that Blizzard has already balanced the spell bonus against the
	// stats on the item level and quality.  This is a big assumption as the stats are not penalized
	// from having the extra damage.  This could really create some unique sought after weapons that exploit this.
	// modified ratio ((s1 / existing iLevel) * newIlevel) * (0.20 Rare or 0.30 Epic or 0.4 for Legendary).

	otherSpells, err := item.GetNonStatSpells()
	if err != nil {
		log.Printf("failed to get non stat spells: %v", err)
	}

	log.Printf("\n\n\n -------------------- COUNT OF other spells %v \n\n", len(otherSpells))

	item.Spells = []spells.Spell{}
	// Spells that can not be scaled into stats must get new spells scaled and created
	for _, spell := range otherSpells {
		// log.Printf(" --^^^^^^--------SPELL --- Spell %v (%v) Effect %v  AuraEffect %v Spell Desc: %v basePoints %v", spell.Name, spell.ID, spell.Effect1, spell.EffectAura1, spell.Description, spell.EffectBasePoints1)
		// Use ForceScaleSpell instead of ScaleSpell to ensure all spells are scaled properly
		// Determine tier based on item level
		tier := 1
		if itemLevel >= 200 {
			tier = 5
		} else if itemLevel >= 175 {
			tier = 4
		} else if itemLevel >= 150 {
			tier = 3
		} else if itemLevel >= 125 {
			tier = 2
		}

		log.Printf("Scaling spell %v (ID: %v) with tier %d modifier", spell.Name, spell.ID, tier)
		err := spell.ForceScaleSpell(fromItemLevel, itemLevel, *item.Quality, tier)
		if err != nil {
			log.Printf("Failed to scale spell: %v, Spell %v", err, spell.ID)
			continue
		}

		// ForceScaleSpell modifies the spell in place, so we use the original spell ID
		item.UpdateField(fmt.Sprintf("SpellId%v", spell.ItemSpellSlot), spell.ID)
		item.Spells = append(item.Spells, spell)

		// do one last check on all setting StatsCount based on how many stats have been set

		// log.Printf(" --SCALED---SPELL --- Spell %v (%v) Effect %v AuraEffect %v Spell Desc: %v basePoints %v", spell.Name, spell.ID, spell.Effect1, spell.EffectAura1, spell.Description, spell.EffectBasePoints1)
	}

	return true, nil

}

func (item *Item) GetField(fieldName string) (int, error) {
	itemValue := reflect.ValueOf(item).Elem()
	field := itemValue.FieldByName(fieldName)
	if !field.IsValid() {
		return 0, fmt.Errorf("failed to find field %s", fieldName)
	}

	switch field.Kind() {
	case reflect.Ptr:
		if field.IsNil() {
			return 0, fmt.Errorf("field %s is nil", fieldName)
		}
		return int(field.Elem().Int()), nil
	default:
		return 0, fmt.Errorf("field %s is not a pointer", fieldName)
	}
}

// Updates a dynamic field on the item struct useful for stat replacements or spells
func (item *Item) UpdateField(fieldName string, value int) {
	itemValue := reflect.ValueOf(item).Elem()
	field := itemValue.FieldByName(fieldName)
	if !field.IsValid() {
		log.Printf("failed to find field %s", fieldName)
		return
	}

	switch field.Kind() {
	case reflect.Ptr:
		newValue := reflect.ValueOf(&value)
		field.Set(newValue)
	default:
	}
}

func (item *Item) emptyStats() {
	*item.StatType1 = 0
	*item.StatValue1 = 0
	*item.StatType2 = 0
	*item.StatValue2 = 0
	*item.StatType3 = 0
	*item.StatValue3 = 0
	*item.StatType4 = 0
	*item.StatValue4 = 0
	*item.StatType5 = 0
	*item.StatValue5 = 0
	*item.StatType6 = 0
	*item.StatValue6 = 0
	*item.StatType7 = 0
	*item.StatValue7 = 0
	*item.StatType8 = 0
	*item.StatValue8 = 0
	*item.StatType9 = 0
	*item.StatValue9 = 0
	*item.StatType10 = 0
	*item.StatValue10 = 0
}

// Cleans up spells from the item that have been converted to stats and leaves only the ones that are not
func (item *Item) cleanSpells() {
	for i := 1; i < 3; i++ {
		currentId, err := item.GetField(fmt.Sprintf("SpellId%v", i))

		log.Printf("Checking spell id %v - value %v", i, currentId)
		if err != nil {
			log.Printf("ERROR: Failed to get spell id %v err: %v", i, err)
			continue
		}

		// if there no spellId set then check the next one if it is set move it and clear it
		if currentId == 0 {
			nextSpellId, err := item.GetField(fmt.Sprintf("SpellId%v", i+1))
			if err != nil {
				log.Printf("ERROR: Failed to get spell id %v err: %v", i+1, err)
			}

			if nextSpellId != 0 {
				item.UpdateField(fmt.Sprintf("SpellId%v", i), nextSpellId)
				item.UpdateField(fmt.Sprintf("SpellId%v", i+1), 0)
				log.Printf("Moved spell %v to %v to replace removed spell", nextSpellId, i)
				continue
			}

			continue
		}

	}
}

func (item *Item) addStats(stats map[int]*ItemStat) {
	item.emptyStats()
	i := 1

	// itemValue := reflect.ValueOf(item).Elem() // Get value of underlying struct

	for statId, stat := range stats {
		if i > 10 {
			break
		}

		statTypeField := fmt.Sprintf("StatType%d", i)
		statValueField := fmt.Sprintf("StatValue%d", i)

		// MP5 adjustment
		if statId == 43 {
			stat.Value = int(math.Round(float64(stat.Value) * 0.85))
		}

		if statId == 12 {
			stat.Value = int(math.Round(float64(stat.Value) * 0.85))
		}

		if statId == 12 {
			stat.Value = int(math.Round(float64(stat.Value) * 0.75))
		}

		if statId == 13 {
			stat.Value = int(math.Round(float64(stat.Value) * 0.75))
		}

		if statId == 31 {
			stat.Value = int(math.Round(float64(stat.Value) * 0.65))
		}

		// Update the item with new stats from scaling
		item.UpdateField(statTypeField, statId)
		item.UpdateField(statValueField, stat.Value)

		// Get the stats for logging purposes
		// tmpType, _ := item.GetField(statTypeField)
		// tmpStat, _ := item.GetField(statValueField)
		// log.Printf("Updated %s to %v, %s to %v", statTypeField, tmpType, statValueField, tmpStat)

		i++
	}
}

// Scale formula ((ItemLevel * QualityModifier * ItemTypeModifier)^1.7095 * %ofStats) ^ (1/1.7095)) / StatModifier
func scaleStat(itemLevel int, itemType int, itemQuality int, percOfStat float64, statModifier float64) int {
	scaledUp := (math.Pow((float64(itemLevel)*config.QualityModifiers[itemQuality]*config.InvTypeModifiers[itemType]), 1.7095) * percOfStat)

	// leaving modifier off for now but not changing signature in case I need to add it back
	_ = statModifier
	return int(math.Ceil(math.Pow(scaledUp, 1/1.7095))) // normalized
}

func scaleStatv2(scaleParams StatScaleParams) int {
	modifier := config.QualityModifiers[scaleParams.Quality] * config.ScalingFactor[scaleParams.StatTypeId]
	modifier *= float64(scaleParams.NewItemLevel) / float64(scaleParams.ItemLevel)
	scaledValue := float64(scaleParams.StatValue) * modifier // * config.InvTypeModifiers[scaleParams.ItemType]

	log.Printf("------- scaledValue: %v modifier: %v", scaledValue, modifier)
	return int(math.Ceil(scaledValue))
}

func scaleStatv3(scaleParams StatScaleParams, difficulty int) int {
	// Calculate the quality and inventory type modifiers
	qualityModifier := config.QualityModifiers[scaleParams.Quality]
	// invTypeModifier := config.InvTypeModifiers[scaleParams.ItemType]

	// Calculate the base scaling factor
	baseScalingFactor := config.ScalingFactor[scaleParams.StatTypeId]

	// Calculate the level ratio (new item level / original item level)
	levelRatio := float64(scaleParams.NewItemLevel) * 1.0795 / float64(scaleParams.ItemLevel)

	// Apply the comprehensive scaling formula
	scaledValue := float64(scaleParams.StatValue) *
		math.Pow(levelRatio, baseScalingFactor)

	if difficulty == 3 {
		scaledValue = scaledValue * qualityModifier
	} else {
		// Apply the legendary modifier only
		if scaleParams.Quality == 5 {
			scaledValue = scaledValue * 1.25
		}
	}

	// // Log the details for debugging
	// log.Printf("------- scaledValue: %v, levelRatio: %v, qualityModifier: %v, baseScalingFactor: %v",
	// 	scaledValue, levelRatio, qualityModifier, baseScalingFactor)

	// Return the scaled value, rounded up
	return int(math.Ceil(scaledValue))
}

// This will copy higher value of spell power and attack powers into one unit.  This is to fix items that have both
func correctSpellAttackPower(item *Item, allStats map[int]*ItemStat) {
	// do some manual corrections for stats oddly getting attack power and spell power
	itemStats, err := item.GetStatList()
	if err != nil {
		log.Printf("Failed to get stat list: %v not attempting to fix stats", err)
	}
	if slices.Contains(itemStats, STAT.AttackPower) && slices.Contains(itemStats, STAT.SpellPower) {

		// if the Attack power is greater than spell power then add it spell power and remove spell power
		if allStats[STAT.AttackPower] != nil && allStats[STAT.SpellPower] != nil {
			if allStats[STAT.AttackPower].Value > allStats[STAT.SpellPower].Value {
				allStats[STAT.AttackPower].Value += allStats[STAT.SpellPower].Value
				delete(allStats, STAT.SpellPower)
			} else {
				allStats[STAT.SpellPower].Value += allStats[STAT.AttackPower].Value
				delete(allStats, STAT.AttackPower)
			}
		}
	}
	if slices.Contains(itemStats, STAT.RangedAttackPower) && slices.Contains(itemStats, STAT.SpellPower) {

		if allStats[STAT.RangedAttackPower] != nil && allStats[STAT.SpellPower] != nil {
			if allStats[STAT.RangedAttackPower].Value > allStats[STAT.SpellPower].Value {
				allStats[STAT.RangedAttackPower].Value += allStats[STAT.SpellPower].Value
				delete(allStats, STAT.SpellPower)
			} else {
				allStats[STAT.SpellPower].Value += allStats[STAT.RangedAttackPower].Value
				delete(allStats, STAT.RangedAttackPower)
			}
		}
	}
}

/**
 * This will determine the class type that would be the user of the item
 * Melee Strength Attacker: 1
 * Melee Agility Attacker: 2
 * Ranged Attacker: 3
 * Mage: 4
 * Healer: 5
 * Tank: 6
 * Generic: 7 (Could not determine)
 * @return int
 **/
func (item *Item) GetClassUserType() int {
	// loop over the stats and check if any of them are parry, defense, block
	for i := 1; i <= 7; i++ {
		statTypeField := fmt.Sprintf("StatType%d", i)
		statTypePtr, _ := item.GetField(statTypeField)

		// Tanking weapons will have defensive stats on them
		if statTypePtr == STAT.ParryRating || statTypePtr == STAT.DefenseSkillRating || statTypePtr == STAT.BlockRating || statTypePtr == STAT.BlockValue {
			return 6
		}

		// Check for a healer stats like MP5 and Spell Healing Done
		if statTypePtr == STAT.ManaRegeneration || statTypePtr == STAT.SpellHealingDone {
			return 5
		}

		// Check for a Mage stat if they have spell penetration we know it is a mage
		if statTypePtr == STAT.SpellPenetration {
			return 4
		}

		if statTypePtr == STAT.RangedAttackPower || statTypePtr == STAT.CritRangedRating || statTypePtr == STAT.HitRangedRating {
			return 3
		}
	}

	// For armor we can use the type to determine the class type
	if *item.Class == 4 {
		// if the item is cloth its a mage and did not have healer stats just treat as a mage item
		if *item.Subclass == 1 && *item.InventoryType != 16 {
			return 4
		}

		// If it is plate and not a tank then it is a strength melee attack
		if *item.Subclass == 4 {
			return 1
		}

		// If it is mail/leather armor then it is limited to Mage, Agility Fighter
		if *item.Subclass == 2 || *item.Subclass == 3 {
			// check for spellpower, spellcrit, spellhit, intellect
			for i := 1; i <= 7; i++ {
				statTypeField := fmt.Sprintf("StatType%d", i)
				statTypePtr, _ := item.GetField(statTypeField)
				if statTypePtr == STAT.SpellPower || statTypePtr == STAT.CritSpellRating ||
					statTypePtr == STAT.HitSpellRating || statTypePtr == STAT.Intellect || statTypePtr == STAT.Spirit {
					return 4
				}
			}

			return 2
		}
	}

	// Do some weapon checks
	if *item.Class == 2 {
		// If it is a fist weapon or ranged throwing weapons its agility class type
		if *item.Subclass == 13 || *item.Subclass == 16 {
			return 2
		}

		if *item.Subclass == 19 {
			return 4
		}

		// if it is a polearm or spear 17 or 6 and strength then its strength class type
		if *item.Subclass == 17 || *item.Subclass == 6 {
			for i := 1; i <= 7; i++ {
				statTypeField := fmt.Sprintf("StatType%d", i)
				statTypePtr, _ := item.GetField(statTypeField)
				if statTypePtr == STAT.Strength {
					return 1
				}

				// or attack power
				if statTypePtr == STAT.AttackPower {
					return 1
				}
			}

			// otherwise check for agility
			for i := 1; i <= 7; i++ {
				statTypeField := fmt.Sprintf("StatType%d", i)
				statTypePtr, _ := item.GetField(statTypeField)
				if statTypePtr == STAT.Agility {
					return 2
				}
			}

			// last assume it is a healer
			return 5
		}

		if *item.Subclass == 2 || *item.Subclass == 3 || *item.Subclass == 18 {
			for i := 1; i <= 7; i++ {
				statTypeField := fmt.Sprintf("StatType%d", i)
				statTypePtr, _ := item.GetField(statTypeField)
				if statTypePtr == STAT.Strength {
					return 1
				}
			}

			return 3
		}
	}

	// Most specific cases have been addressed now just use the base stats to make a decision for the remaining
	for i := 1; i <= 7; i++ {
		statTypeField := fmt.Sprintf("StatType%d", i)
		statTypePtr, _ := item.GetField(statTypeField)
		// fmt.Printf("itemName: %s StatType%d: %v \n", item.Name, i, statTypePtr)
		if statTypePtr == STAT.Spirit {
			return 5
		}
	}

	for i := 1; i <= 7; i++ {
		statTypeField := fmt.Sprintf("StatType%d", i)
		statTypePtr, _ := item.GetField(statTypeField)
		if statTypePtr == STAT.Intellect {
			return 4
		}
	}

	for i := 1; i <= 7; i++ {
		statTypeField := fmt.Sprintf("StatType%d", i)
		statTypePtr, _ := item.GetField(statTypeField)
		if statTypePtr == STAT.Strength {
			return 1
		}
	}

	for i := 1; i <= 7; i++ {
		statTypeField := fmt.Sprintf("StatType%d", i)
		statTypePtr, _ := item.GetField(statTypeField)
		if statTypePtr == STAT.Agility {
			return 2
		}
	}

	// If it is attack power melee haste melee crit or anything else then it is a agility
	for i := 1; i <= 7; i++ {
		statTypeField := fmt.Sprintf("StatType%d", i)
		statTypePtr, _ := item.GetField(statTypeField)
		if statTypePtr == STAT.AttackPower || statTypePtr == STAT.HasteMeleeRating || statTypePtr == STAT.CritMeleeRating {
			return 2
		}
	}

	// If it is spell power spell crit spell hit then it is a mage
	for i := 1; i <= 7; i++ {
		statTypeField := fmt.Sprintf("StatType%d", i)
		statTypePtr, _ := item.GetField(statTypeField)
		if statTypePtr == STAT.SpellPower || statTypePtr == STAT.CritSpellRating || statTypePtr == STAT.HitSpellRating {
			return 4
		}
	}

	// If it is ranged attack power ranged haste ranged crit or anything else then it is a ranged
	for i := 1; i <= 7; i++ {
		statTypeField := fmt.Sprintf("StatType%d", i)
		statTypePtr, _ := item.GetField(statTypeField)
		if statTypePtr == STAT.RangedAttackPower || statTypePtr == STAT.HasteRangedRating || statTypePtr == STAT.CritRangedRating {
			return 3
		}
	}

	// if we have made it here then the only thing left to do is base it purely on armor material type
	if *item.Class == 4 && *item.Subclass == 1 {
		return 4
	}

	if *item.Class == 4 && *item.Subclass == 4 {
		return 1
	}

	if *item.Class == 4 && (*item.Subclass == 2 || *item.Subclass == 3) {
		return 2
	}

	return 7
}

func (item *Item) ApplyTierModifiers(optionalTier ...int) {
	// Use provided tier or default to 0 if not set
	var tier int
	if len(optionalTier) > 0 {
		tier = optionalTier[0]
	} else {
		tier = 0
	}

	// Default tier modifier is 1.0 (no modification)
	tierModifier := 1.0

	// This is a necessary bonus to catch gear up from previous v2 version
	catchUpBonus := 1.5

	// If tier is valid (1-5), get the modifier from config
	if tier > 0 && tier <= 5 {
		if mod, ok := config.GearTierModifiers[tier]; ok {
			tierModifier = mod
		}
	}

	// Apply tier modifier to all stats on the item
	for i := 1; i <= 10; i++ {
		// Get the stat type and value fields using reflection
		statTypeField := fmt.Sprintf("StatType%d", i)
		statValueField := fmt.Sprintf("StatValue%d", i)

		// Get the current values
		statTypePtr, err1 := item.GetField(statTypeField)
		statValuePtr, err2 := item.GetField(statValueField)

		// Skip if any errors or if stat type is 0 or stat value is 0
		if err1 != nil || err2 != nil || statTypePtr == 0 || statValuePtr == 0 {
			continue
		}

		// Get the stat modifier (inverse of the cost modifier)
		statModifier, ok := config.StatModifiers[statTypePtr]
		if !ok {
			statModifier = 1.0
		}

		// Inverse of the stat modifier (e.g., 0.5 cost means 2.0 multiplier)
		inverseModifier := 1.0
		if statModifier > 0 {
			inverseModifier = 1.0 / statModifier
		}

		// Apply tier modifier and stat modifier
		newValue := int(float64(statValuePtr) * tierModifier * inverseModifier * catchUpBonus)

		// Update the item's stat value
		item.UpdateField(statValueField, newValue)

		// We've already updated the field directly with UpdateField above
		// No need to update StatsMap as we're focusing on the direct stat values
	}

	// Apply tier modifier to spells
	// spells, err := item.GetSpells()
	// if err == nil && len(spells) > 0 {
	// 	for i := range spells {
	// 		// Get the item level
	// 		currentLevel := 0
	// 		if item.ItemLevel != nil {
	// 			currentLevel = *item.ItemLevel
	// 		}

	// 		// Get the item quality
	// 		quality := 2 // Default to uncommon
	// 		if item.Quality != nil {
	// 			quality = *item.Quality
	// 		}

	// 		// Scale spells with the tier modifier
	// 		spells[i].ForceScaleSpell(currentLevel, currentLevel, quality, tier)
	// 	}

	// 	// Update the item's spells
	// 	item.Spells = spells
	// }
}

func ItemToSql(item Item, reqLevel int, difficulty int) string {

	fmt.Printf("-- Required level: %v\n", reqLevel)

	var name string = item.Name

	entryBump := 20000000
	spellBump := 30000000

	if *item.Quality == 4 {
		spellBump = 31000000
	}
	if *item.Quality == 5 {
		spellBump = 32000000
	}

	if difficulty == 4 {
		entryBump = 21000000
	}
	if difficulty == 5 {
		entryBump = 22000000
	}

	name = getRandomWord(difficulty) + " " + name

	spellList := ""
	if len(item.Spells) > 0 {
		for i, spell := range item.Spells {

			spellList += spells.SpellToSql(spell, *item.Quality)
			item.UpdateField(fmt.Sprintf("SpellId%v", i), spellBump+spell.ID)
		}
	}

	delete := fmt.Sprintf("DELETE FROM acore_world.item_template WHERE entry = %v;", entryBump+item.Entry)

	clone := fmt.Sprintf(`
	INSERT INTO acore_world.item_template  (
		entry, class, subclass, SoundOverrideSubclass, name, displayid, Quality, Flags, FlagsExtra, BuyCount, 
		BuyPrice, SellPrice, InventoryType, AllowableClass, AllowableRace, ItemLevel, RequiredLevel, 
		RequiredSkill, RequiredSkillRank, requiredspell, requiredhonorrank, RequiredCityRank, 
		RequiredReputationFaction, RequiredReputationRank, maxcount, stackable, ContainerSlots, StatsCount, 
		stat_type1, stat_value1, stat_type2, stat_value2, stat_type3, stat_value3, stat_type4, stat_value4, 
		stat_type5, stat_value5, stat_type6, stat_value6, stat_type7, stat_value7, stat_type8, stat_value8, 
		stat_type9, stat_value9, stat_type10, stat_value10, ScalingStatDistribution, ScalingStatValue, 
		dmg_min1, dmg_max1, dmg_type1, dmg_min2, dmg_max2, dmg_type2, armor, holy_res, fire_res, nature_res, 
		frost_res, shadow_res, arcane_res, delay, ammo_type, RangedModRange, spellid_1, spelltrigger_1, 
		spellcharges_1, spellppmRate_1, spellcooldown_1, spellcategory_1, spellcategorycooldown_1, spellid_2, 
		spelltrigger_2, spellcharges_2, spellppmRate_2, spellcooldown_2, spellcategory_2, spellcategorycooldown_2, 
		spellid_3, spelltrigger_3, spellcharges_3, spellppmRate_3, spellcooldown_3, spellcategory_3, 
		spellcategorycooldown_3, spellid_4, spelltrigger_4, spellcharges_4, spellppmRate_4, spellcooldown_4, 
		spellcategory_4, spellcategorycooldown_4, spellid_5, spelltrigger_5, spellcharges_5, spellppmRate_5, 
		spellcooldown_5, spellcategory_5, spellcategorycooldown_5, bonding, description, PageText, LanguageID, 
		PageMaterial, startquest, lockid, Material, sheath, RandomProperty, RandomSuffix, block, itemset, 
		MaxDurability, area, Map, BagFamily, TotemCategory, socketColor_1, socketContent_1, socketColor_2, 
		socketContent_2, socketColor_3, socketContent_3, socketBonus, GemProperties, RequiredDisenchantSkill, 
		ArmorDamageModifier, duration, ItemLimitCategory, HolidayId, ScriptName, DisenchantID, FoodType, 
		minMoneyLoot, maxMoneyLoot, flagsCustom, VerifiedBuild
	  )
	  SELECT 
		entry + %v, class, subclass, SoundOverrideSubclass, name, displayid, Quality, Flags, FlagsExtra, BuyCount, 
		BuyPrice, SellPrice, InventoryType, AllowableClass, AllowableRace, ItemLevel, RequiredLevel, 
		RequiredSkill, RequiredSkillRank, requiredspell, requiredhonorrank, RequiredCityRank, 
		RequiredReputationFaction, RequiredReputationRank, maxcount, stackable, ContainerSlots, StatsCount, 
		stat_type1, stat_value1, stat_type2, stat_value2, stat_type3, stat_value3, stat_type4, stat_value4, 
		stat_type5, stat_value5, stat_type6, stat_value6, stat_type7, stat_value7, stat_type8, stat_value8, 
		stat_type9, stat_value9, stat_type10, stat_value10, ScalingStatDistribution, ScalingStatValue, 
		dmg_min1, dmg_max1, dmg_type1, dmg_min2, dmg_max2, dmg_type2, armor, holy_res, fire_res, nature_res, 
		frost_res, shadow_res, arcane_res, delay, ammo_type, RangedModRange, spellid_1, spelltrigger_1, 
		spellcharges_1, spellppmRate_1, spellcooldown_1, spellcategory_1, spellcategorycooldown_1, spellid_2, 
		spelltrigger_2, spellcharges_2, spellppmRate_2, spellcooldown_2, spellcategory_2, spellcategorycooldown_2, 
		spellid_3, spelltrigger_3, spellcharges_3, spellppmRate_3, spellcooldown_3, spellcategory_3, 
		spellcategorycooldown_3, spellid_4, spelltrigger_4, spellcharges_4, spellppmRate_4, spellcooldown_4, 
		spellcategory_4, spellcategorycooldown_4, spellid_5, spelltrigger_5, spellcharges_5, spellppmRate_5, 
		spellcooldown_5, spellcategory_5, spellcategorycooldown_5, bonding, description, PageText, LanguageID, 
		PageMaterial, startquest, lockid, Material, sheath, RandomProperty, RandomSuffix, block, itemset, 
		MaxDurability, area, Map, BagFamily, TotemCategory, socketColor_1, socketContent_1, socketColor_2, 
		socketContent_2, socketColor_3, socketContent_3, socketBonus, GemProperties, RequiredDisenchantSkill, 
		ArmorDamageModifier, duration, ItemLimitCategory, HolidayId, ScriptName, DisenchantID, FoodType, 
		minMoneyLoot, maxMoneyLoot, flagsCustom, VerifiedBuild
	  FROM acore_world.item_template as src
	  WHERE src.entry = %v ON DUPLICATE KEY UPDATE entry = src.entry + %v;	  
	`, entryBump, item.Entry, entryBump)

	update := fmt.Sprintf(`
	UPDATE acore_world.item_template
	SET 
	  Quality = %v,
	  name = '%s',
	  ItemLevel = %v,
	  RequiredLevel = %v,
	  dmg_min1 = %v,
	  dmg_max1 = %v,
	  dmg_min2 = %v,
	  dmg_max2 = %v,
	  StatsCount = %v,
	  stat_type1 = %v,
	  stat_value1 = %v,
	  stat_type2 = %v,
	  stat_value2 = %v,
	  stat_type3 = %v,
	  stat_value3 = %v,
	  stat_type4 = %v,
	  stat_value4 = %v,
	  stat_type5 = %v,
	  stat_value5 = %v,
	  stat_type6 = %v,
	  stat_value6 = %v,
	  stat_type7 = %v,
	  stat_value7 = %v,
	  stat_type8 = %v,
	  stat_value8 = %v,
	  stat_type9 = %v,
	  stat_value9 = %v,
	  stat_type10 = %v,
	  stat_value10 = %v,
	  spellid_1 = %v,
	  spellid_2 = %v,
	  spellid_3 = %v,
	  spelltrigger_1 = %v,
	  spelltrigger_2 = %v,
	  spelltrigger_3 = %v,
	  socketColor_1 = %v,
	  socketContent_1 = %v,
	  socketColor_2 = %v,
	  socketContent_2 = %v,
	  socketColor_3 = %v,
	  socketContent_3 = %v,
	  socketBonus = %v,
	  GemProperties = %v,
	  RequiredDisenchantSkill = %v,
	  DisenchantID = %v,
	  SellPrice = FLOOR(100000 + (RAND() * 400001)),
	  Armor = %v
	WHERE entry = %v;
	`, *item.Quality, strings.ReplaceAll(name, "'", "''"), *item.ItemLevel, reqLevel, *item.MinDmg1, *item.MaxDmg1, *item.MinDmg2, *item.MaxDmg2, *item.StatsCount,
		*item.StatType1, *item.StatValue1, *item.StatType2, *item.StatValue2, *item.StatType3, *item.StatValue3, *item.StatType4, *item.StatValue4,
		*item.StatType5, *item.StatValue5, *item.StatType6, *item.StatValue6, *item.StatType7, *item.StatValue7, *item.StatType8, *item.StatValue8,
		*item.StatType9, *item.StatValue9, *item.StatType10, *item.StatValue10, *item.SpellId1, *item.SpellId2, *item.SpellId3, *item.SpellTrigger1, *item.SpellTrigger2,
		*item.SpellTrigger3, *item.SocketColor1, *item.SocketContent1, *item.SocketColor2, *item.SocketContent2,
		*item.SocketColor3, *item.SocketContent3, *item.SocketBonus, *item.GemProperties,
		375, 68, *item.Armor, entryBump+item.Entry)

	return fmt.Sprintf("%s %s \n %s \n %s", spellList, delete, clone, update)
}

func getRandomWord(difficulty int) string {
	mythic := []string{"Mythic", "Powerful", "Stalwart", "Venerated", "Mighty", "Unyielding"}
	legendary := []string{"Legendary", "Fabled", "Exalted", "Magnificent", "Pristine", "Supreme", "Glorious"}
	ascendant := []string{"Ascendant", "Godlike", "Celestial", "Transcendant", "Divine", "Omnipotent", "Demonforged", "Immortal", "Omniscient", "Ethereal"}

	r := rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), 2))

	switch difficulty {
	case 3: // Mythic
		randomIndex := r.IntN(len(mythic))
		return mythic[randomIndex]
	case 4: // Legendary
		randomIndex := r.IntN(len(legendary))
		return legendary[randomIndex]
	case 5: // Ascendant
		randomIndex := r.IntN(len(ascendant))
		return ascendant[randomIndex]
	default:
		return ""
	}
}
